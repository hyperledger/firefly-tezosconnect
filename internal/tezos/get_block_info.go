package tezos

import (
	"context"

	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/OneOf-Inc/firefly-tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// BlockInfoByNumber gets block information from the specified position (block number/index) in the canonical chain currently known to the local node
func (c *tezosConnector) BlockInfoByNumber(ctx context.Context, req *ffcapi.BlockInfoByNumberRequest) (*ffcapi.BlockInfoByNumberResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// BlockInfoByHash gets block information using the hash of the block
func (c *tezosConnector) BlockInfoByHash(ctx context.Context, req *ffcapi.BlockInfoByHashRequest) (*ffcapi.BlockInfoByHashResponse, ffcapi.ErrorReason, error) {
	blockInfo, err := c.getBlockInfoByHash(ctx, req.BlockHash)
	if err != nil {
		return nil, ffcapi.ErrorReason(""), err
	}
	if blockInfo == nil {
		return nil, ffcapi.ErrorReasonNotFound, i18n.NewError(ctx, msgs.MsgBlockNotAvailable)
	}

	res := &ffcapi.BlockInfoByHashResponse{}
	transformBlockInfo(blockInfo, &res.BlockInfo)
	return res, "", nil
}

func (c *tezosConnector) getBlockInfoByHash(ctx context.Context, hashString string) (*rpc.Block, error) {
	var blockInfo *rpc.Block

	blockHash, err := tezos.ParseBlockHash(hashString)
	if err != nil {
		return nil, err
	}

	blockInfo, err = c.client.GetBlock(ctx, blockHash)
	if err != nil {
		return nil, err
	}

	return blockInfo, nil
}

func transformBlockInfo(bi *rpc.Block, t *ffcapi.BlockInfo) {
	t.BlockNumber = fftypes.NewFFBigInt(bi.Header.Level)
	t.BlockHash = bi.Hash.String()
	t.ParentHash = bi.Header.Predecessor.String()
	stringHashes := make([]string, 0)
	for _, batch := range bi.Operations {
		for _, tx := range batch {
			stringHashes = append(stringHashes, tx.Hash.String())
		}
	}
	t.TransactionHashes = stringHashes
}
