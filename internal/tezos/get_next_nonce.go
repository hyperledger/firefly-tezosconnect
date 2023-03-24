package tezos

import (
	"context"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// NextNonceForSigner is used when there are no outstanding transactions for a given signing identity, to determine the next nonce to use for submission of a transaction
func (c *tezosConnector) NextNonceForSigner(ctx context.Context, req *ffcapi.NextNonceForSignerRequest) (*ffcapi.NextNonceForSignerResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}
