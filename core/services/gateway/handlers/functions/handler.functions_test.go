package functions_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/api"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/config"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/handlers"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/handlers/functions"
	functions_mocks "github.com/smartcontractkit/chainlink/v2/core/services/gateway/handlers/functions/mocks"
	handlers_mocks "github.com/smartcontractkit/chainlink/v2/core/services/gateway/handlers/mocks"
)

var nodes = []string{"0x11", "0x22", "0x33", "0x44"}

func newFunctionsHandlerForATestDON(t *testing.T) (handlers.Handler, *handlers_mocks.DON, *functions_mocks.OnchainAllowlist) {
	cfg := &functions.FunctionsHandlerConfig{}
	donConfig := &config.DONConfig{
		Members: []config.NodeConfig{
			{Name: "nodeA", Address: nodes[0]},
			{Name: "nodeB", Address: nodes[1]},
			{Name: "nodeC", Address: nodes[2]},
			{Name: "nodeD", Address: nodes[3]},
		},
		F: 1,
	}
	don := handlers_mocks.NewDON(t)
	allowlist := functions_mocks.NewOnchainAllowlist(t)
	rateLimiter := handlers.NewRateLimiter(handlers.RateLimiterConfig{GlobalRPS: 100.0, GlobalBurst: 100, PerUserRPS: 100.0, PerUserBurst: 100})

	handler := functions.NewFunctionsHandler(cfg, donConfig, don, allowlist, rateLimiter, logger.TestLogger(t))
	return handler, don, allowlist
}

func TestFunctionsHandler_EmptyConfigNilMessage(t *testing.T) {
	t.Parallel()

	handler, err := functions.NewFunctionsHandlerFromConfig(json.RawMessage("{}"), &config.DONConfig{}, nil, nil, logger.TestLogger(t))
	require.NoError(t, err)

	// nil message
	err = handler.HandleUserMessage(context.Background(), nil, nil)
	require.Error(t, err)
}

func TestFunctionsHandler_HandleUserMessage_SuccessfulSet(t *testing.T) {
	t.Parallel()

	handler, don, allowlist := newFunctionsHandlerForATestDON(t)

	userRequestMsg := api.Message{
		Body: api.MessageBody{
			MessageId: "1234",
			Method:    "secrets_set",
			DonId:     "don_id",
		},
	}
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	err = userRequestMsg.Sign(privateKey)
	userRequestMsg.Body.Sender = strings.ToLower(address.Hex())
	require.NoError(t, err)

	nodeResponseMsg := userRequestMsg
	nodeResponseMsg.Body.Payload = []byte("signed ACK")
	nodeResponseMsg.Signature = ""

	callbachCh := make(chan handlers.UserCallbackPayload)
	done := make(chan struct{})
	go func() {
		// wait on a response from Gateway to the user
		response := <-callbachCh
		require.Equal(t, api.NoError, response.ErrCode)
		require.Equal(t, "1234", response.Msg.Body.MessageId)
		close(done)
	}()

	allowlist.On("Allow", address).Return(true, nil)
	// Two nodes will succeed and two will fail
	don.On("SendToNode", mock.Anything, nodes[0], mock.Anything).Run(func(args mock.Arguments) {
		err := handler.HandleNodeMessage(context.Background(), &nodeResponseMsg, nodes[0])
		require.NoError(t, err)
	}).Return(nil)
	don.On("SendToNode", mock.Anything, nodes[1], mock.Anything).Run(func(args mock.Arguments) {
		err := handler.HandleNodeMessage(context.Background(), &nodeResponseMsg, nodes[1])
		require.NoError(t, err)
	}).Return(nil)
	don.On("SendToNode", mock.Anything, nodes[2], mock.Anything).Return(errors.New("failed"))
	don.On("SendToNode", mock.Anything, nodes[3], mock.Anything).Return(errors.New("failed"))

	err = handler.HandleUserMessage(context.Background(), &userRequestMsg, callbachCh)
	require.NoError(t, err)
	<-done
}
