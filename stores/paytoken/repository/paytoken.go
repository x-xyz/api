package repository

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/query"
)

type payTokenMongoRepo struct {
	q query.Mongo
}

func NewPayTokenRepo(q query.Mongo) domain.PayTokenRepo {
	return &payTokenMongoRepo{
		q: q,
	}
}

func (r *payTokenMongoRepo) FindOne(ctx bCtx.Ctx, chainId domain.ChainId, tokenAddress domain.Address) (*domain.PayToken, error) {
	payToken := &domain.PayToken{}
	if qry, err := mongoclient.MakeBsonM(&domain.PayToken{ChainId: chainId, Address: tokenAddress}); err != nil {
		ctx.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return nil, err
	} else if err := r.q.FindOne(ctx, domain.TablePayTokens, qry, payToken); err != nil && err != query.ErrNotFound {
		ctx.WithField("err", err).Error("q.FindOne failed")
		return nil, err
	} else if err == query.ErrNotFound {
		return nil, nil
	}
	return payToken, nil
}

func (r *payTokenMongoRepo) Create(ctx bCtx.Ctx, payToken *domain.PayToken) error {
	if err := r.q.Insert(ctx, domain.TablePayTokens, payToken); err != nil {
		ctx.WithField("err", err).Error("q.Insert failed")
		return err
	}
	return nil
}

func (r *payTokenMongoRepo) Upsert(ctx bCtx.Ctx, payToken *domain.PayToken) error {
	selector, err := mongoclient.MakeBsonM(payToken.ToId())
	if err != nil {
		ctx.WithField("err", err).Error("failed to make bson.M")
		return err
	}
	if err := r.q.Upsert(ctx, domain.TablePayTokens, selector, payToken); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  payToken.ToId(),
		}).Error("failed to update")
		return err
	}
	return nil
}
