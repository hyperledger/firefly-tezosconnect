package tezos

import (
	"container/list"
	"context"
	"sync"
	"time"

	"blockwatch.cc/tzgo/rpc"
	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

type blockUpdateConsumer struct {
	id      *fftypes.UUID // could be an event stream ID for example - must be unique
	ctx     context.Context
	updates chan<- *ffcapi.BlockHashEvent
}

// blockListener has two functions:
// 1) To establish and keep track of what the head block height of the blockchain is, so event streams know how far from the head they are
// 2) To feed new block information to any registered consumers
type blockListener struct {
	ctx                        context.Context
	c                          *tezosConnector
	listenLoopDone             chan struct{}
	initialBlockHeightObtained chan struct{}
	highestBlock               int64
	mux                        sync.Mutex
	consumers                  map[fftypes.UUID]*blockUpdateConsumer
	blockPollingInterval       time.Duration
	unstableHeadLength         int
	canonicalChain             *list.List
}

type minimalBlockInfo struct {
	number     int64
	hash       string
	parentHash string
}

func newBlockListener(ctx context.Context, c *tezosConnector, conf config.Section) *blockListener {
	bl := &blockListener{
		ctx:                        log.WithLogField(ctx, "role", "blocklistener"),
		c:                          c,
		initialBlockHeightObtained: make(chan struct{}),
		highestBlock:               -1,
		consumers:                  make(map[fftypes.UUID]*blockUpdateConsumer),
		blockPollingInterval:       conf.GetDuration(BlockPollingInterval),
		canonicalChain:             list.New(),
		unstableHeadLength:         int(c.checkpointBlockGap),
	}
	return bl
}

func (bl *blockListener) addConsumer(c *blockUpdateConsumer) {
	bl.mux.Lock()
	defer bl.mux.Unlock()
	bl.checkStartedLocked()
	bl.consumers[*c.id] = c
}

func (bl *blockListener) checkStartedLocked() {
	if bl.listenLoopDone == nil {
		bl.listenLoopDone = make(chan struct{})
		go bl.listenLoop()
	}
}

func (bl *blockListener) listenLoop() {
	defer close(bl.listenLoopDone)

	err := bl.establishBlockHeightWithRetry()
	close(bl.initialBlockHeightObtained)
	if err != nil {
		log.L(bl.ctx).Warnf("Block listener exiting before establishing initial block height: %s", err)
	}

	mon := rpc.NewBlockHeaderMonitor()
	defer mon.Close()

	// register the block monitor with our client
	if err := bl.c.client.MonitorBlockHeader(bl.ctx, mon); err != nil {
		log.L(bl.ctx).Error(err)
	}

	// var filter string
	failCount := 0
	for {
		if failCount > 0 {
			if bl.c.doFailureDelay(bl.ctx, failCount) {
				log.L(bl.ctx).Debugf("Block listener loop exiting")
				return
			}
		} else {
			// Sleep for the polling interval
			select {
			case <-time.After(bl.blockPollingInterval):
			case <-bl.ctx.Done():
				log.L(bl.ctx).Debugf("Block listener loop stopping")
				return
			}
		}

		//TODO: listen the chain and keep the blocks in the memory
	}
}

// getBlockHeightWithRetry keeps retrying attempting to get the initial block height until successful
func (bl *blockListener) establishBlockHeightWithRetry() error {
	return bl.c.retry.Do(bl.ctx, "get initial block height", func(attempt int) (retry bool, err error) {
		headBlock, err := bl.c.client.GetHeadBlock(bl.ctx)
		if err != nil {
			log.L(bl.ctx).Warnf("Block height could not be obtained: %s", err.Error())
			return true, err
		}

		bl.mux.Lock()
		bl.highestBlock = headBlock.GetLevel()
		bl.mux.Unlock()
		return false, nil
	})
}

func (bl *blockListener) getHighestBlock(ctx context.Context) int64 {
	bl.mux.Lock()
	bl.checkStartedLocked()
	highestBlock := bl.highestBlock
	bl.mux.Unlock()
	// if not yet initialized, wait to be initialized
	if highestBlock < 0 {
		select {
		case <-bl.initialBlockHeightObtained:
		case <-ctx.Done():
		}
	}
	bl.mux.Lock()
	highestBlock = bl.highestBlock
	bl.mux.Unlock()
	log.L(ctx).Debugf("ChainHead=%d", highestBlock)
	return highestBlock
}

func (bl *blockListener) waitClosed() {
	bl.mux.Lock()
	listenLoopDone := bl.listenLoopDone
	bl.mux.Unlock()
	if listenLoopDone != nil {
		<-listenLoopDone
	}
}
