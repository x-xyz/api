package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/service/query"
)

type orderNonceRepoImpl struct {
	q query.Mongo
}

func NewOrderNonceRepo(q query.Mongo) account.OrderNonceRepo {
	return &orderNonceRepoImpl{q}
}

func (im *orderNonceRepoImpl) FindOne(ctx ctx.Ctx, id account.OrderNonceId) (*account.OrderNonce, error) {
	qry, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("MakeBsonM failed")
	}

	var nonce account.OrderNonce
	err = im.q.FindOne(ctx, domain.TableOrderNonces, qry, &nonce)
	if err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"query": qry,
		}).Error("query.FindOne failed")
		return nil, err
	}
	return &nonce, nil
}

func (im *orderNonceRepoImpl) Upsert(ctx ctx.Ctx, nonce *account.OrderNonce) error {
	id := nonce.ToId()
	selector, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("MakeBsonM failed")
		return err
	}
	if err := im.q.Upsert(ctx, domain.TableOrderNonces, selector, nonce); err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"selector": selector,
			"nonce":    nonce,
		}).Error("q.Upsert failed")
		return err
	}
	return nil
}

func (im *orderNonceRepoImpl) Update(ctx ctx.Ctx, id account.OrderNonceId, patchable account.OrderNoncePatchable) error {
	selector, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("MakeBsonM failed")
		return err
	}
	updater, err := mongoclient.MakeBsonM(patchable)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"patchable": patchable,
		}).Error("MakeBsonM failed")
		return err
	}
	if err := im.q.Patch(ctx, domain.TableOrderNonces, selector, updater); err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"selector": selector,
			"updater":  updater,
		}).Error("q.Patch failed")
		return err
	}
	return nil
}
