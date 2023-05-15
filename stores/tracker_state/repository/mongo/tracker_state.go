package mongo

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/query"
)

type trackerStateMongoRepo struct {
	m query.Mongo
}

func NewTrackerStateMongoRepo(mCon query.Mongo) domain.TrackerStateRepo {
	return &trackerStateMongoRepo{m: mCon}
}

func (r *trackerStateMongoRepo) Get(ctx bCtx.Ctx, id *domain.TrackerStateId) (*domain.TrackerState, error) {
	qry, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to make bson.M")
		return nil, err
	}

	state := &domain.TrackerState{}
	if err := r.m.FindOne(ctx, domain.TableTrackerStates, qry, state); err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  qry,
		}).Error("failed to FindOne")
		return nil, err
	}
	return state, nil
}

func (r *trackerStateMongoRepo) Update(ctx bCtx.Ctx, state *domain.TrackerState) error {
	selector, err := mongoclient.MakeBsonM(state.ToId())
	if err != nil {
		ctx.WithField("err", err).Error("failed to make bson.M")
		return err
	}
	if err := r.m.Patch(ctx, domain.TableTrackerStates, selector, state); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  state.ToId(),
		}).Error("failed to update")
		return err
	}
	return nil
}

func (r *trackerStateMongoRepo) Store(ctx bCtx.Ctx, state *domain.TrackerState) error {
	if err := r.m.Insert(ctx, domain.TableTrackerStates, state); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  state.ToId(),
		}).Error("failed to store")
		return err
	}
	return nil
}
