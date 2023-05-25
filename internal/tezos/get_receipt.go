package tezos

import (
	"context"
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/rpc"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/syndtr/goleveldb/leveldb"
)

// TransactionReceipt queries to see if a receipt is available for a given transaction hash
func (c *tezosConnector) TransactionReceipt(ctx context.Context, req *ffcapi.TransactionReceiptRequest) (*ffcapi.TransactionReceiptResponse, ffcapi.ErrorReason, error) {
	// TODO: Now Tezos client also acts as a comfirmation manager and listen the blockchain to get tx receipt.
	// FF tx manager should deal with it instead. This solution is temporary, for MVP purpose only.
	db, err := leveldb.OpenFile("/tmp/txs", nil)
	if err != nil {
		return nil, "", err
	}
	defer db.Close()

	receiptData, err := db.Get([]byte(req.TransactionHash), nil)
	if err != nil {
		return nil, ffcapi.ErrorReasonNotFound, err
	}

	var receipt rpc.Receipt
	err = json.Unmarshal(receiptData, &receipt)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	// TODO: reconsider getting block id from receipt
	blockNumber := receipt.Block.Int64()
	block, _, err := c.BlockInfoByHash(ctx, &ffcapi.BlockInfoByHashRequest{
		BlockHash: receipt.Block.String(),
	})
	if err != nil {
		fmt.Println("Error getting Block: ", err)
	} else {
		blockNumber = block.BlockNumber.Int64()
	}

	resp := &ffcapi.TransactionReceiptResponse{
		BlockNumber:      fftypes.NewFFBigInt(blockNumber),
		TransactionIndex: fftypes.NewFFBigInt(int64(receipt.Pos)),
		BlockHash:        receipt.Block.String(),
		Success:          receipt.IsSuccess(),
		ProtocolID:       receipt.Op.Protocol.String(),
		ExtraInfo:        nil,
		ContractLocation: nil,
	}

	return resp, "", nil
}
