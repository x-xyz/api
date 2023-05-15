package repository

import (
	"fmt"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type orderRepoImpl struct {
	q query.Mongo
}

func NewOrderRepo(q query.Mongo) order.OrderRepo {
	return &orderRepoImpl{q}
}

func (im *orderRepoImpl) makeQuery(opts ...order.OrderFindAllOptionsFunc) (bson.M, error) {
	options, err := order.GetOrderFindAllOptions(opts...)
	if err != nil {
		return nil, err
	}
	query := bson.M{}

	if options.ChainId != nil {
		query["chainId"] = *options.ChainId
	}

	if options.OrderHash != nil {
		query["orderHash"] = *options.OrderHash
	}

	if options.IsAsk != nil {
		query["isAsk"] = *options.IsAsk
	}

	if options.Signer != nil {
		query["signer"] = *options.Signer
	}

	if options.Nonce != nil {
		query["nonce"] = *options.Nonce
	}

	startTimeQuery := bson.M{}
	if options.StartTimeGT != nil {
		startTimeQuery["$gt"] = fmt.Sprint(options.StartTimeGT.Unix())
	}

	if options.StartTimeLT != nil {
		startTimeQuery["$lt"] = fmt.Sprint(options.StartTimeLT.Unix())
	}

	if len(startTimeQuery) > 0 {
		query["startTime"] = startTimeQuery
	}

	endTimeQuery := bson.M{}
	if options.EndTimeGT != nil {
		endTimeQuery["$gt"] = fmt.Sprint(options.EndTimeGT.Unix())
	}

	if options.EndTimeLT != nil {
		endTimeQuery["$lt"] = fmt.Sprint(options.EndTimeLT.Unix())
	}

	if len(endTimeQuery) > 0 {
		query["endTime"] = endTimeQuery
	}

	return query, nil
}

func (im *orderRepoImpl) FindAll(ctx ctx.Ctx, opts ...order.OrderFindAllOptionsFunc) ([]*order.Order, error) {
	query, err := im.makeQuery(opts...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("im.makeQuery")
		return nil, err
	}

	res := []*order.Order{}
	err = im.q.Search(ctx, domain.TableOrders, 0, 0, "_id", query, &res)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"query": query,
		}).Error("failed to q.Search")
		return nil, err
	}

	return res, nil
}

func (im *orderRepoImpl) Count(ctx ctx.Ctx, opts ...order.OrderFindAllOptionsFunc) (int, error) {
	query, err := im.makeQuery(opts...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("im.makeQuery")
		return 0, err
	}

	cnt, err := im.q.Count(ctx, domain.TableOrders, query)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"query": query,
		}).Error("failed to q.Count")
		return 0, err
	}

	return cnt, nil
}

func (im *orderRepoImpl) RemoveAll(ctx ctx.Ctx, opts ...order.OrderFindAllOptionsFunc) error {
	query, err := im.makeQuery(opts...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("im.makeQuery")
		return err
	}

	_, err = im.q.RemoveAll(ctx, domain.TableOrders, query)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"query": query,
		}).Error("q.RemoveAll failed")
		return err
	}

	return nil
}

func (im *orderRepoImpl) FindOne(ctx ctx.Ctx, id order.OrderId) (*order.Order, error) {
	qry, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to mongoclient.MakeBsonM")
		return nil, err
	}

	res := order.Order{}
	err = im.q.FindOne(ctx, domain.TableOrders, qry, &res)
	if err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"query": qry,
		}).Error("failed to q.FindOne")
		return nil, err
	}

	return &res, nil
}

func (im *orderRepoImpl) Upsert(ctx ctx.Ctx, _order *order.Order) error {
	id := _order.ToId()
	selector, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to mongoclient.MakeBsonM")
		return err
	}

	err = im.q.Upsert(ctx, domain.TableOrders, selector, _order)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"selector": selector,
			"order":    *_order,
		}).Error("failed to q.Upsert")
		return err
	}
	return nil
}

func (im *orderRepoImpl) Update(ctx ctx.Ctx, id order.OrderId, patchable order.OrderPatchable) error {
	selector, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to mongoclient.MakeBsonM")
		return err
	}

	updater, err := mongoclient.MakeBsonM(patchable)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"patchable": patchable,
		}).Error("failed to mongoclient.MakeBsonM")
		return err
	}

	err = im.q.Patch(ctx, domain.TableOrders, selector, updater)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"selector": selector,
			"updater":  updater,
		}).Error("failed to q.Patch")
		return err
	}

	return nil
}
