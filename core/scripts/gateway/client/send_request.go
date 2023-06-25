package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/api"
)

func hexToPrivateKey(hexStr string) (*ecdsa.PrivateKey, error) {
	rawPrivKey, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	var privateKey ecdsa.PrivateKey
	curve := elliptic.P256()
	privateKey.PublicKey.Curve = curve
	privateKey.D = big.NewInt(0).SetBytes(rawPrivKey)
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(rawPrivKey)
	return &privateKey, nil
}

func main() {
	gatewayURL := flag.String("gateway_url", "", "Gateway URL")
	messageId := flag.String("message_id", "", "Request ID")
	methodName := flag.String("method", "", "Method name")
	donId := flag.String("don_id", "", "DON ID")
	payloadJSON := flag.String("payload_json", "", "Payload JSON")
	privateKey := flag.String("private_key", "", "Private key	to sign the message with")
	flag.Parse()

	msg := &api.Message{
		Body: api.MessageBody{
			MessageId: *messageId,
			Method:    *methodName,
			DonId:     *donId,
			Payload:   json.RawMessage(*payloadJSON),
		},
	}
	key, err := hexToPrivateKey(*privateKey)
	if err != nil {
		fmt.Println("error parsing private key", err)
		return
	}
	err = msg.Sign(key)
	if err != nil {
		fmt.Println("error signing", err)
		return
	}
	codec := api.JsonRPCCodec{}
	rawMsg, err := codec.EncodeRequest(msg)
	if err != nil {
		fmt.Println("error JSON-RPC encoding", err)
		return
	}
	req, err := http.NewRequestWithContext(context.Background(), "POST", *gatewayURL, bytes.NewBuffer(rawMsg))
	if err != nil {
		fmt.Println("error creating an HTTP request", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error sending a request", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error sending a request", err)
		return
	}

	fmt.Println("Status Code: ", resp.StatusCode)
	fmt.Println("Body: ", string(body))
}
