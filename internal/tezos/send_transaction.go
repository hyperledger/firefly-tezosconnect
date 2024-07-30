package tezos

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/trilitech/tzgo/codec"
	"github.com/trilitech/tzgo/tezos"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// TransactionSend combines a previously prepared encoded transaction, with a current gas price, and submits it to the transaction pool of the blockchain for mining
func (c *tezosConnector) TransactionSend(ctx context.Context, req *ffcapi.TransactionSendRequest) (*ffcapi.TransactionSendResponse, ffcapi.ErrorReason, error) {
	opBytes, err := hex.DecodeString(req.TransactionData)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	op, err := codec.DecodeOp(opBytes)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	// auto-complete op with branch, source, nonce, chain params
	err = c.completeOp(ctx, op, req.From, req.Nonce)
	if err != nil {
		return nil, "", err
	}

	// sign tx
	err = c.signTxRemotely(ctx, op)
	if err != nil {
		return nil, "", err
	}

	// broadcast
	hash, err := c.client.Broadcast(ctx, op)
	if err != nil {
		return nil, mapError(sendRPCMethods, err), err
	}

	return &ffcapi.TransactionSendResponse{
		TransactionHash: hash.String(),
	}, "", nil
}

func (c *tezosConnector) signTxRemotely(ctx context.Context, op *codec.Op) error {
	if op == nil {
		return errors.New("operation is empty")
	}
	url := c.signatoryURL + "/keys/" + op.Source.String()
	requestBody, _ := json.Marshal(hex.EncodeToString(op.WatermarkedBytes()))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("signatory resp with wrong status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var signatureJSON struct {
		Signature string
	}
	err = json.Unmarshal(body, &signatureJSON)
	if err != nil {
		return err
	}

	var sig tezos.Signature
	err = sig.UnmarshalText([]byte(signatureJSON.Signature))
	if err != nil {
		return err
	}

	op.WithSignature(sig)
	return nil
}
