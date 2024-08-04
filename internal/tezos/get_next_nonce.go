package tezos

import (
	"context"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

// NextNonceForSigner is used when there are no outstanding transactions for a given signing identity, to determine the next nonce to use for submission of a transaction
func (c *tezosConnector) NextNonceForSigner(ctx context.Context, req *ffcapi.NextNonceForSignerRequest) (*ffcapi.NextNonceForSignerResponse, ffcapi.ErrorReason, error) {
	state, err := c.client.GetContractExt(ctx, tezos.MustParseAddress(req.Signer), rpc.Head)
	if err != nil {
		return nil, "", err
	}

	nextCounter := state.Counter + 1

	return &ffcapi.NextNonceForSignerResponse{
		Nonce: fftypes.NewFFBigInt(nextCounter),
	}, "", nil
}
