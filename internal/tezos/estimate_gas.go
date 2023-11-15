package tezos

import (
	"context"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

func (c *tezosConnector) GasEstimate(_ context.Context, _ *ffcapi.TransactionInput) (*ffcapi.GasEstimateResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}
