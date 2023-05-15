package tracker

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/xerrors"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/chain"
	"github.com/x-xyz/goapi/service/query"
)

var metOnce sync.Once
var met metrics.Service

type CurrentBlockProvider interface {
	BlockNumber(context.Context) (uint64, error)
}

type EventHandler interface {
	GetFilterTopics() [][]common.Hash
	ProcessEvents(bCtx.Ctx, []logWithBlockTime) error
}

const Version = 1
const CaughtUpBlock = 5
const TooManyLogsTimeout = 30 * time.Second

type EventTrackerCfg struct {
	ChainId             int64
	BlockTime           time.Duration
	CurrentBlockGetter  CurrentBlockProvider
	Mongo               query.Mongo
	WsClient            domain.EthClientRepo
	RpcClient           domain.EthClientRepo
	ClientWithArchive   domain.EthClientRepo
	TrackerStateUseCase domain.TrackerStateUseCase
	BlockUseCase        chain.BlockUseCase

	// contract address = 0x0000000000000000000000000000000000000000 means get events from all addresses
	ContractAddress common.Address

	EventHandl         EventHandler
	ErrorCh            chan<- error
	SkipMissingBlock   bool
	TrackerTag         string
	ShouldDecodeSender bool
	FollowDistance     uint64
}

type EventTracker struct {
	chainId             int64
	blockTime           time.Duration
	currentBlockGetter  CurrentBlockProvider
	q                   query.Mongo
	wsClient            domain.EthClientRepo
	rpcClient           domain.EthClientRepo
	clientWithArchive   domain.EthClientRepo
	signer              types.Signer
	trackerStateUseCase domain.TrackerStateUseCase
	blockUseCase        chain.BlockUseCase
	contractAddress     common.Address
	eventHandler        EventHandler
	errorCh             chan<- error
	skipMissingBlock    bool
	filter              ethereum.FilterQuery
	trackerState        *domain.TrackerState
	trackerTag          string
	shouldDecodeSender  bool
	followDistance      uint64
	stoppedCh           chan interface{}
}

func NewEventTracker(cfg *EventTrackerCfg) (*EventTracker, error) {
	metOnce.Do(func() {
		met = metrics.New("tracker")
	})
	filter := ethereum.FilterQuery{
		Topics: cfg.EventHandl.GetFilterTopics(),
	}
	if domain.EmptyAddress.Equals(domain.Address(cfg.ContractAddress.String())) {
		if !cfg.SkipMissingBlock {
			return nil, errors.New("config error: SkipMissingBlock must be true when tracking all addresses")
		}
	} else {
		filter.Addresses = []common.Address{cfg.ContractAddress}
	}
	signer := types.LatestSignerForChainID(new(big.Int).SetInt64(cfg.ChainId))
	return &EventTracker{
		chainId:             cfg.ChainId,
		blockTime:           cfg.BlockTime,
		currentBlockGetter:  cfg.CurrentBlockGetter,
		q:                   cfg.Mongo,
		wsClient:            cfg.WsClient,
		rpcClient:           cfg.RpcClient,
		clientWithArchive:   cfg.ClientWithArchive,
		signer:              signer,
		trackerStateUseCase: cfg.TrackerStateUseCase,
		blockUseCase:        cfg.BlockUseCase,
		contractAddress:     cfg.ContractAddress,
		eventHandler:        cfg.EventHandl,
		errorCh:             cfg.ErrorCh,
		skipMissingBlock:    cfg.SkipMissingBlock,
		trackerTag:          cfg.TrackerTag,
		shouldDecodeSender:  cfg.ShouldDecodeSender,
		followDistance:      cfg.FollowDistance,
		filter:              filter,
		stoppedCh:           make(chan interface{}),
	}, nil
}

func (f *EventTracker) Start(ctx bCtx.Ctx) {
	go func() {
		defer close(f.stoppedCh)
		if err := f.loop(ctx); err != nil {
			f.errorCh <- err
		}
	}()
}

func (f *EventTracker) Wait() {
	<-f.stoppedCh
}

func (f *EventTracker) loop(ctx bCtx.Ctx) error {
	if !f.skipMissingBlock {
		state, err := f.setupTrackerState(ctx)
		if err != nil {
			ctx.WithField("err", err).Error("setupTrackerState failed")
			return err
		}
		f.trackerState = state
	} else {
		// set dummy tracker state to current block
		current, err := f.currentBlockGetter.BlockNumber(ctx)
		if err != nil {
			return err
		}
		f.trackerState = &domain.TrackerState{
			ChainId:               domain.ChainId(f.chainId),
			ContractAddress:       domain.Address(ToLowerHexStr(f.contractAddress)),
			Tag:                   f.trackerTag,
			Version:               Version,
			LastBlockProcessed:    current + 1,
			LastLogIndexProcessed: -1,
		}
	}

	// fast fetch
	if err := f.fastFetch(ctx); err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"chainId":  f.chainId,
			"contract": f.contractAddress,
		}).Error(fmt.Sprintf("fastFetch failed: %s err=%s", f.contractAddress.String(), err.Error()))
		return err
	}

	ch := make(chan types.Log, 1024)
	// remove from/to blocks is required
	filter := ethereum.FilterQuery{
		Addresses: f.filter.Addresses,
		Topics:    f.filter.Topics,
	}
	sub, err := f.wsClient.SubscribeFilterLogs(ctx, filter, ch)
	if err != nil {
		ctx.WithField("err", err).Error("client.SubscribeFilterLogs failed")
		return err
	}
	defer sub.Unsubscribe()
	ctx.WithField("contract", f.contractAddress).Info("subscription")

	// set dummy pending, so we won't miss the logs between last process block ~ current block
	current, err := f.currentBlockGetter.BlockNumber(ctx)
	if err != nil {
		return err
	}
	met.BumpAvg("blockchain.lastBlock", float64(current), "chainId", fmt.Sprint(f.chainId))
	lastPending := current
	pending := []uint64{current}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			ctx.WithField("err", err).Error("sub.Err()")
			return err
		case l := <-ch:
			// add log block number to pending, and wait for confirmation (follow distance)
			if l.BlockNumber < lastPending {
				ctx.WithFields(log.Fields{
					"contract":         f.contractAddress,
					"log_block_number": l.BlockNumber,
					"last_pending":     lastPending,
				}).Warn("received old logs")
			}

			if l.BlockNumber > lastPending {
				lastPending = l.BlockNumber
				pending = append(pending, l.BlockNumber)
			}

			ctx.WithFields(log.Fields{
				"contract":         f.contractAddress,
				"log_block_number": l.BlockNumber,
				"last_pending":     lastPending,
				"numPending":       len(pending),
			}).Info("receive log")

		case <-ticker.C:
			// no pending event
			if len(pending) == 0 {
				continue
			}

			current, err := f.currentBlockGetter.BlockNumber(ctx)
			if err != nil {
				ctx.WithField("err", err).Error("currentBlockGetter.BlockNumber failed")
				return err
			}
			met.BumpAvg("blockchain.lastBlock", float64(current), "chainId", fmt.Sprint(f.chainId))
			target := current - f.followDistance

			// keep waiting
			if pending[0] > target {
				continue
			}

			start := f.trackerState.LastBlockProcessed
			end := target
			if end < start {
				continue
			}

			blkRange := newBlockRange(start, end)
			err = f.processBlkRange(ctx, blkRange)
			if err != nil {
				ctx.WithField("err", err).Error("f.processBlkRange failed")
				return err
			}
			ctx.Info(fmt.Sprintf("process block range start=%d end=%d last=%d contract=%s", start, end, f.trackerState.LastBlockProcessed, f.contractAddress.String()))
			met.BumpAvg("collection.lastBlock", float64(f.trackerState.LastBlockProcessed), "chainId", fmt.Sprint(f.chainId), "contract", f.contractAddress.String())

			// remove pending <= target
			i := 0
			for _, p := range pending {
				if p > target {
					break
				}
				i += 1
			}
			pending = pending[i:]
		}
	}
}

func (f *EventTracker) fastFetch(ctx bCtx.Ctx) error {
	startBlk := f.trackerState.LastBlockProcessed
	endBlk, err := f.currentBlockGetter.BlockNumber(ctx)
	if err != nil {
		return err
	}
	endBlk = endBlk - f.followDistance
	ctx.Info(fmt.Sprintf("fast fetch %s start=%d end=%d", f.contractAddress.String(), startBlk, endBlk))
	for startBlk+CaughtUpBlock < endBlk {
		blkRange := newBlockRange(startBlk, endBlk)
		err = f.processBlkRange(ctx, blkRange)
		if err != nil {
			return err
		}
		startBlk = endBlk + 1
		endBlk, err = f.currentBlockGetter.BlockNumber(ctx)
		if err != nil {
			return err
		}
		endBlk = endBlk - f.followDistance
	}
	return nil
}

func (f *EventTracker) setupTrackerState(ctx bCtx.Ctx) (*domain.TrackerState, error) {
	addr := ToLowerHexStr(f.contractAddress)
	id := &domain.TrackerStateId{
		ChainId:         domain.ChainId(f.chainId),
		ContractAddress: domain.Address(addr),
		Tag:             f.trackerTag,
	}
	state, err := f.trackerStateUseCase.Get(ctx, id)
	if err == nil {
		// check version
		if state.Version != Version {
			// migrate logic
			if state.Version != Version-1 {
				return nil, fmt.Errorf("cannot migrate tracker state from %d to %d", state.Version, Version)
			}
			state.Version = Version
			state.LastBlockProcessed = state.LastBlockProcessed + 1
			state.LastLogIndexProcessed = -1
			if err := f.trackerStateUseCase.Update(ctx, state); err != nil {
				return nil, err
			}
		}
		return state, nil
	}
	if errors.Is(err, domain.ErrNotFound) {
		deployedBlk, err := getDeployedBlock(ctx, f.clientWithArchive, f.contractAddress)
		if err != nil {
			ctx.Logger.WithFields(map[string]interface{}{
				"chainId":  f.chainId,
				"contract": f.contractAddress,
				"tag":      f.trackerTag,
				"error":    err,
			}).Error("failed to get deployed block")
			return nil, err
		}
		ctx.WithFields(log.Fields{
			"chainId":       f.chainId,
			"contract":      f.contractAddress,
			"tag":           f.trackerTag,
			"deployedBlock": deployedBlk,
		}).Info("got deployedBlock")
		state := &domain.TrackerState{
			ChainId:               domain.ChainId(f.chainId),
			ContractAddress:       domain.Address(addr),
			Tag:                   f.trackerTag,
			Version:               Version,
			LastBlockProcessed:    deployedBlk,
			LastLogIndexProcessed: -1,
		}
		err = f.trackerStateUseCase.Store(ctx, state)
		if err != nil {
			ctx.Logger.WithFields(map[string]interface{}{
				"chainId":  f.chainId,
				"contract": f.contractAddress,
				"tag":      f.trackerTag,
				"error":    err,
			}).Error("failed to store tracker state")
			return nil, err
		}
		return state, nil
	}
	// repo error
	return nil, err
}

func (f *EventTracker) processBlkRange(ctx bCtx.Ctx, blkRange *blockRange) error {
	ranges := []*blockRange{blkRange}
	for len(ranges) > 0 {
		idx := len(ranges) - 1
		r := ranges[idx]
		ranges = ranges[:idx]
		f.filter.FromBlock = r.begin
		f.filter.ToBlock = r.end
		tCtx, cancel := bCtx.WithTimeout(ctx, TooManyLogsTimeout)
		logs, err := f.rpcClient.FilterLogs(tCtx, f.filter)
		cancel()
		if err != nil {
			// TODO: check error code (though only infura has it...)
			if r.begin.Cmp(r.end) == 0 {
				ctx.WithFields(log.Fields{
					"err":      err,
					"begin":    r.begin.String(),
					"end":      r.end.String(),
					"chainId":  f.chainId,
					"contract": f.contractAddress,
				}).Error("failed to get logs within one block")
				return err
			}
			r1, r2 := r.split()
			ranges = append(ranges, r2, r1)
			ctx.Logger.WithFields(map[string]interface{}{
				"chainId":       f.chainId,
				"contract":      f.contractAddress,
				"tag":           f.trackerTag,
				"originalRange": r.String(),
				"range1":        r1.String(),
				"range2":        r2.String(),
			}).Info("splitting blockRange")
			continue
		}
		ctx.Logger.WithFields(map[string]interface{}{
			"chainId":    f.chainId,
			"contract":   f.contractAddress,
			"tag":        f.trackerTag,
			"beginBlock": r.begin.String(),
			"endBlock":   r.end.String(),
			"#logs":      len(logs),
		}).Info(fmt.Sprintf("recieved #%d logs", len(logs)))

		// skip processed logs
		nonProcessedIndex := 0
		for _, log := range logs {
			if log.BlockNumber > f.trackerState.LastBlockProcessed {
				break
			}

			if log.BlockNumber == f.trackerState.LastBlockProcessed {
				if int64(log.Index) > f.trackerState.LastLogIndexProcessed {
					break
				}
			}
			nonProcessedIndex += 1
		}
		logs = logs[nonProcessedIndex:]

		logsWithBlockTime, err := f.toLogsWithBlockTime(ctx, logs)
		if err != nil {
			ctx.WithField("err", err).Error("f.toLogsWithBlockTime failed")
			return xerrors.Errorf("failed to inject block time: %+w", err)
		}

		batchSize := 5
		numLogs := len(logsWithBlockTime)
		i := 0
		for i < numLogs {
			j := i + batchSize
			if j > numLogs {
				j = numLogs
			}

			batchLogs := logsWithBlockTime[i:j]
			i = j

			n := len(batchLogs)
			end := batchLogs[n-1].BlockNumber
			logIndex := int64(batchLogs[n-1].Index)

			if err := f.processEvents(ctx, batchLogs, end, logIndex); err != nil {
				ctx.WithField("err", err).Error("f.processEvents failed")
				return err
			}
		}

		// update end and logIndex to end+1 and -1 of this block range
		if err := f.processEvents(ctx, nil, r.end.Uint64()+1, -1); err != nil {
			ctx.WithField("err", err).Error("f.processEvents failed")
			return err
		}
	}
	return nil
}

func (f *EventTracker) processEvents(ctx bCtx.Ctx, logsWithBlockTime []logWithBlockTime, end uint64, logIndex int64) error {
	run := func(c bCtx.Ctx) error {
		err := f.eventHandler.ProcessEvents(c, logsWithBlockTime)
		if err != nil {
			return xerrors.Errorf("failed to process events: %+w", err)
		}
		f.trackerState.LastBlockProcessed = end
		f.trackerState.LastLogIndexProcessed = logIndex
		if !f.skipMissingBlock {
			err = f.trackerStateUseCase.Update(c, f.trackerState)
			if err != nil {
				return xerrors.Errorf("failed to store tracker state: %w", err)
			}
		}
		return nil
	}

	return f.q.RunWithTransaction(ctx, run)
}

func (f *EventTracker) toLogsWithBlockTime(ctx bCtx.Ctx, logs []types.Log) ([]logWithBlockTime, error) {
	var (
		lastBlk  uint64
		lastTime time.Time
	)
	logsWithTime := make([]logWithBlockTime, len(logs))
	for idx, l := range logs {
		msgSender := domain.Address("")
		if f.shouldDecodeSender {
			tx, _, err := f.rpcClient.TransactionByHash(ctx, l.TxHash)
			if err != nil {
				ctx.WithFields(log.Fields{
					"err":      err,
					"chainId":  f.chainId,
					"contract": f.contractAddress,
					"txHash":   l.TxHash.Hex(),
				}).Error("TransactionByHash failed")
				return nil, err
			}
			_msgSender, err := types.Sender(f.signer, tx)
			if err != nil {
				ctx.WithFields(log.Fields{
					"err":      err,
					"chainId":  f.chainId,
					"contract": f.contractAddress,
					"txHash":   l.TxHash.Hex(),
				}).Error("types.Sender failed")
				return nil, err
			}
			msgSender = toDomainAddress(_msgSender)
		}
		if lastBlk != l.BlockNumber {
			blkTime, err := f.getBlockTime(ctx, l.BlockNumber)
			if err != nil {
				ctx.WithField("err", err).Error("failed to get blocktime")
				return nil, err
			}
			lastBlk = l.BlockNumber
			lastTime = *blkTime
		}
		logsWithTime[idx] = logWithBlockTime{Log: l, blockTime: lastTime, msgSender: msgSender}
	}
	return logsWithTime, nil
}

func (f *EventTracker) getBlockTime(ctx bCtx.Ctx, number uint64) (*time.Time, error) {
	blk, err := f.blockUseCase.FindOne(
		ctx,
		&chain.BlockId{
			ChainId: domain.ChainId(f.chainId),
			Number:  domain.BlockNumber(number),
		},
	)
	if err == nil {
		return &blk.Time, nil
	}

	// not found in db, get from chain
	retryCount := 20
	h, err := f.headerByNumberWithRetry(ctx, number, retryCount, time.Second)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":        err,
			"number":     number,
			"chainId":    f.chainId,
			"contract":   f.contractAddress,
			"retryCount": retryCount,
		}).Error("failed to get header")
		return nil, err
	}

	t := time.Unix(int64(h.Time), 0)
	err = f.blockUseCase.Upsert(ctx, &chain.Block{
		ChainId: domain.ChainId(f.chainId),
		Number:  domain.BlockNumber(number),
		Hash:    domain.BlockHash(ToLowerHexStr(h.Hash())),
		Time:    t,
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (f *EventTracker) headerByNumberWithRetry(ctx bCtx.Ctx, number uint64, retryLimit int, interval time.Duration) (*types.Header, error) {
	var (
		err error
		h   *types.Header
	)
	blk := new(big.Int).SetUint64(number)
	for i := 0; i < retryLimit; i++ {
		if i > 0 {
			ctx.WithFields(log.Fields{
				"chainId":  f.chainId,
				"contract": f.contractAddress,
				"retry":    i,
				"interval": interval,
				"blk":      blk,
			}).Warn("rpcClient.HeaderByNumber failed, retry")
			select {
			case <-ctx.Done():
				ctx.WithFields(log.Fields{
					"chainId":  f.chainId,
					"contract": f.contractAddress,
					"retry":    i,
					"interval": interval,
					"blk":      blk,
				}).Error("headerByNumberWithRetry: context canceled")
				return nil, xerrors.New("context canceled")
			case <-time.After(interval):
			}
			interval *= 2
		}
		h, err = f.rpcClient.HeaderByNumber(ctx, blk)
		if err == nil {
			break
		}
	}
	return h, err
}

func getDeployedBlock(ctx bCtx.Ctx, c domain.EthClientRepo, addr common.Address) (uint64, error) {
	blk, err := c.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}
	l := blk
	s := blk
	for l > 0 {
		step := l / 2
		mid := s - step - 1
		b, err := c.CodeAt(ctx, addr, new(big.Int).SetUint64(mid))
		if err != nil {
			return 0, err
		}
		if len(b) > 0 {
			s = mid
			l -= step + 1
		} else {
			l = step
		}
	}
	return s, nil
}
