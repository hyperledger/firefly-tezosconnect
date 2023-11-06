package tezos

import (
	"encoding/json"
	"errors"
	"testing"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/contract"
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

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).
		Return(&rpc.Receipt{
			Op: &rpc.Operation{
				Contents: []rpc.TypedOperation{
					rpc.Transaction{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Status: tezos.OpStatusApplied,
									},
								},
							},
						},
					},
				},
			},
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

func TestTransactionPrepareSimulateError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).Return(nil, errors.New("error"))

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

func TestTransactionPrepareWrongSimulateStatusError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).
		Return(&rpc.Receipt{
			Op: &rpc.Operation{
				Contents: []rpc.TypedOperation{
					rpc.Transaction{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Errors: []rpc.OperationError{
											{
												GenericError: rpc.GenericError{
													ID:   "error id: script_rejected",
													Kind: "error: script_rejected",
												},
												Raw: json.RawMessage{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
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

	_, reason, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonTransactionReverted)
}

func Test_estimateAndAssignTxCostIgnoreLimitsOk(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).
		Return(&rpc.Receipt{
			Op: &rpc.Operation{
				Contents: []rpc.TypedOperation{
					rpc.Transaction{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Status: tezos.OpStatusApplied,
									},
								},
							},
						},
					},
				},
			},
		}, nil)

	op := codec.NewOp()
	txArgs := contract.TxArgs{}
	op.WithContents(txArgs.Encode())

	opts := &rpc.DefaultOptions
	opts.IgnoreLimits = true

	_, err := c.estimateAndAssignTxCost(ctx, op, opts)
	assert.NoError(t, err)
}

func Test_estimateAndAssignExceedMaxLimitError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).
		Return(&rpc.Receipt{
			Op: &rpc.Operation{
				Contents: []rpc.TypedOperation{
					rpc.Transaction{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Status: tezos.OpStatusApplied,
									},
								},
							},
						},
					},
				},
			},
		}, nil)

	op := codec.NewOp()
	txArgs := contract.TxArgs{}
	op.WithContents(txArgs.Encode())
	op.WithLimits([]tezos.Limits{
		{
			Fee: 100,
		},
	}, 100)

	opts := &rpc.DefaultOptions
	opts.MaxFee = 1

	_, err := c.estimateAndAssignTxCost(ctx, op, opts)
	assert.Error(t, err)
}
