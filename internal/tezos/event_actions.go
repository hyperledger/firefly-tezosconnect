package tezos

import (
	"context"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// EventStreamStart starts an event stream with an initial set of listeners (which might be empty), a channel to deliver events, and a context that will close to stop the stream
func (c *tezosConnector) EventStreamStart(ctx context.Context, req *ffcapi.EventStreamStartRequest) (*ffcapi.EventStreamStartResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventStreamStopped informs a connector that an event stream has been requested to stop, and the context has been cancelled. So the state associated with it can be removed (and a future start of the same ID can be performed)
func (c *tezosConnector) EventStreamStopped(ctx context.Context, req *ffcapi.EventStreamStoppedRequest) (*ffcapi.EventStreamStoppedResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventListenerVerifyOptions validates the configuration options for a listener, applying any defaults needed by the connector, and returning the update options for FFTM to persist
func (c *tezosConnector) EventListenerVerifyOptions(ctx context.Context, req *ffcapi.EventListenerVerifyOptionsRequest) (*ffcapi.EventListenerVerifyOptionsResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventListenerAdd begins/resumes listening on set of events that must be consistently ordered. Blockchain specific signatures of the events are included, along with initial conditions (initial block number etc.), and the last stored checkpoint (if any)
func (c *tezosConnector) EventListenerAdd(ctx context.Context, req *ffcapi.EventListenerAddRequest) (*ffcapi.EventListenerAddResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventListenerRemove ends listening on a set of events previous started
func (c *tezosConnector) EventListenerRemove(ctx context.Context, req *ffcapi.EventListenerRemoveRequest) (*ffcapi.EventListenerRemoveResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventListenerHWM queries the current high water mark checkpoint for a listener. Called at regular intervals when there are no events in flight for a listener, to ensure checkpoint are written regularly even when there is no activity
func (c *tezosConnector) EventListenerHWM(ctx context.Context, req *ffcapi.EventListenerHWMRequest) (*ffcapi.EventListenerHWMResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventStreamNewCheckpointStruct used during checkpoint restore, to get the specific into which to restore the JSON bytes
func (c *tezosConnector) EventStreamNewCheckpointStruct() ffcapi.EventListenerCheckpoint {
	return nil
}
