package tezos

import (
	"context"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// IsLive confirms if the connector up and running
func (c *tezosConnector) IsLive(_ context.Context) (*ffcapi.LiveResponse, ffcapi.ErrorReason, error) {
	return &ffcapi.LiveResponse{
		Up: true,
	}, "", nil
}

// IsReady confirms if the connector is connected to the downstream JSONRPC endpoint and therefore ready to receive traffic
func (c *tezosConnector) IsReady(ctx context.Context) (*ffcapi.ReadyResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}
