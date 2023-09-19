package tezos

import (
	"testing"

	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewBlockListenerOK(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	c.blockListener.blockPollingInterval = 1
	mRPC.On("GetHeadBlock", mock.Anything).Return(
		&rpc.Block{
			Hash: tezos.MustParseBlockHash("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg"),
			Header: rpc.BlockHeader{
				Predecessor: tezos.MustParseBlockHash("BLc1BjmZ7WevMoMoj8jxh4k2wLoRqoMUxjrQuDmKzAsApfRRjFL"),
				Level:       12345,
			},
		}, nil).Maybe()

	req := &ffcapi.NewBlockListenerRequest{
		ID:              fftypes.NewUUID(),
		ListenerContext: ctx,
		BlockListener:   make(chan<- *ffcapi.BlockHashEvent),
	}
	res, _, err := c.NewBlockListener(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
