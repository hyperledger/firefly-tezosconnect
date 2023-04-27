package tezos

import (
	"context"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// NextNonceForSigner. In Tezos, transactions are identified by their hash, which is unique for each transaction, so nonce is not necessary.
// Return zero value for compatibility
func (c *tezosConnector) NextNonceForSigner(ctx context.Context, req *ffcapi.NextNonceForSignerRequest) (*ffcapi.NextNonceForSignerResponse, ffcapi.ErrorReason, error) {
	return &ffcapi.NextNonceForSignerResponse{
		Nonce: fftypes.NewFFBigInt(0),
	}, "", nil
}
