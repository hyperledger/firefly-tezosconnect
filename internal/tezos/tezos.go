package tezos

import (
	"context"
	"math/big"
	"sync"
	"time"

	"blockwatch.cc/tzgo/rpc"
	"github.com/OneOf-Inc/firefly-tezosconnect/internal/msgs"
	lru "github.com/hashicorp/golang-lru"
	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-common/pkg/retry"
	"github.com/hyperledger/firefly-signer/pkg/abi"
	"github.com/hyperledger/firefly-signer/pkg/rpcbackend"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

type tezosConnector struct {
	backend                    rpcbackend.Backend
	serializer                 *abi.Serializer
	gasEstimationFactor        *big.Float
	catchupPageSize            int64
	catchupThreshold           int64
	checkpointBlockGap         int64
	retry                      *retry.Retry
	eventBlockTimestamps       bool
	blockListener              *blockListener
	eventFilterPollingInterval time.Duration

	client       *rpc.Client
	networkName  string
	signatoryURL string

	mux          sync.Mutex
	eventStreams map[fftypes.UUID]*eventStream
	blockCache   *lru.Cache
	txCache      *lru.Cache
}

func NewTezosConnector(ctx context.Context, conf config.Section) (cc ffcapi.API, err error) {
	c := &tezosConnector{
		eventStreams:               make(map[fftypes.UUID]*eventStream),
		catchupPageSize:            conf.GetInt64(EventsCatchupPageSize),
		catchupThreshold:           conf.GetInt64(EventsCatchupThreshold),
		checkpointBlockGap:         conf.GetInt64(EventsCheckpointBlockGap),
		eventBlockTimestamps:       conf.GetBool(EventsBlockTimestamps),
		eventFilterPollingInterval: conf.GetDuration(EventsFilterPollingInterval),
		retry: &retry.Retry{
			InitialDelay: conf.GetDuration(RetryInitDelay),
			MaximumDelay: conf.GetDuration(RetryMaxDelay),
			Factor:       conf.GetFloat64(RetryFactor),
		},
	}
	if c.catchupThreshold < c.catchupPageSize {
		log.L(ctx).Warnf("Catchup threshold %d must be at least as large as the catchup page size %d (overridden to %d)", c.catchupThreshold, c.catchupPageSize, c.catchupPageSize)
		c.catchupThreshold = c.catchupPageSize
	}
	c.blockCache, err = lru.New(conf.GetInt(BlockCacheSize))
	if err != nil {
		return nil, i18n.WrapError(ctx, err, msgs.MsgCacheInitFail, "block")
	}

	c.txCache, err = lru.New(conf.GetInt(TxCacheSize))
	if err != nil {
		return nil, i18n.WrapError(ctx, err, msgs.MsgCacheInitFail, "transaction")
	}

	rpcClientURL := conf.GetString(BlockchainRPC)
	if rpcClientURL == "" {
		return nil, i18n.WrapError(ctx, err, msgs.MsgMissingRpcUrl)
	}
	c.client, err = rpc.NewClient(conf.GetString(BlockchainRPC), nil)
	if err != nil {
		return nil, i18n.WrapError(ctx, err, msgs.MsgFailedRpcInitialization)
	}
	c.networkName = conf.GetString(BlockchainNetwork)
	c.signatoryURL = conf.GetString(BlockchainSignatory)

	c.blockListener = newBlockListener(ctx, c, conf)

	return c, nil
}

// WaitClosed can be called after cancelling all the contexts, to wait for everything to close down
func (c *tezosConnector) WaitClosed() {
	if c.blockListener != nil {
		c.blockListener.waitClosed()
	}
	for _, s := range c.eventStreams {
		<-s.streamLoopDone
	}
}
