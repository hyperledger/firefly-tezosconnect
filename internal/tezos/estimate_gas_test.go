package tezos

import (
	"testing"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
)

func TestGasEstimate(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	resp, reason, err := c.GasEstimate(ctx, &ffcapi.TransactionInput{})
	assert.Nil(t, resp)
	assert.Empty(t, reason)
	assert.NoError(t, err)
}
