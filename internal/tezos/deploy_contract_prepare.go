package tezos

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

func (c *tezosConnector) DeployContractPrepare(ctx context.Context, req *ffcapi.ContractDeployPrepareRequest) (*ffcapi.TransactionPrepareResponse, ffcapi.ErrorReason, error) {
	if req.Contract == nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, i18n.NewError(ctx, "Missing contract", req.Contract)
	}

	var sc micheline.Script
	_ = json.Unmarshal([]byte(req.Contract.String()), &sc)
	orig := &codec.Origination{
		Script: sc,
	}

	addr, err := tezos.ParseAddress(req.From)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, i18n.NewError(ctx, msgs.MsgInvalidFromAddress, req.From, err)
	}

	headBlockHash, _ := c.client.GetBlockHash(ctx, rpc.Head)
	op := codec.NewOp().
		WithContents(orig).
		WithSource(addr).
		WithBranch(headBlockHash)

	err = c.completeOp(ctx, op, req.From, req.Nonce)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}
	opts := &rpc.DefaultOptions
	if reason, err := c.estimateAndAssignTxCost(ctx, op, opts); err != nil {
		return nil, reason, err
	}

	log.L(ctx).Infof("Prepared deploy transaction dataLen=%d", len(op.Bytes()))

	return &ffcapi.TransactionPrepareResponse{
		Gas:             req.Gas,
		TransactionData: hex.EncodeToString(op.Bytes()),
	}, "", nil
}
