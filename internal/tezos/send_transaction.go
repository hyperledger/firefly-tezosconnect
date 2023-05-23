package tezos

import (
	"context"
	"encoding/hex"
	"fmt"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
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

	// TODO: remove when switch to mainnet
	op.WithParams(tezos.GhostnetParams)

	// TODO: remove after it returns correct branch from 'PrepareTransaction'
	fmt.Println("Branch 1: ", op.Branch.String())
	hash, _ := c.client.GetBlockHash(ctx, rpc.Head)
	op.WithBranch(hash)
	fmt.Println("Branch 2: ", op.Branch.String())

	receipt, err := c.client.Send(ctx, op, nil)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	fmt.Println(receipt)

	return &ffcapi.TransactionSendResponse{
		TransactionHash: receipt.Op.Hash.String(),
	}, "", nil
}
