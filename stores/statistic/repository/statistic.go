package repository

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/statistic"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type repo struct {
	q query.Mongo
}

func New(q query.Mongo) statistic.Repo {
	return &repo{q}
}

func (r *repo) FindOne(ctx bCtx.Ctx, key string) (*statistic.Statistic, error) {
	qry := bson.M{
		"key": key,
	}
	res := &statistic.Statistic{}
	err := r.q.FindOne(ctx, domain.TableStatistics, qry, &res)
	if err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *repo) Upsert(ctx bCtx.Ctx, s *statistic.Statistic) error {
	id := s.ToId()
	selector, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("MakeBsonM failed")
		return err
	}
	if err := r.q.Upsert(ctx, domain.TableStatistics, selector, s); err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"selector": selector,
			"s":        s,
		}).Error("q.Upsert failed")
		return err
	}
	return nil
}
