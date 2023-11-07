package tezos

import (
	"errors"
	"testing"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
)

func TestMapError(t *testing.T) {
	testCases := []struct {
		name        string
		methodType  tezosRPCMethodCategory
		err         error
		errorReason ffcapi.ErrorReason
	}{
		{
			name:        "BlockRPCMethods with 404 Status error",
			methodType:  blockRPCMethods,
			err:         errors.New("status 404"),
			errorReason: ffcapi.ErrorReasonNotFound,
		},
		{
			name:        "SendRPCMethods with counter_in_the_past error",
			methodType:  sendRPCMethods,
			err:         errors.New("counter_in_the_past"),
			errorReason: ffcapi.ErrorReasonNonceTooLow,
		},
		{
			name:        "BlockRPCMethods with some non 404 Status error",
			methodType:  blockRPCMethods,
			err:         errors.New("error"),
			errorReason: ffcapi.ErrorReason(""),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := mapError(tc.methodType, tc.err)

			assert.Equal(t, tc.errorReason, res)
		})
	}
}

func TestErrorStatusNonHttpError(t *testing.T) {
	errorStatus := ErrorStatus(errors.New("error"))

	assert.Equal(t, errorStatus, 0)
}

func TestErrorStatusHttpError(t *testing.T) {
	err := httpError{
		request:    "request",
		status:     "status",
		statusCode: 400,
		body:       []byte("body"),
	}

	errorStatus := ErrorStatus(&err)

	assert.Equal(t, errorStatus, err.StatusCode())
	assert.Equal(t, err.Status(), "status")
	assert.Equal(t, err.StatusCode(), 400)
	assert.Equal(t, err.Request(), "request")
	assert.Equal(t, err.Body(), []byte("body"))
	assert.Equal(t, err.Error(), "rpc: request status 400 (body)")
}
