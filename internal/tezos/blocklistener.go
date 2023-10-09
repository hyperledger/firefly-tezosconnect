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

	var mon *rpc.BlockHeaderMonitor
	defer func() {
		if mon != nil {
			mon.Close()
		}
	}()
	failCount := 0
	gapPotential := true
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

		// (re)connect
		if mon == nil {
			mon = rpc.NewBlockHeaderMonitor()

			// register the block monitor with our client
			if err := bl.c.client.MonitorBlockHeader(bl.ctx, mon); err != nil {
				mon.Close()
				mon = nil
				if ErrorStatus(err) == 404 {
					log.L(bl.ctx).Errorf("monitor: event mode unsupported. %s", err.Error())
				} else {
					log.L(bl.ctx).Debugf("monitor: %s", err.Error())

					<-bl.ctx.Done()
					return
				}
				continue
			}
		}

		// wait for new block headers
		blockHead, err := mon.Recv(bl.ctx)
		// reconnect on error unless context was cancelled
		if err != nil {
			log.L(bl.ctx).Debugf("monitor: %s", err.Error())
			mon.Close()
			mon = nil

			gapPotential = true
			failCount++
			continue
		}

		update := &ffcapi.BlockHashEvent{GapPotential: gapPotential}
		var notifyPos *list.Element

		candidate := bl.reconcileCanonicalChain(blockHead)
		// Check this is the lowest position to notify from
		if candidate != nil && (notifyPos == nil || candidate.Value.(*minimalBlockInfo).number < notifyPos.Value.(*minimalBlockInfo).number) {
			notifyPos = candidate
		}

		if notifyPos != nil {
			// We notify for all hashes from the point of change in the chain onwards
			for notifyPos != nil {
				update.BlockHashes = append(update.BlockHashes, notifyPos.Value.(*minimalBlockInfo).hash)
				notifyPos = notifyPos.Next()
			}

			// Take a copy of the consumers in the lock
			bl.mux.Lock()
			consumers := make([]*blockUpdateConsumer, 0, len(bl.consumers))
			for _, c := range bl.consumers {
				consumers = append(consumers, c)
			}
			bl.mux.Unlock()

			// Spin through delivering the block update
			bl.dispatchToConsumers(consumers, update)
		}

		// Reset retry count when we have a full successful loop
		failCount = 0
		gapPotential = false
	}
}

// reconcileCanonicalChain takes an update on a block, and reconciles it against the in-memory view of the
// head of the canonical chain we have. If these blocks do not just fit onto the end of the chain, then we
// work backwards building a new view and notify about all blocks that are changed in that process.
func (bl *blockListener) reconcileCanonicalChain(bhle *rpc.BlockHeaderLogEntry) *list.Element {
	mbi := &minimalBlockInfo{
		number:     bhle.Level,
		hash:       bhle.Hash.String(),
		parentHash: bhle.Predecessor.String(),
	}
	bl.mux.Lock()
	if mbi.number > bl.highestBlock {
		bl.highestBlock = mbi.number
	}
	bl.mux.Unlock()

	// Find the position of this block in the block sequence
	pos := bl.canonicalChain.Back()
	for {
		if pos == nil || pos.Value == nil {
			// We've eliminated all the existing chain (if there was any)
			return bl.handleNewBlock(mbi, nil)
		}
		posBlock := pos.Value.(*minimalBlockInfo)
		switch {
		case posBlock.number == mbi.number && posBlock.hash == mbi.hash && posBlock.parentHash == mbi.parentHash:
			// This is a duplicate - no need to notify of anything
			return nil
		case posBlock.number == mbi.number:
			// We are replacing a block in the chain
			return bl.handleNewBlock(mbi, pos.Prev())
		case posBlock.number < mbi.number:
			// We have a position where this block goes
			return bl.handleNewBlock(mbi, pos)
		default:
			// We've not wound back to the point this block fits yet
			pos = pos.Prev()
		}
	}
}

// handleNewBlock rebuilds the canonical chain around a new block, checking if we need to rebuild our
// view of the canonical chain behind it, or trimming anything after it that is invalidated by a new fork.
func (bl *blockListener) handleNewBlock(mbi *minimalBlockInfo, addAfter *list.Element) *list.Element {

	// If we have an existing canonical chain before this point, then we need to check we've not
	// invalidated that with this block. If we have, then we have to re-verify our whole canonical
	// chain from the first block. Then notify from the earliest point where it has diverged.
	if addAfter != nil {
		prevBlock := addAfter.Value.(*minimalBlockInfo)
		if prevBlock.number != (mbi.number-1) || prevBlock.hash != mbi.parentHash {
			log.L(bl.ctx).Infof("Notified of block %d / %s that does not fit after block %d / %s (expected parent: %s)", mbi.number, mbi.hash, prevBlock.number, prevBlock.hash, mbi.parentHash)
			return bl.rebuildCanonicalChain()
		}
	}

	// Ok, we can add this block
	var newElem *list.Element
	if addAfter == nil {
		_ = bl.canonicalChain.Init()
		newElem = bl.canonicalChain.PushBack(mbi)
	} else {
		newElem = bl.canonicalChain.InsertAfter(mbi, addAfter)
		// Trim everything from this point onwards. Note that the following cases are covered on other paths:
		// - This was just a duplicate notification of a block that fits into our chain - discarded in reconcileCanonicalChain()
		// - There was a gap before us in the chain, and the tail is still valid - we would have called rebuildCanonicalChain() above
		nextElem := newElem.Next()
		for nextElem != nil {
			toRemove := nextElem
			nextElem = nextElem.Next()
			_ = bl.canonicalChain.Remove(toRemove)
		}
	}

	// Trim the amount of history we keep based on the configured amount of instability at the front of the chain
	for bl.canonicalChain.Len() > bl.unstableHeadLength {
		_ = bl.canonicalChain.Remove(bl.canonicalChain.Front())
	}

	log.L(bl.ctx).Debugf("Added block %d / %s parent=%s to in-memory canonical chain (new length=%d)", mbi.number, mbi.hash, mbi.parentHash, bl.canonicalChain.Len())

	return newElem
}

// rebuildCanonicalChain is called (only on non-empty case) when our current chain does not seem to line up with
// a recent block advertisement. So we need to work backwards to the last point of consistency with the current
// chain and re-query the chain state from there.
func (bl *blockListener) rebuildCanonicalChain() *list.Element {
	log.L(bl.ctx).Debugf("Rebuilding in-memory canonical chain")

	// If none of our blocks were valid, start from the first block number we've notified about previously
	lastValidBlock := bl.trimToLastValidBlock()
	var nextBlockNumber int64
	var expectedParentHash string
	if lastValidBlock != nil {
		nextBlockNumber = lastValidBlock.number + 1
		expectedParentHash = lastValidBlock.hash
	} else {
		firstBlock := bl.canonicalChain.Front()
		if firstBlock == nil || firstBlock.Value == nil {
			return nil
		}
		nextBlockNumber = firstBlock.Value.(*minimalBlockInfo).number
		// Clear out the whole chain
		bl.canonicalChain = bl.canonicalChain.Init()
	}
	var notifyPos *list.Element
	for {
		var bi *rpc.Block
		var reason ffcapi.ErrorReason
		err := bl.c.retry.Do(bl.ctx, "rebuild listener canonical chain", func(attempt int) (retry bool, err error) {
			bi, reason, err = bl.c.getBlockInfoByNumber(bl.ctx, nextBlockNumber, false, "")
			return reason != ffcapi.ErrorReasonNotFound, err
		})
		if err != nil {
			if reason != ffcapi.ErrorReasonNotFound {
				return nil // Context must have been cancelled
			}
		}
		if bi == nil {
			log.L(bl.ctx).Debugf("Block listener canonical chain view rebuilt to head at block %d", nextBlockNumber-1)
			break
		}
		mbi := &minimalBlockInfo{
			number:     bi.GetLevel(),
			hash:       bi.Hash.String(),
			parentHash: bi.Header.Predecessor.String(),
		}

		// It's possible the chain will change while we're doing this, and we fall back to the next block notification
		// to sort that out.
		if expectedParentHash != "" && mbi.parentHash != expectedParentHash {
			log.L(bl.ctx).Debugf("Block listener canonical chain view rebuilt up to new re-org at block %d", nextBlockNumber)
			break
		}
		expectedParentHash = mbi.hash
		nextBlockNumber++

		// Note we do not trim to a length here, as we need to notify for every block we haven't notified for.
		// Trimming to a length will happen when we get blocks that slot into our existing view
		newElem := bl.canonicalChain.PushBack(mbi)
		if notifyPos == nil {
			notifyPos = newElem
		}

		bl.mux.Lock()
		if mbi.number > bl.highestBlock {
			bl.highestBlock = mbi.number
		}
		bl.mux.Unlock()

	}
	return notifyPos
}

func (bl *blockListener) trimToLastValidBlock() (lastValidBlock *minimalBlockInfo) {
	// First remove from the end until we get a block that matches the current un-cached query view from the chain
	lastElem := bl.canonicalChain.Back()
	for lastElem != nil && lastElem.Value != nil {

		// Query the block that is no at this blockNumber
		currentViewBlock := lastElem.Value.(*minimalBlockInfo)
		var freshBlockInfo *rpc.Block
		var reason ffcapi.ErrorReason
		err := bl.c.retry.Do(bl.ctx, "rebuild listener canonical chain", func(attempt int) (retry bool, err error) {
			freshBlockInfo, reason, err = bl.c.getBlockInfoByNumber(bl.ctx, currentViewBlock.number, false, "")
			return reason != ffcapi.ErrorReasonNotFound, err
		})
		if err != nil {
			if reason != ffcapi.ErrorReasonNotFound {
				return nil // Context must have been cancelled
			}
		}

		if freshBlockInfo != nil && freshBlockInfo.Hash.String() == currentViewBlock.hash {
			log.L(bl.ctx).Debugf("Canonical chain matches current chain up to block %d", currentViewBlock.number)
			lastValidBlock = currentViewBlock
			// Trim everything after this point, as it's invalidated
			nextElem := lastElem.Next()
			for nextElem != nil {
				toRemove := lastElem
				nextElem = nextElem.Next()
				_ = bl.canonicalChain.Remove(toRemove)
			}
			break
		}
		lastElem = lastElem.Prev()

	}
	return lastValidBlock
}

func (bl *blockListener) dispatchToConsumers(consumers []*blockUpdateConsumer, update *ffcapi.BlockHashEvent) {
	for _, c := range consumers {
		log.L(bl.ctx).Tracef("Notifying consumer %s of blocks %v (gap=%t)", c.id, update.BlockHashes, update.GapPotential)
		select {
		case c.updates <- update:
		case <-bl.ctx.Done(): // loop, we're stopping and will exit on next loop
		case <-c.ctx.Done():
			log.L(bl.ctx).Debugf("Block update consumer %s closed", c.id)
			bl.mux.Lock()
			delete(bl.consumers, *c.id)
			bl.mux.Unlock()
		}
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
