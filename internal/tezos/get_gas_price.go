package tezos

import (
	"context"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// GasPriceEstimate provides a blockchain specific gas price estimate
func (c *tezosConnector) GasPriceEstimate(ctx context.Context, req *ffcapi.GasPriceEstimateRequest) (*ffcapi.GasPriceEstimateResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}
