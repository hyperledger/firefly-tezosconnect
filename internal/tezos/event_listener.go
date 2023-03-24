package tezos

import (
	"sync"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-signer/pkg/ethtypes"
)

// listenerOptions are Tezos specific custom options that can be specified when creating a listener
type listenerOptions struct {
	// TODO: define for Tezos events
	// Methods []*abi.Entry `json:"methods,omitempty"` // An optional array of ABI methods. If specified and the input data for a transaction matches, the decoded inputs will be included in the event
	Signer bool `json:"signer,omitempty"` // An optional boolean for whether to extract the signer of the transaction that emitted the event
}

// listenerConfig is the configuration parsed from generic FFCAPI connector framework JSON, into our Tezos specific options
type listenerConfig struct {
	name      string
	fromBlock string
	options   *listenerOptions
	filters   []*eventFilter
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

type logJSONRPC struct {
	Removed bool `json:"removed"`
	// TODO: define Tezos types
	LogIndex         *ethtypes.HexInteger        `json:"logIndex"`
	TransactionIndex *ethtypes.HexInteger        `json:"transactionIndex"`
	BlockNumber      *ethtypes.HexInteger        `json:"blockNumber"`
	TransactionHash  ethtypes.HexBytes0xPrefix   `json:"transactionHash"`
	BlockHash        ethtypes.HexBytes0xPrefix   `json:"blockHash"`
	Address          *ethtypes.Address0xHex      `json:"address"`
	Data             ethtypes.HexBytes0xPrefix   `json:"data"`
	Topics           []ethtypes.HexBytes0xPrefix `json:"topics"`
}
