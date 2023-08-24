package tezos

import (
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
