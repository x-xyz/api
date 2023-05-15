package nft_indexer

import (
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain/account"
)

type FolderStatUpdaterCfg struct {
	Folder   account.FolderUseCase
	Interval time.Duration
	ErrorCh  chan<- error
}

type FolderStatUpdater struct {
	folder    account.FolderUseCase
	interval  time.Duration
	errorCh   chan<- error
	stoppedCh chan interface{}
}

func NewFolderStatUpdater(cfg *FolderStatUpdaterCfg) *FolderStatUpdater {
	return &FolderStatUpdater{
		folder:    cfg.Folder,
		interval:  cfg.Interval,
		errorCh:   cfg.ErrorCh,
		stoppedCh: make(chan interface{}),
	}
}

func (i *FolderStatUpdater) Start(ctx bCtx.Ctx) {
	go i.loop(ctx)
}

func (i *FolderStatUpdater) Wait() {
	<-i.stoppedCh
}

func (i *FolderStatUpdater) loop(ctx bCtx.Ctx) {
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
			ctx.WithField("offset", offset).Info("updating folder stats")
			folders, err := i.folder.GetFolders(ctx, account.WithPagination(offset, limit))
			if err != nil {
				ctx.WithFields(log.Fields{
					"offset": offset,
					"limit":  limit,
					"err":    err,
				}).Error("folder.GetFolders failed")
				errAndStop(err)
				return
			}
			for _, folder := range folders {

				if err := i.folder.RefreshStat(ctx, folder.Id); err != nil {
					ctx.WithFields(log.Fields{
						"folderId": folder.Id,
						"err":      err,
					}).Error("folder.RefreshStat failed")
					errAndStop(err)
					return
				}
			}
			if len(folders) < int(limit) {
				nextTick = i.interval
				offset = 0
			} else {
				nextTick = time.Second * 0
				offset += limit
			}
		}
	}
}
