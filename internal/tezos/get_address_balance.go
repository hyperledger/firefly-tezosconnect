package tezos

import (
	"context"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// AddressBalance gets the balance of the specified address
func (c *tezosConnector) AddressBalance(ctx context.Context, req *ffcapi.AddressBalanceRequest) (*ffcapi.AddressBalanceResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}
