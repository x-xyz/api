package repository

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type openseaRepo struct {
	q query.Mongo
}

func NewOpenseaDataRepo(q query.Mongo) domain.OpenseaDataRepo {
	return &openseaRepo{q: q}
}

func (r *openseaRepo) FindAll(ctx bCtx.Ctx, optsFns ...domain.OpenseaDataFindAllOptions) ([]domain.OpenseaData, error) {
	opts, err := domain.GetOpenseaDataFindAllOptions(optsFns...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"optsFns": optsFns,
			"err":     err,
		}).Error("GetOpenseaDataFindAllOptions failed")
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

	if opts.Addresses != nil {
		query["address"] = bson.M{
			"$in": *opts.Addresses,
		}
	}

	res := []domain.OpenseaData{}
	if err := r.q.Search(ctx, domain.TableOpenseaData, offset, limit, sort, query, &res); err != nil {
		ctx.WithField("err", err).Error("q.Search failed")
		return nil, err
	}
	return res, nil
}

func (r *openseaRepo) FindOne(ctx bCtx.Ctx, id domain.OpenseaDataId) (*domain.OpenseaData, error) {
	v := &domain.OpenseaData{}
	err := r.q.FindOne(ctx, domain.TableOpenseaData, id, v)
	if err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		ctx.WithFields(log.Fields{
			"id":  id,
			"err": err,
		}).Error("q.FindOne failed")
		return nil, err
	}
	return v, nil
}

func (r *openseaRepo) Upsert(ctx bCtx.Ctx, v domain.OpenseaData) error {
	if err := r.q.Upsert(ctx, domain.TableOpenseaData, v.ToId(), v); err != nil {
		ctx.WithFields(log.Fields{
			"openseaData": v,
			"err":         err,
		}).Error("q.Upsert failed")
		return err
	}
	return nil
}
