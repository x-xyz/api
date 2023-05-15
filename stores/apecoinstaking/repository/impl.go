package repository

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/apecoinstaking"
	"github.com/x-xyz/goapi/service/query"
)

type repo struct {
	q query.Mongo
}

func New(q query.Mongo) apecoinstaking.Repo {
	return &repo{q}
}

func (r *repo) FindOne(ctx bCtx.Ctx, id apecoinstaking.Id) (*apecoinstaking.ApecoinStaking, error) {
	qry, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{"id": id, "err": err}).Error("MakeBsonM failed")
		return nil, err
	}
	res := &apecoinstaking.ApecoinStaking{}
	err = r.q.FindOne(ctx, domain.TableApeStakings, qry, &res)
	if err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		ctx.WithFields(log.Fields{"qry": qry, "err": err}).Error("q.FindOne failed")
		return nil, err
	}
	return res, nil
}

func (r *repo) Upsert(ctx bCtx.Ctx, s *apecoinstaking.ApecoinStaking) error {
	id := s.ToId()
	selector, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("MakeBsonM failed")
		return err
	}
	if err := r.q.Upsert(ctx, domain.TableApeStakings, selector, s); err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"selector": selector,
			"s":        s,
		}).Error("q.Upsert failed")
		return err
	}
	return nil
}
