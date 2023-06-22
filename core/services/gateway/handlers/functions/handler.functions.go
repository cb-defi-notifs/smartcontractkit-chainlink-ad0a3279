package functions

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/smartcontractkit/chainlink/v2/core/chains/evm"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/api"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/config"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/handlers"
	"github.com/smartcontractkit/chainlink/v2/core/utils"
)

type FunctionsHandlerConfig struct {
	AllowlistConfig          *FunctionsAllowlistConfig   `json:"allowlistConfig"`
	RateLimiterConfig        *handlers.RateLimiterConfig `json:"rateLimiterConfig"`
	MaxPendingRequestPerUser int                         `json:"maxPendingRequestPerUser"`
	MaxPendingRequestsGlobal int                         `json:"maxPendingRequestsGlobal"`
	RequestTimeoutMillis     int64                       `json:"requestTimeoutMillis"`
}

type FunctionsAllowlistConfig struct {
	ChainID            int64  `json:"chainID"`
	ContractAddress    string `json:"contractAddress"`
	BlockConfirmations int64  `json:"blockConfirmations"`
	UpdateFrequencySec int    `json:"updateFrequencySec"`
	UpdateTimeoutSec   int    `json:"updateTimeoutSec"`
}

type functionsHandler struct {
	utils.StartStopOnce

	handlerConfig     *FunctionsHandlerConfig
	donConfig         *config.DONConfig
	don               handlers.DON
	pendingRequests   map[requestKey]*pendingRequest
	pendingRequestsMu sync.Mutex
	allowlist         OnchainAllowlist
	rateLimiter       *handlers.RateLimiter
	chStop            utils.StopChan
	shutdownWaitGroup sync.WaitGroup
	lggr              logger.Logger
}

// TODO(FUN-645): move pending requests structure to utils, support removal and timeouts, improve concurrency
type requestKey struct {
	sender string
	id     string
}

type pendingRequest struct {
	responses  map[string]*api.Message
	callbackCh chan<- handlers.UserCallbackPayload
}

var _ handlers.Handler = (*functionsHandler)(nil)

const (
	methodSecretsSet  = "secrets_set"
	methodSecretsList = "secrets_list"
)

func NewFunctionsHandlerFromConfig(handlerConfig json.RawMessage, donConfig *config.DONConfig, don handlers.DON, chains evm.ChainSet, lggr logger.Logger) (handlers.Handler, error) {
	cfg, err := ParseConfig(handlerConfig)
	if err != nil {
		return nil, err
	}
	var allowlist OnchainAllowlist
	if cfg.AllowlistConfig != nil {
		chain, err := chains.Get(big.NewInt(cfg.AllowlistConfig.ChainID))
		if err != nil {
			return nil, err
		}
		allowlist, err = NewOnchainAllowlist(chain.Client(), common.HexToAddress(cfg.AllowlistConfig.ContractAddress), cfg.AllowlistConfig.BlockConfirmations, lggr)
		if err != nil {
			return nil, err
		}
	}
	var rateLimiter *handlers.RateLimiter
	if cfg.RateLimiterConfig != nil {
		rateLimiter = handlers.NewRateLimiter(*cfg.RateLimiterConfig)
	}
	return NewFunctionsHandler(cfg, donConfig, don, allowlist, rateLimiter, lggr), nil
}

func NewFunctionsHandler(
	cfg *FunctionsHandlerConfig,
	donConfig *config.DONConfig,
	don handlers.DON,
	allowlist OnchainAllowlist,
	rateLimiter *handlers.RateLimiter,
	lggr logger.Logger) handlers.Handler {
	return &functionsHandler{
		handlerConfig:   cfg,
		donConfig:       donConfig,
		don:             don,
		pendingRequests: make(map[requestKey]*pendingRequest),
		allowlist:       allowlist,
		rateLimiter:     rateLimiter,
		chStop:          make(utils.StopChan),
		lggr:            lggr,
	}
}

func ParseConfig(handlerConfig json.RawMessage) (*FunctionsHandlerConfig, error) {
	var cfg FunctionsHandlerConfig
	if err := json.Unmarshal(handlerConfig, &cfg); err != nil {
		return nil, err
	}
	if cfg.AllowlistConfig != nil {
		if !common.IsHexAddress(cfg.AllowlistConfig.ContractAddress) {
			return nil, errors.New("ContractAddress is not a valid hex address")
		}
		if cfg.AllowlistConfig.UpdateFrequencySec <= 0 {
			return nil, errors.New("UpdateFrequencySec must be positive")
		}
		if cfg.AllowlistConfig.UpdateTimeoutSec <= 0 {
			return nil, errors.New("UpdateTimeoutSec must be positive")
		}
	}
	if cfg.RateLimiterConfig != nil {
		if cfg.RateLimiterConfig.GlobalRPS <= 0.0 || cfg.RateLimiterConfig.PerUserRPS <= 0.0 {
			return nil, errors.New("RPS values must be positive")
		}
		if cfg.RateLimiterConfig.GlobalBurst <= 0 || cfg.RateLimiterConfig.PerUserBurst <= 0 {
			return nil, errors.New("burst values must be positive")
		}
	}
	return &cfg, nil
}

func (h *functionsHandler) HandleUserMessage(ctx context.Context, msg *api.Message, callbackCh chan<- handlers.UserCallbackPayload) error {
	if err := msg.Validate(); err != nil {
		h.lggr.Debug("received invalid message", "err", err)
		return err
	}
	sender := common.HexToAddress(msg.Body.Sender)
	if h.allowlist != nil && !h.allowlist.Allow(sender) {
		h.lggr.Debug("received a message from a non-allowlisted address", "sender", msg.Body.Sender)
		return errors.New("sender not allowlisted")
	}
	if h.rateLimiter != nil && !h.rateLimiter.Allow(msg.Body.Sender) {
		h.lggr.Debug("rate-limited", "sender", msg.Body.Sender)
		return errors.New("rate-limited")
	}
	h.lggr.Debug("received a valid message", "sender", msg.Body.Sender)
	switch msg.Body.Method {
	case methodSecretsSet, methodSecretsList:
		return h.handleSecretsRequest(ctx, msg, callbackCh)
	default:
		h.lggr.Debug("unsupported method", "method", msg.Body.Method)
		return errors.New("unsupported method")
	}
}

func (h *functionsHandler) handleSecretsRequest(ctx context.Context, msg *api.Message, callbackCh chan<- handlers.UserCallbackPayload) error {
	// Save a new pending request.
	key := requestKey{sender: msg.Body.Sender, id: msg.Body.MessageId}
	var err error
	h.pendingRequestsMu.Lock()
	_, ok := h.pendingRequests[key]
	if ok {
		err = errors.New("duplicate request")
	} else {
		h.pendingRequests[key] = &pendingRequest{callbackCh: callbackCh, responses: make(map[string]*api.Message)}
	}
	h.pendingRequestsMu.Unlock()
	if err != nil {
		return err
	}
	// Send to all nodes.
	errCount := 0
	for _, member := range h.donConfig.Members {
		err := h.don.SendToNode(ctx, member.Address, msg)
		if err != nil {
			errCount++
		}
	}
	// Require at least F+1 successful sends.
	if errCount >= len(h.donConfig.Members)-h.donConfig.F {
		return errors.New("failed to send to enough nodes, please re-try")
	}
	return nil
}

func (h *functionsHandler) HandleNodeMessage(ctx context.Context, msg *api.Message, nodeAddr string) error {
	key := requestKey{sender: msg.Body.Sender, id: msg.Body.MessageId}
	var response *handlers.UserCallbackPayload
	h.pendingRequestsMu.Lock()
	found, ok := h.pendingRequests[key]
	if !ok {
		h.lggr.Warnw("received a node message for an unknown request", "sender", msg.Body.Sender, "id", msg.Body.MessageId, "node", nodeAddr)
	} else {
		found.responses[nodeAddr] = msg
		if len(found.responses) == h.donConfig.F+1 {
			h.lggr.Debugw("received enough responses for a request", "sender", msg.Body.Sender, "id", msg.Body.MessageId)
			// TODO(FUN-645): combine messages and decide on aggregate success/failure
			response = &handlers.UserCallbackPayload{Msg: msg, ErrCode: api.NoError, ErrMsg: ""}
			delete(h.pendingRequests, key)
		}
	}
	h.pendingRequestsMu.Unlock()
	if response != nil {
		found.callbackCh <- *response
		close(found.callbackCh)
	}
	return nil
}

func (h *functionsHandler) Start(ctx context.Context) error {
	return h.StartOnce("FunctionsHandler", func() error {
		h.lggr.Info("starting FunctionsHandler")
		if h.allowlist != nil {
			checkFreq := time.Duration(h.handlerConfig.AllowlistConfig.UpdateFrequencySec) * time.Second
			checkTimeout := time.Duration(h.handlerConfig.AllowlistConfig.UpdateTimeoutSec) * time.Second
			h.shutdownWaitGroup.Add(1)
			go func() {
				serviceCtx, cancel := h.chStop.NewCtx()
				h.allowlist.UpdatePeriodically(serviceCtx, checkFreq, checkTimeout)
				cancel()
				h.shutdownWaitGroup.Done()
			}()
		}
		return nil
	})
}

func (h *functionsHandler) Close() error {
	return h.StopOnce("FunctionsHandler", func() (err error) {
		close(h.chStop)
		h.shutdownWaitGroup.Wait()
		return nil
	})
}
