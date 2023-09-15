package tezos

import (
	"sync"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
)

// listenerConfig is the configuration parsed from generic FFCAPI connector framework JSON, into our Tezos specific options
type listenerConfig struct {
	name      string
	fromBlock string
	// options   *listenerOptions
	// filters   []*eventFilter
	signature string
}

// listener is the state we hold in memory for each individual listener that has been added
type listener struct {
	id              *fftypes.UUID
	c               *tezosConnector
	es              *eventStream
	hwmMux          sync.Mutex // Protects checkpoint of an individual listener. May hold ES lock when taking this, must NOT attempt to obtain ES lock while holding this
	hwmBlock        int64
	config          listenerConfig
	removed         bool
	catchup         bool
	catchupLoopDone chan struct{}
}
