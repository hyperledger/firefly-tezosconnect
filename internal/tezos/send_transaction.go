package tezos

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/syndtr/goleveldb/leveldb"
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

	// TODO: get last block from block listener cache
	hash, _ := c.client.GetBlockHash(ctx, rpc.Head)
	op.WithBranch(hash)

	receipt, err := c.client.Send(ctx, op, nil)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	// TODO: Now Tezos client also acts as a comfirmation manager and listen the blockchain to get tx receipt.
	// FF tx manager should deal with it instead. This solution is temporary, for MVP purpose only.
	db, err := leveldb.OpenFile("/tmp/txs", nil)
	if err != nil {
		return nil, "", err
	}
	defer db.Close()

	receiptData, _ := json.Marshal(receipt)

	err = db.Put([]byte(receipt.Op.Hash.String()), receiptData, nil)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	return &ffcapi.TransactionSendResponse{
		TransactionHash: receipt.Op.Hash.String(),
	}, "", nil
}
