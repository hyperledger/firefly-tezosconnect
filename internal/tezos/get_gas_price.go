package tezos

import (
	"context"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// In Tezos, there is no concept of "gas price" as it exists in EVM chains.
// Tezos uses a different mechanism for handling transaction costs and execution fees.
// Instead of gas and gas prices, Tezos employs a concept known as "transaction fees."
func (c *tezosConnector) GasPriceEstimate(ctx context.Context, _ *ffcapi.GasPriceEstimateRequest) (*ffcapi.GasPriceEstimateResponse, ffcapi.ErrorReason, error) {
	return &ffcapi.GasPriceEstimateResponse{
		GasPrice: fftypes.JSONAnyPtr(`"0"`),
	}, "", nil
}
