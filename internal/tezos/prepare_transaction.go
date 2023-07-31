package tezos

import (
	"context"
	"encoding/hex"
	"strings"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/OneOf-Inc/firefly-tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// TransactionPrepare validates transaction inputs against the supplied schema/Michelson and performs any binary serialization required (prior to signing) to encode a transaction from JSON into the native blockchain format
func (c *tezosConnector) TransactionPrepare(ctx context.Context, req *ffcapi.TransactionPrepareRequest) (*ffcapi.TransactionPrepareResponse, ffcapi.ErrorReason, error) {
	params, err := c.prepareInputParams(ctx, &req.TransactionInput)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	op, err := c.buildOp(ctx, params, req.From, req.To, req.Nonce)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	return &ffcapi.TransactionPrepareResponse{
		Gas:             req.Gas,
		TransactionData: hex.EncodeToString(op.Bytes()),
	}, "", nil
}

func (c *tezosConnector) prepareInputParams(ctx context.Context, req *ffcapi.TransactionInput) (micheline.Parameters, error) {
	var tezosParams micheline.Parameters

	for i, p := range req.Params {
		if p != nil {
			err := tezosParams.UnmarshalJSON([]byte(*p))
			if err != nil {
				return tezosParams, i18n.NewError(ctx, msgs.MsgUnmarshalParamFail, i, err)
			}
		}
	}

	return tezosParams, nil
}

func (c *tezosConnector) buildOp(ctx context.Context, params micheline.Parameters, fromString, toString string, nonce *fftypes.FFBigInt) (*codec.Op, error) {
	op := codec.NewOp()

	toAddress, err := tezos.ParseAddress(toString)
	if err != nil {
		return nil, i18n.NewError(ctx, msgs.MsgInvalidToAddress, toString, err)
	}

	txArgs := contract.TxArgs{}
	txArgs.WithParameters(params)
	txArgs.WithDestination(toAddress)
	op.WithContents(txArgs.Encode())

	c.completeOp(ctx, op, fromString, nonce)

	return op, nil
}

func (c *tezosConnector) completeOp(ctx context.Context, op *codec.Op, fromString string, nonce *fftypes.FFBigInt) error {
	fromAddress, err := tezos.ParseAddress(fromString)
	if err != nil {
		return i18n.NewError(ctx, msgs.MsgInvalidFromAddress, fromString, err)
	}
	op.WithSource(fromAddress)

	hash, _ := c.client.GetBlockHash(ctx, rpc.Head)
	op.WithBranch(hash)

	op.WithParams(getNetworkParamsByName(c.networkName))

	// assign nonce
	nextCounter := nonce.Int64()
	for _, op := range op.Contents {
		// skip non-manager ops
		if op.GetCounter() < 0 {
			continue
		}
		op.WithCounter(nextCounter)
		nextCounter++
	}

	return nil
}

func getNetworkParamsByName(name string) *tezos.Params {
	switch strings.ToLower(name) {
	case "ghostnet":
		return tezos.GhostnetParams
	case "nairobinet":
		return tezos.NairobinetParams
	default:
		return tezos.DefaultParams
	}
}
