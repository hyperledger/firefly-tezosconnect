package tezos

import (
	"context"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// eventStream is the state we hold in memory for each eventStream
type eventStream struct {
	id             *fftypes.UUID
	ctx            context.Context
	c              *tezosConnector
	events         chan<- *ffcapi.ListenerEvent
	listeners      map[fftypes.UUID]*listener
	headBlock      int64
	streamLoopDone chan struct{}
	catchup        bool
}

func (es *eventStream) addEventListener(ctx context.Context, req *ffcapi.EventListenerAddRequest) (*listener, error) {
	// TODO: impl
	return nil, nil
}

func (es *eventStream) startEventListener(l *listener) {
	// TODO: impl
}

func (es *eventStream) streamLoop() {
	defer close(es.streamLoopDone)

	for {
		// When we first start, we might find our leading pack of listeners are all way behind
		// the head of the chain. So we run a catchup mode loop to ensure we don't ask the blockchain
		// node to process an excessive amount of logs
		if es.leadGroupCatchup() {
			return
		}

		// We then transition to our steady state, filtering from the front of the chain.
		// But we might fall behind and need to go back to the catchup mode.
		if es.leadGroupSteadyState() {
			return
		}
	}
}

// leadGroupCatchup is called whenever the steam loop restarts, to see how far it is behind the head of the
// chain and if it's a way behind then we catch up all this head group as one set (rather than with individual
// catchup routines as is the case if one listener starts a way behind the pack)
//
//nolint:unparam
func (es *eventStream) leadGroupCatchup() bool {

	// For API status, we keep a track of whether we're in catchup mode or not
	es.catchup = true
	defer func() { es.catchup = false }()

	// TODO: impl
	return false
}

func (es *eventStream) leadGroupSteadyState() bool {
	// TODO: impl
	return true
}
