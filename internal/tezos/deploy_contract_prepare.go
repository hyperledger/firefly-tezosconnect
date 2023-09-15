package tezos

import (
	"context"
	"errors"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

func (c *tezosConnector) DeployContractPrepare(ctx context.Context, req *ffcapi.ContractDeployPrepareRequest) (*ffcapi.TransactionPrepareResponse, ffcapi.ErrorReason, error) {
	return nil, "", errors.New("contract deployment is not supported")
}
