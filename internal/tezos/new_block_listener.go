package tezos

import (
	"context"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// NewBlockListener creates a new block listener, decoupled from an event stream
func (c *tezosConnector) NewBlockListener(ctx context.Context, req *ffcapi.NewBlockListenerRequest) (*ffcapi.NewBlockListenerResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}
