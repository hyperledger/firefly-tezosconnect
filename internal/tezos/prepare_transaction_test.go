package tezos

import (
	"errors"
	"testing"

	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionPrepareOk(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	res, reason, err := c.TransactionPrepare(ctx, req)
	assert.NoError(t, err)
	assert.Empty(t, reason)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.TransactionData)
}

func TestTransactionPrepareWrongParamsError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("wrong"),
			},
		},
	}
	_, reason, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
	assert.Regexp(t, "FF23014", err)
	assert.Equal(t, ffcapi.ErrorReasonInvalidInputs, reason)
}

func TestTransactionPrepareWrongToAddressError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "wrong",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	_, _, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
	assert.Regexp(t, "FF23020", err)
}

func TestTransactionPrepareWrongFromAddressError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "wrong",
				To:   "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	_, _, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
	assert.Regexp(t, "FF23019", err)
}

func TestTransactionPrepareGetContractExtError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(nil, errors.New("error"))

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	_, _, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
}
