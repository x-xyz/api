package nft_indexer

import (
	"sync"
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/token"
)

type MetadataUpdaterCfg struct {
	TokenUC      token.Usecase
	ChainId      domain.ChainId
	TargetStates []nftitem.IndexerState
	Collections  []domain.Address
	RetryLimit   int
	Batch        int
	Workers      int
	Interval     time.Duration
	ErrorCh      chan<- error
}

type MetadataUpdater struct {
	tokenUC      token.Usecase
	chainId      domain.ChainId
	targetStates []nftitem.IndexerState
	collections  []domain.Address
	retryLimit   int
	batch        int
	workers      int
	interval     time.Duration
	taskCh       chan *nftitem.NftItem
	errorCh      chan<- error
	stoppedCh    chan interface{}
}

func NewMetadataUpdater(cfg *MetadataUpdaterCfg) *MetadataUpdater {
	return &MetadataUpdater{
		tokenUC:      cfg.TokenUC,
		chainId:      cfg.ChainId,
		targetStates: cfg.TargetStates,
		retryLimit:   cfg.RetryLimit,
		collections:  cfg.Collections,
		batch:        cfg.Batch,
		workers:      cfg.Workers,
		interval:     cfg.Interval,
		taskCh:       make(chan *nftitem.NftItem, cfg.Batch),
		stoppedCh:    make(chan interface{}),
	}
}

func (i *MetadataUpdater) Start(ctx bCtx.Ctx) {
	go i.loop(ctx)
}

func (i *MetadataUpdater) Wait() {
	<-i.stoppedCh
}

func (i *MetadataUpdater) loop(ctx bCtx.Ctx) {
	workerCtx, cancel := bCtx.WithCancel(ctx)
	workerWg := sync.WaitGroup{}
	resCh := make(chan error, i.workers)

	errAndStop := func(err error) {
		i.errorCh <- err
		cancel()
		workerWg.Wait()
		close(i.stoppedCh)
	}

	for j := 0; j < i.workers; j++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for {
				select {
				case <-workerCtx.Done():
					return
				case task := <-i.taskCh:
					err := i.processNft(workerCtx, task)
					if err != nil {
						resCh <- err
						return
					}
					resCh <- nil
				}
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			workerWg.Wait()
			close(i.stoppedCh)
			return
		case <-time.After(i.interval):
			for _, c := range i.collections {
				ctx.WithFields(log.Fields{
					"chainId":    i.chainId,
					"collection": c,
				}).Info("refreshing metadata")
				for {
					items, _, err := i.tokenUC.SearchForIndexerState(ctx, i.targetStates, i.retryLimit,
						token.WithChainId(i.chainId),
						token.WithCollections(c),
						token.WithPagination(0, int32(i.batch)),
					)
					if err != nil {
						errAndStop(err)
						return
					}
					ctx.WithFields(log.Fields{
						"collection": c,
						"#items":     len(items),
						"states":     i.targetStates,
					}).Info("search result")
					for _, item := range items {
						i.taskCh <- item
					}
					for j := 0; j < len(items); j++ {
						select {
						case <-ctx.Done():
							cancel()
							workerWg.Wait()
							close(i.stoppedCh)
							return
						case err := <-resCh:
							if err != nil {
								errAndStop(err)
								return
							}
						}
					}
					if len(items) < i.batch {
						break
					}
				}
				for {
					pendingItems, _, err := i.tokenUC.SearchForIndexerState(ctx, []nftitem.IndexerState{nftitem.IndexerStatePendingTokenURIRefreshing}, i.retryLimit,
						token.WithChainId(i.chainId),
						token.WithCollections(c),
						token.WithPagination(0, int32(i.batch)),
					)
					if err != nil {
						errAndStop(err)
						return
					}
					ctx.WithFields(log.Fields{
						"collection":    c,
						"#pendingItems": len(pendingItems),
						"state":         nftitem.IndexerStatePendingTokenURIRefreshing,
					}).Info("search result")
					for _, item := range pendingItems {
						i.taskCh <- item
					}
					for j := 0; j < len(pendingItems); j++ {
						select {
						case <-ctx.Done():
							cancel()
							workerWg.Wait()
							close(i.stoppedCh)
							return
						case err := <-resCh:
							if err != nil {
								errAndStop(err)
								return
							}
						}
					}
					if len(pendingItems) < i.batch {
						break
					}
				}
			}
		}
	}
}

func (i *MetadataUpdater) processNft(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	ctx = bCtx.WithValues(ctx, map[string]interface{}{
		"chainId":  item.ChainId,
		"contract": item.ContractAddress,
		"tokenId":  item.TokenId,
	})

	switch item.IndexerState {
	case nftitem.IndexerStateHasTokenURI:
		fallthrough
	case nftitem.IndexerStateHasTokenURIRefreshing:
		fallthrough
	case nftitem.IndexerStateHasImageURL:
		fallthrough
	case nftitem.IndexerStateHasHostedImage:
		fallthrough
	case nftitem.IndexerStateParsingAttributes:
		fallthrough
	case nftitem.IndexerStateFetchingAnimation:
		fallthrough
	case nftitem.IndexerStateDone:
		if err := i.PatchNft(ctx, item, nftitem.IndexerStatePendingTokenURIRefreshing); err != nil {
			ctx.WithField("err", err).Error("token.PatchNft failed")
			return err
		}
	case nftitem.IndexerStatePendingTokenURIRefreshing:
		if err := i.PatchNft(ctx, item, nftitem.IndexerStateHasTokenURIRefreshing); err != nil {
			ctx.WithField("err", err).Error("token.PatchNft failed")
			return err
		}
	default:
		return nil
	}
	return nil
}

func (i *MetadataUpdater) PatchNft(ctx bCtx.Ctx, item *nftitem.NftItem, state nftitem.IndexerState) error {
	patchable := &nftitem.PatchableNftItem{
		IndexerState:      (*nftitem.IndexerState)(ptr.String(string(state))),
		IndexerRetryCount: ptr.Int32(0),
	}
	err := i.tokenUC.PatchNft(ctx, item.ToId(), patchable)
	if err != nil {
		ctx.WithField("err", err).Error("nftitem.Patch failed")
		return err
	}
	return nil
}
