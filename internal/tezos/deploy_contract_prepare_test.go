package tezos

import (
	"context"
	"testing"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
)

func TestDeployContractPrepare(t *testing.T) {
	_, c, _, done := newTestConnector(t)
	defer done()

	_, _, err := c.DeployContractPrepare(context.Background(), &ffcapi.ContractDeployPrepareRequest{})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "contract deployment is not supported")
}
