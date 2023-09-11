package tezos

import (
	"testing"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
)

func TestGetGasPriceOK(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	var req ffcapi.GasPriceEstimateRequest
	res, reason, err := c.GasPriceEstimate(ctx, &req)
	assert.NoError(t, err)
	assert.Empty(t, reason)
	assert.Equal(t, `"0"`, res.GasPrice.String())
}
