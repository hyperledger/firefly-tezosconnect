package tezos

import (
	"context"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// BlockInfoByNumber gets block information from the specified position (block number/index) in the canonical chain currently known to the local node
func (c *tezosConnector) BlockInfoByNumber(ctx context.Context, req *ffcapi.BlockInfoByNumberRequest) (*ffcapi.BlockInfoByNumberResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// BlockInfoByHash gets block information using the hash of the block
func (c *tezosConnector) BlockInfoByHash(ctx context.Context, req *ffcapi.BlockInfoByHashRequest) (*ffcapi.BlockInfoByHashResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}
