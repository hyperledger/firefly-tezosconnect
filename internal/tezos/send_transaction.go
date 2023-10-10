package tezos

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-common/pkg/log"
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

	opts := &rpc.DefaultOptions

	// simulate to check tx validity and estimate cost
	sim, err := c.client.Simulate(ctx, op, opts)
	if err != nil {
		return nil, mapError(sendRPCMethods, err), err
	}
	// fail with Tezos error when simulation failed
	if !sim.IsSuccess() {
		return nil, "", sim.Error()
	}

	// apply simulated cost as limits to tx list
	if !opts.IgnoreLimits {
		op.WithLimits(sim.MinLimits(), rpc.ExtraSafetyMargin)
	}

	// log info about tx costs
	costs := sim.Costs()
	for i, v := range op.Contents {
		verb := "used"
		if opts.IgnoreLimits {
			verb = "forced"
		}
		limits := v.Limits()
		log.L(ctx).Debugf("OP#%03d: %s gas_used(sim)=%d storage_used(sim)=%d storage_burn(sim)=%d alloc_burn(sim)=%d fee(%s)=%d gas_limit(%s)=%d storage_limit(%s)=%d ",
			i, v.Kind(), costs[i].GasUsed, costs[i].StorageUsed, costs[i].StorageBurn, costs[i].AllocationBurn,
			verb, limits.Fee, verb, limits.GasLimit, verb, limits.StorageLimit,
		)
	}

	// check minFee calc against maxFee if set
	if opts.MaxFee > 0 {
		if l := op.Limits(); l.Fee > opts.MaxFee {
			return nil, "", fmt.Errorf("estimated cost %d > max %d", l.Fee, opts.MaxFee)
		}
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

// nolint:unparam
func (c *tezosConnector) signTxRemotely(ctx context.Context, op *codec.Op) error {
	url := c.signatoryURL + "/keys/" + op.Source.String()
	requestBody, _ := json.Marshal(hex.EncodeToString(op.WatermarkedBytes()))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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
