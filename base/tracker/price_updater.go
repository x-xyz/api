package tracker

import (
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	pricefomatter "github.com/x-xyz/goapi/base/price_fomatter"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/domain/token"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PriceUpdaterCfg struct {
	ChainId        domain.ChainId
	Collection     collection.Usecase
	Token          token.Usecase
	Order          order.UseCase
	Interval       time.Duration
	PriceFormatter pricefomatter.PriceFormatter
	ErrorCh        chan<- error
}

type PriceUpdater struct {
	chainId        domain.ChainId
	collection     collection.Usecase
	token          token.Usecase
	order          order.UseCase
	interval       time.Duration
	priceFormatter pricefomatter.PriceFormatter
	errorCh        chan<- error
	stoppedCh      chan interface{}
}

func NewPriceUpdater(cfg *PriceUpdaterCfg) *PriceUpdater {
	return &PriceUpdater{
		chainId:        cfg.ChainId,
		collection:     cfg.Collection,
		token:          cfg.Token,
		order:          cfg.Order,
		interval:       cfg.Interval,
		priceFormatter: cfg.PriceFormatter,
		errorCh:        cfg.ErrorCh,
		stoppedCh:      make(chan interface{}),
	}
}

func (u *PriceUpdater) Start(ctx bCtx.Ctx) {
	go u.loop(ctx)
}

func (u *PriceUpdater) Wait() {
	<-u.stoppedCh
}

func (u *PriceUpdater) loop(ctx bCtx.Ctx) {
	errAndStop := func(err error) {
		u.errorCh <- err
		close(u.stoppedCh)
	}

	nextTick := time.Second * 0
	offset := int32(0)
	limit := int32(500)

	for {
		select {
		case <-ctx.Done():
			close(u.stoppedCh)
			return
		case <-time.After(nextTick):
			ctx.Info("updating prices")
			cols, err := u.collection.FindAll(ctx, collection.WithPagination(offset, limit), collection.WithChainId(u.chainId))
			if err != nil {
				ctx.WithFields(log.Fields{
					"chainId": u.chainId,
					"offset":  offset,
					"limit":   limit,
				}).Error("collection.FindAll failed")
				errAndStop(err)
				return
			}
			for _, col := range cols.Items {
				id := collection.CollectionId{ChainId: col.ChainId, Address: col.Erc721Address}
				ctx.WithField("collection", id).Info("updating price")
				if err := u.updatePricesForCollection(ctx, &col.Collection); err != nil {
					ctx.WithFields(log.Fields{
						"id":  id,
						"err": err,
					}).Error("updatePricesForCollection failed")
					errAndStop(err)
					return
				}
			}
			if len(cols.Items) < int(limit) {
				nextTick = u.interval
				offset = 0
			} else {
				nextTick = time.Second * 0
				offset += limit
			}
		}
	}
}

func (u *PriceUpdater) updatePricesForCollection(ctx bCtx.Ctx, col *collection.Collection) error {
	id := collection.CollectionId{ChainId: col.ChainId, Address: col.Erc721Address}
	if err := u.updatePrices(ctx, id); err != nil {
		ctx.WithFields(log.Fields{
			"id":  id,
			"err": err,
		}).Error("updatePricesForListings failed")
	}
	return nil

	// TODO: add auction, bids
}

func (u *PriceUpdater) updatePriceForToken(ctx bCtx.Ctx, item *token.TokenWithDetail) error {
	ctx.WithFields(log.Fields{
		"id": *item.ToId(),
	}).Info("update price")
	if err := u.order.RefreshOrders(ctx, *item.ToId()); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  *item.ToId(),
		}).Error("failed to order.RefreshOrders")
		return err
	}

	if err := u.token.RefreshListingAndOfferState(ctx, *item.ToId()); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  *item.ToId(),
		}).Error("failed to token.RefreshListingAndOfferState")
		return err
	}
	return nil
}

func (u *PriceUpdater) updatePrices(ctx bCtx.Ctx, id collection.CollectionId) error {
	limit := int32(500)
	lastObjectId := primitive.NewObjectID()
	for {
		items, err := u.token.SearchV2(ctx,
			token.WithObjectIdLT(lastObjectId),
			token.WithChainId(u.chainId),
			token.WithCollections(id.Address),
			token.WithPagination(0, limit),
			token.WithHasOrder(true),
		)
		if err != nil {
			ctx.WithFields(log.Fields{
				"id":           id,
				"lastObjectId": lastObjectId.String(),
				"limit":        limit,
			}).Error("token.SearchV2 failed")
			return err
		}
		for _, item := range items.Items {
			if err := u.updatePriceForToken(ctx, item); err != nil {
				ctx.WithFields(log.Fields{
					"id":  item.ToId(),
					"err": err,
				}).Error("updatePriceForToken failed")
				return err
			}
		}
		if len(items.Items) < int(limit) {
			return nil
		}
		lastObjectId = items.Items[len(items.Items)-1].ObjectId
	}
}
