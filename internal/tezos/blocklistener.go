package tezos

import (
	"container/list"
	"context"
	"sync"
	"time"

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

func (bl *blockListener) waitClosed() {
	bl.mux.Lock()
	listenLoopDone := bl.listenLoopDone
	bl.mux.Unlock()
	if listenLoopDone != nil {
		<-listenLoopDone
	}
}
