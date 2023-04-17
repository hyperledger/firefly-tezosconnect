package tezos

import (
	"context"

	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// AddressBalance gets the balance of the specified address
func (c *tezosConnector) AddressBalance(ctx context.Context, req *ffcapi.AddressBalanceRequest) (*ffcapi.AddressBalanceResponse, ffcapi.ErrorReason, error) {
	addr, err := tezos.ParseAddress(req.Address)
	if err != nil {
		return nil, "Invalid address", err
	}

	headBlock, err := c.client.GetHeadBlock(ctx)
	if err != nil {
		return nil, "GetHeadBlock error", err
	}

	balance, err := c.client.GetContractBalance(ctx, addr, headBlock.Hash)
	if err != nil {
		return nil, "GetContractBalance error", err
	}

	return &ffcapi.AddressBalanceResponse{
		Balance: (*fftypes.FFBigInt)(&balance),
	}, "", nil
}
