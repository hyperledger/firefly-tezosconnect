package tezos

import (
	"context"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/rpc"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// QueryInvoke executes a method on a blockchain smart contract, which might execute Smart Contract code, but does not affect the blockchain state.
func (c *tezosConnector) QueryInvoke(_ context.Context, req *ffcapi.QueryInvokeRequest) (*ffcapi.QueryInvokeResponse, ffcapi.ErrorReason, error) {
	// TODO: to implement
	return nil, "", nil
}

func (c *tezosConnector) callTransaction(ctx context.Context, op *codec.Op, opts *rpc.CallOptions) (*rpc.Receipt, ffcapi.ErrorReason, error) {
	sim, err := c.client.Simulate(ctx, op, opts)
	if err != nil {
		return nil, mapError(callRPCMethods, err), err
	}
	// fail with Tezos error when simulation failed
	if !sim.IsSuccess() {
		return nil, mapError(callRPCMethods, sim.Error()), sim.Error()
	}

	return sim, "", nil
}
