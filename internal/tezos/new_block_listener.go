package tezos

import (
	"context"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// NewBlockListener creates a new block listener, decoupled from an event stream
func (c *tezosConnector) NewBlockListener(ctx context.Context, req *ffcapi.NewBlockListenerRequest) (*ffcapi.NewBlockListenerResponse, ffcapi.ErrorReason, error) {
	// Add the block consumer
	c.blockListener.addConsumer(&blockUpdateConsumer{
		id:      req.ID,
		ctx:     req.ListenerContext,
		updates: req.BlockListener,
	})

	return &ffcapi.NewBlockListenerResponse{}, "", nil
}
