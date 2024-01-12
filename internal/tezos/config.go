package tezos

import (
	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/hyperledger/firefly-common/pkg/ffresty"
)

const (
	ConfigGasEstimationFactor   = "gasEstimationFactor"
	ConfigDataFormat            = "dataFormat"
	BlockPollingInterval        = "blockPollingInterval"
	BlockCacheSize              = "blockCacheSize"
	EventsCatchupPageSize       = "events.catchupPageSize"
	EventsCatchupThreshold      = "events.catchupThreshold"
	EventsCheckpointBlockGap    = "events.checkpointBlockGap"
	EventsBlockTimestamps       = "events.blockTimestamps"
	EventsFilterPollingInterval = "events.filterPollingInterval"
	RetryInitDelay              = "retry.initialDelay"
	RetryMaxDelay               = "retry.maxDelay"
	RetryFactor                 = "retry.factor"
	TxCacheSize                 = "txCacheSize"
	BlockchainRPC               = "blockchain.rpc"
	BlockchainNetwork           = "blockchain.network"
	BlockchainSignatory         = "blockchain.signatory"
)

const (
	DefaultListenerPort        = 5102
	DefaultGasEstimationFactor = 1.5

	DefaultCatchupPageSize          = 500
	DefaultEventsCatchupThreshold   = 500
	DefaultEventsCheckpointBlockGap = 50

	DefaultRetryInitDelay   = "100ms"
	DefaultRetryMaxDelay    = "30s"
	DefaultRetryDelayFactor = 2.0
)

func InitConfig(conf config.Section) {
	ffresty.InitConfig(conf)
	conf.AddKnownKey(BlockCacheSize, 250)
	conf.AddKnownKey(BlockPollingInterval, "1s")
	conf.AddKnownKey(ConfigDataFormat, "map")
	conf.AddKnownKey(ConfigGasEstimationFactor, DefaultGasEstimationFactor)
	conf.AddKnownKey(EventsBlockTimestamps, true)
	conf.AddKnownKey(EventsFilterPollingInterval, "1s")
	conf.AddKnownKey(EventsCatchupPageSize, DefaultCatchupPageSize)
	conf.AddKnownKey(EventsCatchupThreshold, DefaultEventsCatchupThreshold)
	conf.AddKnownKey(EventsCheckpointBlockGap, DefaultEventsCheckpointBlockGap)
	conf.AddKnownKey(RetryFactor, DefaultRetryDelayFactor)
	conf.AddKnownKey(RetryInitDelay, DefaultRetryInitDelay)
	conf.AddKnownKey(RetryMaxDelay, DefaultRetryMaxDelay)
	conf.AddKnownKey(TxCacheSize, 250)
	conf.AddKnownKey(BlockchainRPC)
	conf.AddKnownKey(BlockchainNetwork, "mainnet")
	conf.AddKnownKey(BlockchainSignatory)
}
