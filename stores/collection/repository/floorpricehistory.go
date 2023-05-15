package repository

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/service/query"
)

type floorPriceHistoryRepo struct {
	q query.Mongo
}

func NewFloorPriceHistoryRepo(q query.Mongo) collection.FloorPriceHistoryRepo {
	return &floorPriceHistoryRepo{q: q}
}

func (r *floorPriceHistoryRepo) FindOne(ctx bCtx.Ctx, id collection.FloorPriceId) (*collection.FloorPriceHistory, error) {
	h := &collection.FloorPriceHistory{}
	err := r.q.FindOne(ctx, domain.TableFloorPriceHistories, id, h)
	if err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		ctx.WithFields(log.Fields{
			"id":  id,
			"err": err,
		}).Error("q.FindOne failed")
		return nil, err
	}
	return h, nil
}

func (r *floorPriceHistoryRepo) Upsert(ctx bCtx.Ctx, h collection.FloorPriceHistory) error {
	if err := r.q.Upsert(ctx, domain.TableFloorPriceHistories, h.ToId(), h); err != nil {
		ctx.WithFields(log.Fields{
			"FloorPriceHistory": h,
			"err":               err,
		}).Error("q.Upsert failed")
		return err
	}
	return nil
}
