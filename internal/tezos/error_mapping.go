package tezos

import (
	"fmt"
	"strings"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

type tezosRPCMethodCategory int

const (
	blockRPCMethods tezosRPCMethodCategory = iota
	sendRPCMethods
)

// mapErrorToReason provides a common place for mapping Tezos client
// error strings, to a more consistent set of cross-client (and
// cross blockchain) reasons for errors defined by FFCPI for use by
// FireFly Transaction Manager.
func mapError(methodType tezosRPCMethodCategory, err error) ffcapi.ErrorReason {
	errString := strings.ToLower(err.Error())

	switch methodType {
	case blockRPCMethods:
		if strings.Contains(errString, "status 404") {
			return ffcapi.ErrorReasonNotFound
		}
	case sendRPCMethods:
		switch {
		case strings.Contains(errString, "counter_in_the_past"):
			return ffcapi.ErrorReasonNonceTooLow
		}
	}

	// Best default in FFCAPI is to provide no mapping
	return ""
}

func ErrorStatus(err error) int {
	switch e := err.(type) {
	case *httpError:
		return e.statusCode
	default:
		return 0
	}
}

type httpError struct {
	request    string
	status     string
	statusCode int
	body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("rpc: %s status %d (%v)", e.request, e.statusCode, string(e.body))
}

func (e *httpError) Request() string {
	return e.request
}

func (e *httpError) Status() string {
	return e.status
}

func (e *httpError) StatusCode() int {
	return e.statusCode
}

func (e *httpError) Body() []byte {
	return e.body
}
