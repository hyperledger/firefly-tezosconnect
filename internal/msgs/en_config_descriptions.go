package msgs

import (
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"golang.org/x/text/language"
)

var ffc = func(key, translation string, fieldType string) i18n.ConfigMessageKey {
	return i18n.FFC(language.AmericanEnglish, key, translation, fieldType)
}

//revive:disable
var (
	ConfigTezosURL                    = ffc("config.connector.url", "URL of JSON/RPC endpoint for the Tezos node/gateway", "string")
	ConfigTezosDataFormat             = ffc("config.connector.dataFormat", "Configure the JSON data format for query output and events", "map,flat_array,self_describing")
	ConfigTezosGasEstimationFactor    = ffc("config.connector.gasEstimationFactor", "The factor to apply to the gas estimation to determine the gas limit", "float")
	ConfigBlockCacheSize              = ffc("config.connector.blockCacheSize", "Maximum of blocks to hold in the block info cache", i18n.IntType)
	ConfigBlockPollingInterval        = ffc("config.connector.blockPollingInterval", "Interval for polling to check for new blocks", i18n.TimeDurationType)
	ConfigEventsBlockTimestamps       = ffc("config.connector.events.blockTimestamps", "Whether to include the block timestamps in the event information", i18n.BooleanType)
	ConfigEventsCatchupPageSize       = ffc("config.connector.events.catchupPageSize", "Number of blocks to query per poll when catching up to the head of the blockchain", i18n.IntType)
	ConfigEventsCatchupThreshold      = ffc("config.connector.events.catchupThreshold", "How many blocks behind the chain head an event stream or listener must be on startup, to enter catchup mode", i18n.IntType)
	ConfigEventsCheckpointBlockGap    = ffc("config.connector.events.checkpointBlockGap", "The number of blocks at the head of the chain that should be considered unstable (could be dropped from the canonical chain after a re-org). Unless events with a full set of confirmations are detected, the restart checkpoint will this many blocks behind the chain head.", i18n.IntType)
	ConfigEventsFilterPollingInterval = ffc("config.connector.events.filterPollingInterval", "The interval between polling calls to a filter, when checking for newly arrived events", i18n.TimeDurationType)
	ConfigTxCacheSize                 = ffc("config.connector.txCacheSize", "Maximum of transactions to hold in the transaction info cache", i18n.IntType)
)
