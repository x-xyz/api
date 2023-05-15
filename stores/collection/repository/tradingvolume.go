package repository

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/service/query"
)

type tradingVolumeRepo struct {
	q query.Mongo
}

func NewTradingVolumeRepo(q query.Mongo) collection.TradingVolumeRepo {
	return &tradingVolumeRepo{q: q}
}

func (r *tradingVolumeRepo) FindAll(ctx bCtx.Ctx, optsFns ...collection.TradingVolumeFindAllOptions) ([]collection.TradingVolume, error) {
	opts, err := collection.GetTradingVolumeFindAllOptions(optsFns...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"optsFns": optsFns,
			"err":     err,
		}).Error("GetTradingVolumeFindAllOptions failed")
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
	res := []collection.TradingVolume{}
	if err := r.q.Search(ctx, domain.TableTradingVolumes, offset, limit, sort, query, &res); err != nil {
		ctx.WithField("err", err).Error("q.Search failed")
		return nil, err
	}
	return res, nil
}

func (r *tradingVolumeRepo) FindOne(ctx bCtx.Ctx, id collection.TradingVolumeId) (*collection.TradingVolume, error) {
	v := &collection.TradingVolume{}
	err := r.q.FindOne(ctx, domain.TableTradingVolumes, id, v)
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

func (r *tradingVolumeRepo) Upsert(ctx bCtx.Ctx, v collection.TradingVolume) error {
	if err := r.q.Upsert(ctx, domain.TableTradingVolumes, v.ToId(), v); err != nil {
		ctx.WithFields(log.Fields{
			"tradingVolume": v,
			"err":           err,
		}).Error("q.Upsert failed")
		return err
	}
	return nil
}

func (r *tradingVolumeRepo) IncVolume(ctx bCtx.Ctx, id collection.TradingVolumeId, volume float64) (float64, error) {
	v := &collection.TradingVolume{}
	if err := r.q.Increment(ctx, domain.TableTradingVolumes, id, v, "volume", volume); err != nil {
		ctx.WithFields(log.Fields{
			"id":     id,
			"volume": volume,
			"err":    err,
		}).Error("q.Increment failed")
		return 0, err
	}
	return v.Volume, nil
}

func (r *tradingVolumeRepo) Patch(ctx bCtx.Ctx, id collection.TradingVolumeId, patchable collection.TradingVolumeUpdatePayload) error {
	if err := r.q.Patch(ctx, domain.TableTradingVolumes, id, patchable); err != nil {
		ctx.WithFields(log.Fields{
			"id":        id,
			"patchable": patchable,
			"err":       err,
		}).Error("q.Patch failed")
		return err
	}
	return nil
}
