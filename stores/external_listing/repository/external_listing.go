package repository

import (
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/external_listing"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type externalListingRepo struct {
	q query.Mongo
}

func (r *externalListingRepo) FindAll(ctx bCtx.Ctx, optsFns ...external_listing.FindAllOptionsFunc) ([]external_listing.ExternalListing, error) {
	opts, err := external_listing.GetFindAllOptions(optsFns...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"optsFns": optsFns,
			"err":     err,
		}).Error("GetOpenseaListingFindAllOptions failed")
		return nil, err
	}
	var (
		offset int    = 0
		limit  int    = 0
		sort   string = "_id"
	)
	if opts.Offset != nil {
		offset = int(*opts.Offset)
	}
	if opts.Limit != nil {
		limit = int(*opts.Limit)
	}
	if opts.SortBy != nil && opts.SortDir != nil {
		sort = *opts.SortBy
		if *opts.SortDir == domain.SortDirDesc {
			sort = "-" + sort
		}
	}
	query, err := mongoclient.MakeBsonM(opts)
	if err != nil {
		ctx.WithFields(log.Fields{
			"opts": opts,
			"err":  err,
		}).Error("MakeBsonM failed")
		return nil, err
	}
	res := []external_listing.ExternalListing{}
	if err := r.q.Search(ctx, domain.TableExternalListing, offset, limit, sort, query, &res); err != nil {
		ctx.WithField("err", err).Error("q.Search failed")
		return nil, err
	}
	return res, nil
}

func NewExternalListingRepo(q query.Mongo) *externalListingRepo {
	return &externalListingRepo{q: q}
}

func (r *externalListingRepo) BulkUpsert(ctx bCtx.Ctx, items []external_listing.ExternalListing) error {
	ops := []query.UpsertOp{}
	nowTime := time.Now()
	for _, it := range items {
		ops = append(ops, query.UpsertOp{
			Selector: bson.M{
				"owner":           it.Owner,
				"chainId":         it.ChainId,
				"contractAddress": it.ContractAddress,
				"tokenId":         it.TokenId,
			},
			Updater: bson.M{
				"owner":           it.Owner,
				"chainId":         it.ChainId,
				"minter":          it.Minter,
				"contractAddress": it.ContractAddress,
				"tokenId":         it.TokenId,
				"quantity":        it.Quantity,
				"paymentToken":    it.PaymentToken,
				"price":           it.Price,
				"priceInUSD":      it.PriceInUsd,
				"startTime":       it.StartTime,
				"deadline":        it.Deadline,
				"source":          it.Source,
				"updatedTime":     nowTime,
			},
		})
	}
	if _, _, err := r.q.BulkUpsert(ctx, domain.TableExternalListing, ops); err != nil {
		ctx.WithFields(log.Fields{
			"externalListings": items,
			"err":              err,
		}).Error("q.BulkUpsert failed")
		return err
	}
	return nil
}

func (r *externalListingRepo) RemoveAll(c bCtx.Ctx, optFns ...external_listing.RemoveAllOptionsFunc) error {
	opts, err := external_listing.GetRemoveAllOptions(optFns...)
	if err != nil {
		c.WithField("err", err).Error("external_listing.GetRemoveAllOptions failed")
		return err
	}
	sel, err := mongoclient.MakeBsonM(opts)
	if err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return err
	}
	if _, err := r.q.RemoveAll(c, domain.TableExternalListing, sel); err != nil {
		c.WithField("err", err).Error("q.RemoveAll failed")
		return err
	}
	return nil
}
