package tezos

import (
	"context"
	"errors"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

func (t *tezosConnector) GasEstimate(ctx context.Context, transaction *ffcapi.TransactionInput) (*ffcapi.GasEstimateResponse, ffcapi.ErrorReason, error) {
	// TODO: implement
	return nil, ffcapi.ErrorReason("Not implemented"), errors.New("Not implemented")
}
