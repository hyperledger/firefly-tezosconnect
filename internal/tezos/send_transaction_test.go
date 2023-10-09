package tezos

import (
	"encoding/json"
	"errors"
	"testing"

	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionSendDecodeStrError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionSendRequest{
		TransactionData: "1",
	}
	res, reason, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
	assert.Nil(t, res)
}

func TestTransactionSendDecodeOpError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionSendRequest{}
	res, reason, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
	assert.Nil(t, res)
}

func TestTransactionSendWrongFromAddressError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionSendRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "wrong",
		},
		TransactionData: "424d426559724d4a704c577271437337555463466155514365574271736a434c6c00889816a17ae688c971be1ad34bfe1990f8fa5e0f000b0000000130a980e6e41028da2cacfca4ddefea252d18bed900ffff05706175736500000002030a",
	}
	res, _, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
	assert.Regexp(t, "FF23019", err)
	assert.Nil(t, res)
}

func TestTransactionSendGetContractExtError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(nil, errors.New("error"))

	req := &ffcapi.TransactionSendRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
			To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
		},
		TransactionData: "424d426559724d4a704c577271437337555463466155514365574271736a434c6c00889816a17ae688c971be1ad34bfe1990f8fa5e0f000b0000000130a980e6e41028da2cacfca4ddefea252d18bed900ffff05706175736500000002030a",
	}
	_, _, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
}

func TestTransactionSendSimulateError(t *testing.T) {
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

	req := &ffcapi.TransactionSendRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
			To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
		},
		TransactionData: "424d426559724d4a704c577271437337555463466155514365574271736a434c6c00889816a17ae688c971be1ad34bfe1990f8fa5e0f000b0000000130a980e6e41028da2cacfca4ddefea252d18bed900ffff05706175736500000002030a",
	}
	_, _, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
}

func TestTransactionSendWrongSimulateStatusError(t *testing.T) {
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
													ID:   "error id",
													Kind: "error",
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

	req := &ffcapi.TransactionSendRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
			To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
		},
		TransactionData: "424d426559724d4a704c577271437337555463466155514365574271736a434c6c00889816a17ae688c971be1ad34bfe1990f8fa5e0f000b0000000130a980e6e41028da2cacfca4ddefea252d18bed900ffff05706175736500000002030a",
	}
	_, _, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
}
