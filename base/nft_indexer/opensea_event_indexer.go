package nft_indexer

import (
	"time"

	"github.com/x-xyz/goapi/base/backoff"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/service/opensea"
)

type OpenseaEventIndexerCfg struct {
	Collection             collection.Usecase
	ActivityHistoryUsecase account.ActivityHistoryUseCase
	OpenseaClient          opensea.Client
	RetryLimit             int
	Backoff                *backoff.Backoff
	Interval               time.Duration
	ErrorCh                chan<- error
}

type OpenseaEventIndexer struct {
	collection             collection.Usecase
	activityHistoryUsecase account.ActivityHistoryUseCase
	openseaClient          opensea.Client
	retryLimit             int
	backoff                *backoff.Backoff
	interval               time.Duration
	errorCh                chan<- error
	stoppedCh              chan interface{}
}

func NewOpenseaEventIndexer(cfg *OpenseaEventIndexerCfg) *OpenseaEventIndexer {
	return &OpenseaEventIndexer{
		collection:             cfg.Collection,
		activityHistoryUsecase: cfg.ActivityHistoryUsecase,
		openseaClient:          cfg.OpenseaClient,
		retryLimit:             cfg.RetryLimit,
		backoff:                cfg.Backoff,
		interval:               cfg.Interval,
		errorCh:                cfg.ErrorCh,
		stoppedCh:              make(chan interface{}),
	}
}

func (i *OpenseaEventIndexer) Start(ctx bCtx.Ctx) {
	go i.loop(ctx)
}

func (i *OpenseaEventIndexer) Wait() {
	<-i.stoppedCh
}

func (i *OpenseaEventIndexer) loop(ctx bCtx.Ctx) {
	errAndStop := func(err error) {
		i.errorCh <- err
		close(i.stoppedCh)
	}

	nextTick := time.Second * 0
	limit := int32(100)
	offset := int32(0)

	for {
		select {
		case <-ctx.Done():
			close(i.stoppedCh)
			return
		case <-time.After(nextTick):
			collections, err := i.collection.FindAll(ctx,
				collection.WithChainId(ethereumChainId),
				collection.WithPagination(offset, limit),
			)
			if err != nil {
				ctx.WithFields(log.Fields{
					"chainId": ethereumChainId,
					"offset":  offset,
					"limit":   limit,
					"err":     err,
				}).Error("collection.FindAll failed")
				errAndStop(err)
				return
			}
			for _, col := range collections.Items {
				if err := i.index(ctx, &col.Collection); err != nil {
					ctx.WithFields(log.Fields{
						"chainId": col.ChainId,
						"address": col.Erc721Address,
						"err":     err,
					}).Error("i.index failed")
					errAndStop(err)
					return
				}
			}
			if len(collections.Items) < int(limit) {
				nextTick = i.interval
				offset = 0
			} else {
				nextTick = time.Second * 0
				offset += limit
			}
		}
	}
}

func (i *OpenseaEventIndexer) index(ctx bCtx.Ctx, col *collection.Collection) error {
	collectionId := collection.CollectionId{
		ChainId: ethereumChainId,
		Address: col.Erc721Address,
	}
	ctx.WithFields(log.Fields{
		"id":             collectionId,
		"collectionName": col.CollectionName,
	}).Info("fetching events from opensea")

	retries := 0
	cursor := ""

	startTime := col.LastOpenseaEventIndexAt
	endTime := time.Now()

	for {
		var (
			data *opensea.EventResp
			err  error
		)

		i.backoff.Reset()
		for retries < i.retryLimit {
			data, err = i.openseaClient.GetEvent(
				ctx,
				opensea.WithContractAddress(col.Erc721Address.ToLower()),
				opensea.WithEventType(opensea.EventTypeSuccessful),
				opensea.WithBefore(endTime),
				opensea.WithAfter(startTime),
				opensea.WithCursor(cursor),
			)
			// wait 1 second after calling opensea api to prevent rate limit
			time.Sleep(1 * time.Second)
			if err == nil {
				break
			}
			retries++
			if i.backoff.Backoff(ctx) != nil {
				// ctx closed
				break
			}
			ctx.WithFields(log.Fields{
				"retries":    retries,
				"collection": col.CollectionName,
			}).Info("retry openseaClient.GetEvent")
		}
		if err != nil {
			ctx.WithFields(log.Fields{
				"collection": col.CollectionName,
				"retries":    retries,
				"err":        err,
			}).Error("openseaClient.GetEvent failed")
			return err
		}

		for _, ev := range data.AssetEvents {
			err := i.activityHistoryUsecase.ParseAndInsertOpenseaEventToActivityHistory(ctx, ev)
			if err != nil {
				ctx.WithFields(log.Fields{
					"err":   err,
					"event": ev,
				}).Error("failed to ParseAndInsertOpenseaEventToActivityHistory")
				return err
			}
		}

		cursor = data.Next
		if cursor == "" {
			break
		}
	}

	err := i.collection.UpdateLastOpenseaEventIndexAt(ctx, collectionId, endTime)
	if err != nil {
		ctx.WithFields(log.Fields{
			"collection": col.CollectionName,
			"err":        err,
		}).Error("failed to collection.UpdateLastOpenseaEventIndexAt")
		return err
	}

	return nil
}
