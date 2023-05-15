package nft_indexer

import (
	"time"

	"github.com/x-xyz/goapi/base/backoff"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/service/opensea"
)

const ethereumChainId = domain.ChainId(1)

type OpenseaDataIndexerCfg struct {
	Collection    collection.Usecase
	OpenseaData   domain.OpenseaDataUseCase
	OpenseaClient opensea.Client
	RetryLimit    int
	Backoff       *backoff.Backoff
	Interval      time.Duration
	ErrorCh       chan<- error
}

type OpenseaDataIndexer struct {
	collection    collection.Usecase
	openseaData   domain.OpenseaDataUseCase
	openseaClient opensea.Client
	retryLimit    int
	backoff       *backoff.Backoff
	interval      time.Duration
	errorCh       chan<- error
	stoppedCh     chan interface{}
}

func NewOpenseaDataIndexer(cfg *OpenseaDataIndexerCfg) *OpenseaDataIndexer {
	return &OpenseaDataIndexer{
		collection:    cfg.Collection,
		openseaData:   cfg.OpenseaData,
		openseaClient: cfg.OpenseaClient,
		retryLimit:    cfg.RetryLimit,
		backoff:       cfg.Backoff,
		interval:      cfg.Interval,
		errorCh:       cfg.ErrorCh,
		stoppedCh:     make(chan interface{}),
	}
}

func (i *OpenseaDataIndexer) Start(ctx bCtx.Ctx) {
	go i.loop(ctx)
}

func (i *OpenseaDataIndexer) Wait() {
	<-i.stoppedCh
}

func (i *OpenseaDataIndexer) loop(ctx bCtx.Ctx) {
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

func (i *OpenseaDataIndexer) index(ctx bCtx.Ctx, col *collection.Collection) error {
	id := domain.OpenseaDataId{ChainId: ethereumChainId, Address: col.Erc721Address}
	ctx.WithField("id", id).Info("fetching from opensea")
	entry, err := i.getEntry(ctx, id)
	if err != nil {
		return err
	}

	var (
		retries = 0
		data    *opensea.CollectionResp
	)
	i.backoff.Reset()
	for retries < i.retryLimit {
		data, err = i.openseaClient.GetCollectionBySlug(ctx, entry.Slug)
		if err == nil {
			break
		}
		// wait 1 second after calling opensea api to prevent rate limit
		time.Sleep(1 * time.Second)
		retries++
		if i.backoff.Backoff(ctx) != nil {
			// ctx closed
			break
		}
	}
	if err != nil {
		ctx.WithFields(log.Fields{
			"slug":    entry.Slug,
			"retries": retries,
			"err":     err,
		}).Error("openseaClient.GetCollectionBySlug failed")
		return nil
	}
	entry.OneHourVolume = data.Collection.Stats.OneHourVolume
	entry.OneHourChange = data.Collection.Stats.OneHourChange
	entry.OneHourSales = data.Collection.Stats.OneHourSales
	entry.SixHourVolume = data.Collection.Stats.SixHourVolume
	entry.SixHourChange = data.Collection.Stats.SixHourChange
	entry.SixHourSales = data.Collection.Stats.SixHourSales
	entry.OneDayVolume = data.Collection.Stats.OneDayVolume
	entry.OneDayChange = data.Collection.Stats.OneDayChange
	entry.OneDaySales = data.Collection.Stats.OneDaySales
	entry.SevenDayVolume = data.Collection.Stats.SevenDayVolume
	entry.SevenDayChange = data.Collection.Stats.SevenDayChange
	entry.SevenDaySales = data.Collection.Stats.SevenDaySales
	entry.ThirtyDayVolume = data.Collection.Stats.ThirtyDayVolume
	entry.ThirtyDayChange = data.Collection.Stats.ThirtyDayChange
	entry.ThirtyDaySales = data.Collection.Stats.ThirtyDaySales
	entry.TotalVolume = data.Collection.Stats.TotalVolume
	entry.TotalSales = data.Collection.Stats.TotalSales
	entry.FloorPrice = data.Collection.Stats.FloorPrice
	if err := i.openseaData.Upsert(ctx, *entry); err != nil {
		ctx.WithFields(log.Fields{
			"entry": entry,
			"err":   err,
		}).Error("openseaData.Upsert failed")
		return err
	}

	if err := i.collection.UpdateOpenseaFloorPrice(ctx, col.ToId(), data.Collection.Stats.FloorPrice); err != nil {
		ctx.WithFields(log.Fields{
			"entry": entry,
			"err":   err,
		}).Error("UpdateOpenseaFloorPrice failed")
		return err
	}
	return nil
}

func (i *OpenseaDataIndexer) getEntry(ctx bCtx.Ctx, id domain.OpenseaDataId) (*domain.OpenseaData, error) {
	entry, err := i.openseaData.FindOne(ctx, id)
	if err == nil {
		return entry, nil
	} else if err != domain.ErrNotFound {
		ctx.WithFields(log.Fields{
			"id":  id,
			"err": err,
		}).Error("openseaData.FindOne failed")
		return nil, err
	}

	var (
		retries = 0
		resp    *opensea.AssetContractResp
	)
	i.backoff.Reset()
	for retries < i.retryLimit {
		resp, err = i.openseaClient.GetAssetContractByAddress(ctx, id.Address.ToLowerStr())
		if err == nil {
			break
		}
		retries++
		if i.backoff.Backoff(ctx) != nil {
			// ctx closed
			break
		}
	}
	if err != nil {
		ctx.WithFields(log.Fields{
			"address": id.Address,
			"retries": retries,
			"err":     err,
		}).Error("openseaClient.GetAssetContractByAddress failed")
		return nil, err
	}
	return &domain.OpenseaData{
		ChainId:           id.ChainId,
		Address:           id.Address,
		Slug:              resp.Collection.Slug,
		Name:              resp.Collection.Name,
		Description:       resp.Collection.Description,
		ImageUrl:          resp.Collection.Image,
		ExternalUrl:       resp.Collection.Url,
		DiscordUrl:        resp.Collection.Discord,
		TwitterUserName:   resp.Collection.Twitter,
		InstagramUserName: resp.Collection.Instagram,
		MediumUserName:    resp.Collection.Medium,
		TelegramUrl:       resp.Collection.Telegram,
		PayoutAddress:     domain.Address(resp.Collection.PayoutAddress).ToLower(),
		Royalty:           resp.Royalty,
		Symbol:            resp.Symbol,
	}, nil
}
