package tezos

import (
	"context"
	"fmt"

	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// TransactionReceipt queries to see if a receipt is available for a given transaction hash
func (c *tezosConnector) TransactionReceipt(ctx context.Context, req *ffcapi.TransactionReceiptRequest) (*ffcapi.TransactionReceiptResponse, ffcapi.ErrorReason, error) {
	fmt.Println("TRANSACTION RECEIPT REQ")

	// wait for confirmations
	res := rpc.NewResult(tezos.MustParseOpHash(req.TransactionHash)) //.WithTTL(op.TTL).WithConfirmations(opts.Confirmations)

	mon := c.client.BlockObserver

	// ensure block observer is running
	mon.Listen(c.client)

	// wait for confirmations
	res.Listen(mon)
	res.WaitContext(ctx)
	if err := res.Err(); err != nil {
		return nil, "", err
	}

	// return receipt
	receipt, err := res.GetReceipt(ctx)
	if err != nil {
		return nil, "", err
	}

	blockNumber := receipt.Block.Int64()
	block, _, err := c.BlockInfoByHash(ctx, &ffcapi.BlockInfoByHashRequest{
		BlockHash: receipt.Block.String(),
	})
	if err != nil {
		log.L(ctx).Error("error getting block: ", err)
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
