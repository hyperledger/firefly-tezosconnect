package tezos

import (
	"context"
	"encoding/hex"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// TransactionPrepare validates transaction inputs against the supplied schema/Michelson and performs any binary serialization required (prior to signing) to encode a transaction from JSON into the native blockchain format
func (c *tezosConnector) TransactionPrepare(ctx context.Context, req *ffcapi.TransactionPrepareRequest) (*ffcapi.TransactionPrepareResponse, ffcapi.ErrorReason, error) {
	op := codec.NewOp()
	op.WithSource(tezos.MustParseAddress(req.From))

	methodName := req.Method.JSONObject().GetString("name")
	txArgs := contract.TxArgs{}

	// TODO: take value from req after FFI -> Michelson encoder is ready
	params := micheline.Parameters{
		Entrypoint: methodName,
		Value:      micheline.NewInt64(2),
	}
	txArgs.WithParameters(params)
	txArgs.WithSource(tezos.MustParseAddress(req.From))
	txArgs.WithDestination(tezos.MustParseAddress(req.To))
	op.WithContents(txArgs.Encode())

	// TODO: Get last block from cache
	hash, _ := c.client.GetBlockHash(ctx, rpc.Head)
	op.WithBranch(hash)

	// TODO: add gas esimation

	callData := op.Bytes()

	log.L(ctx).Infof("Prepared transaction method=%s dataLen=%d gas=%s", methodName, len(callData), req.Gas.Int())

	return &ffcapi.TransactionPrepareResponse{
		Gas:             req.Gas,
		TransactionData: hex.EncodeToString(callData),
	}, "", nil
}
