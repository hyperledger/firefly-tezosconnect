package tezos

import (
	"context"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/trilitech/tzgo/tezos"
)

// AddressBalance gets the balance of the specified address
func (c *tezosConnector) AddressBalance(ctx context.Context, req *ffcapi.AddressBalanceRequest) (*ffcapi.AddressBalanceResponse, ffcapi.ErrorReason, error) {
	addr, err := tezos.ParseAddress(req.Address)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	headBlock, err := c.client.GetHeadBlock(ctx)
	if err != nil {
		return nil, "", err
	}

	balance, err := c.client.GetContractBalance(ctx, addr, headBlock.Hash)
	if err != nil {
		return nil, "", err
	}

	return &ffcapi.AddressBalanceResponse{
		Balance: (*fftypes.FFBigInt)(&balance),
	}, "", nil
}
