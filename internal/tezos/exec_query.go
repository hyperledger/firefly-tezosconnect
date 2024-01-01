package tezos

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// QueryInvoke executes a method on a blockchain smart contract, which might execute Smart Contract code, but does not affect the blockchain state.
func (c *tezosConnector) QueryInvoke(ctx context.Context, req *ffcapi.QueryInvokeRequest) (*ffcapi.QueryInvokeResponse, ffcapi.ErrorReason, error) {
	if req == nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, errors.New("request is not defined")
	}

	params, err := c.prepareInputParams(ctx, &req.TransactionInput)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	resp, err := c.runView(ctx, params.Entrypoint, req.From, req.To, params.Value)
	if err != nil {
		return nil, ffcapi.ErrorReasonTransactionReverted, err
	}

	outputs, _ := json.Marshal(resp)
	if val, ok := resp.(string); ok {
		if values := strings.Split(val, ","); len(values) > 1 {
			outputs, _ = json.Marshal(values)
		}
	}
	return &ffcapi.QueryInvokeResponse{
		Outputs: fftypes.JSONAnyPtrBytes(outputs),
	}, "", nil
}

func (c *tezosConnector) runView(ctx context.Context, entrypoint, addrFrom, addrTo string, args micheline.Prim) (interface{}, error) {
	toAddress, err := tezos.ParseAddress(addrTo)
	if err != nil {
		return nil, i18n.NewError(ctx, msgs.MsgInvalidToAddress, addrTo, err)
	}

	fromAddress, err := tezos.ParseAddress(addrFrom)
	if err != nil {
		return nil, i18n.NewError(ctx, msgs.MsgInvalidFromAddress, addrFrom, err)
	}

	req := rpc.RunViewRequest{
		Contract:     toAddress,
		View:         entrypoint,
		Input:        args,
		Source:       fromAddress,
		Payer:        fromAddress,
		UnlimitedGas: true,
		Mode:         "Readable",
	}

	var res rpc.RunViewResponse
	err = c.client.RunView(ctx, rpc.Head, &req, &res)
	if err != nil {
		return nil, err
	}

	return res.Data.Value(res.Data.OpCode), nil
}
