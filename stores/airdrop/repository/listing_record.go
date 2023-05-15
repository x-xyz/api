package repository

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/airdrop"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type listingRecordRepoImpl struct {
	q query.Mongo
}

func NewListingRecordRepo(q query.Mongo) airdrop.ListingRecordRepo {
	return &listingRecordRepoImpl{q}
}

func (r *listingRecordRepoImpl) Upsert(ctx bCtx.Ctx, lr *airdrop.ListingRecord) error {
	selector := lr.ToId()
	if err := r.q.Upsert(ctx, domain.TableListingRecords, selector, lr); err != nil {
		ctx.WithFields(log.Fields{
			"selector": selector,
			"lr":       lr,
			"err":      err,
		}).Error("Upsert failed")
		return err
	}
	return nil
}

func (r *listingRecordRepoImpl) FindAll(ctx bCtx.Ctx, optFns ...airdrop.ListingRecordFindAllOptionsFunc) ([]airdrop.ListingRecord, error) {
	opts, err := airdrop.GetListingRecordFindAllOptions(optFns...)
	if err != nil {
		ctx.WithField("err", err).Error("proof.GetListingRecordFindAllOptions failed")
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
		ctx.WithField("err", err).Error("MakeBsonM failed")
		return nil, err
	}

	if opts.SnapshotTimeGTE != nil || opts.SnapshotTimeLT != nil {
		subQuery := bson.M{}
		if opts.SnapshotTimeGTE != nil {
			subQuery["$gte"] = *opts.SnapshotTimeGTE
		}
		if opts.SnapshotTimeLT != nil {
			subQuery["$lt"] = *opts.SnapshotTimeLT
		}
		query["snapshotTime"] = subQuery
	}

	listingRecords := []airdrop.ListingRecord{}
	if err := r.q.Search(ctx, domain.TableListingRecords, offset, limit, sort, query, &listingRecords); err != nil {
		ctx.WithField("err", err).Error("q.Search failed")
		return nil, err
	}
	return listingRecords, err
}
