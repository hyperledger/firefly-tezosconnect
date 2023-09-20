package tezos

import (
	"context"
	"errors"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

func (c *tezosConnector) GasEstimate(_ context.Context, _ *ffcapi.TransactionInput) (*ffcapi.GasEstimateResponse, ffcapi.ErrorReason, error) {
	// TODO: implement
	return nil, ffcapi.ErrorReason("not implemented"), errors.New("not implemented")
}
