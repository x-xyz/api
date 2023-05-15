package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

func makeFindQuery(optFns ...account.FindActivityHistoryOptions) (bson.M, error) {
	opts, err := account.GetFindActivityHistoryOptions(optFns...)
	if err != nil {
		return nil, err
	}

	qry := bson.M{}

	if opts.Account != nil {
		qry["$or"] = bson.A{
			bson.M{"account": *opts.Account},
			bson.M{"to": opts.Account},
		}
	}

	if opts.ChainId != nil {
		qry["chainId"] = *opts.ChainId
	}

	if opts.Contract != nil {
		qry["contractAddress"] = *opts.Contract
	}

	if opts.TokenId != nil {
		qry["tokenId"] = *opts.TokenId
	}

	if opts.TimeGTE != nil {
		qry["time"] = bson.M{"$gte": *opts.TimeGTE}
	}

	if len(opts.Types) > 1 {
		qry["type"] = bson.M{"$in": opts.Types}
	} else if len(opts.Types) > 0 {
		qry["type"] = opts.Types[0]
	}

	if opts.Source != nil {
		qry["source"] = *opts.Source
	}

	return qry, nil
}

type activityHistoryRepo struct {
	q query.Mongo
}

func NewActivityHistoryRepo(q query.Mongo) account.ActivityHistoryRepo {
	return &activityHistoryRepo{q: q}
}

func (r *activityHistoryRepo) Insert(ctx bCtx.Ctx, a *account.ActivityHistory) error {
	if err := r.q.Insert(ctx, domain.TableActivityHistories, a); err != nil {
		ctx.WithFields(log.Fields{
			"activityHistory": a,
			"err":             err,
		}).Error("q.Insert failed")
		return err
	}
	return nil
}

func (r *activityHistoryRepo) FindActivities(c ctx.Ctx, optFns ...account.FindActivityHistoryOptions) ([]account.ActivityHistory, error) {
	opts, err := account.GetFindActivityHistoryOptions(optFns...)
	if err != nil {
		c.WithField("err", err).Error("account.GetFindActivityHistoryOptions failed")
		return nil, err
	}

	qry, err := makeFindQuery(optFns...)
	if err != nil {
		c.WithField("err", err).Error("makeFindQuery failed")
		return nil, err
	}

	offset := 0
	limit := 0

	if opts.Offset != nil {
		offset = *opts.Offset
	}

	if opts.Limit != nil {
		limit = *opts.Limit
	}

	res := []account.ActivityHistory{}

	err = r.q.Search(c, domain.TableActivityHistories, offset, limit, "-time", qry, &res)

	if err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		c.WithField("err", err).WithField("query", qry).Error("q.Search failed")
		return nil, err
	}

	return res, nil
}

func (r *activityHistoryRepo) CountActivities(c ctx.Ctx, optFns ...account.FindActivityHistoryOptions) (int, error) {
	qry, err := makeFindQuery(optFns...)
	if err != nil {
		c.WithField("err", err).Error("makeFindQuery failed")
		return 0, err
	}

	cnt, err := r.q.Count(c, domain.TableActivityHistories, qry)

	if err != nil {
		c.WithField("err", err).WithField("query", qry).Error("q.Count failed")
		return 0, err
	}

	return cnt, nil
}

func (r *activityHistoryRepo) UpsertBySourceEventId(ctx ctx.Ctx, source account.SourceType, sourceEventId string, t account.ActivityHistoryType, ah *account.ActivityHistory) error {
	bsonM, err := mongoclient.MakeBsonM(ah)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":              err,
			"acitivityHistory": *ah,
		}).Error("failed to mongoclient.MakeBsonM")
		return err
	}

	selector := bson.M{
		"source":        source,
		"sourceEventId": sourceEventId,
		"type":          t,
	}

	err = r.q.Upsert(ctx, domain.TableActivityHistories, selector, bsonM)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"selector": selector,
			"updater":  bsonM,
		}).Error("failed to Upsert")
		return err
	}

	return nil
}

func (r *activityHistoryRepo) InsertTransferActivityIfNotExists(ctx ctx.Ctx, ah *account.ActivityHistory) error {
	id := account.TransferID{
		ChainId:         ah.ChainId,
		ContractAddress: ah.ContractAddress,
		Type:            ah.Type,
		TxHash:          ah.TxHash,
		LogIndex:        ah.LogIndex,
	}

	h := &account.ActivityHistory{}

	err := r.q.FindOne(ctx, domain.TableActivityHistories, id, h)
	if err == query.ErrNotFound {
		return r.q.Insert(ctx, domain.TableActivityHistories, ah)
	}
	return err
}
