package nft_indexer

import (
	"fmt"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain/collection"
)

type StatUpdater struct {
	collection collection.Usecase
	// minInterval between every tasks emission
	interval time.Duration
	// concurrent worker count
	concurrency int
	taskCh      chan *collection.CollectionId
	errorCh     chan error
	stoppedCh   chan interface{}
}

func NewStatUpdater(collectionUsacase collection.Usecase, errCh chan error) *StatUpdater {
	return &StatUpdater{
		collection: collectionUsacase,
		errorCh:    errCh,
		stoppedCh:  make(chan interface{}),
	}
}

func (im *StatUpdater) SetInterval(interval time.Duration) *StatUpdater {
	im.interval = interval
	return im
}

func (i *StatUpdater) Start(ctx ctx.Ctx) {
	go i.loop(ctx)
}

func (im *StatUpdater) loop(ctx ctx.Ctx) {
	errAndStop := func(err error) {
		im.errorCh <- err
		close(im.stoppedCh)
	}

	nextTick := time.Second * 0
	limit := int32(100)
	offset := int32(0)

	for {
		select {
		case <-ctx.Done():
			close(im.stoppedCh)
			return
		case <-time.After(nextTick):
			cols, err := im.collection.FindAll(ctx, collection.WithPagination(offset, limit))
			if err != nil {
				ctx.WithFields(log.Fields{
					"offset": offset,
					"limit":  limit,
					"err":    err,
				}).Error("im.collection.FindAll failed")
				errAndStop(err)
				return
			}

			ctx.Info(fmt.Sprintf("stat update progress: %d", offset))

			for _, col := range cols.Items {
				id := collection.CollectionId{
					ChainId: col.ChainId,
					Address: col.Erc721Address,
				}
				if err := im.collection.RefreshStat(ctx, id); err != nil {
					ctx.WithFields(log.Fields{
						"chainId": col.ChainId,
						"address": col.Erc721Address,
						"err":     err,
					}).Error("im.collection.RefreshStat failed")
					errAndStop(err)
					return
				}
			}
			if len(cols.Items) < int(limit) {
				nextTick = im.interval
				offset = 0
			} else {
				nextTick = time.Second * 0
				offset += limit
			}
		}
	}
}

func (im *StatUpdater) Wait() {
	<-im.stoppedCh
}
