package tezos

import (
	"context"
	"sync"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-signer/pkg/ethtypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// eventFilter is Tezos specific filter options - an array of these can be configured on each listener
type eventFilter struct {
	// TODO: define for Tezos events
	// Event     *abi.Entry                `json:"event"`             // The ABI spec of the event to listen to
	Address   *ethtypes.Address0xHex    `json:"address,omitempty"` // An optional address to restrict the
	Topic0    ethtypes.HexBytes0xPrefix `json:"topic0"`            // Topic 0 match
	Signature string                    `json:"signature"`         // The cached signature of this event
}

// eventInfo is the top-level structure we pass to applications for each event (through the FFCAPI framework)
type eventInfo struct {
	logJSONRPC
	InputMethod string                 `json:"inputMethod,omitempty"` // the method invoked, if it matched one of the signatures in the listener definition
	InputArgs   *fftypes.JSONAny       `json:"inputArgs,omitempty"`   // the method parameters, if the method matched one of the signatures in the listener definition
	InputSigner *ethtypes.Address0xHex `json:"inputSigner,omitempty"` // the signing `from` address of the transaction
}

// eventStream is the state we hold in memory for each eventStream
type eventStream struct {
	id             *fftypes.UUID
	ctx            context.Context
	c              *tezosConnector
	events         chan<- *ffcapi.ListenerEvent
	mux            sync.Mutex
	updateCount    int
	listeners      map[fftypes.UUID]*listener
	headBlock      int64
	streamLoopDone chan struct{}
	catchup        bool
}
