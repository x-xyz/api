package repository

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type impl struct {
	q query.Mongo
}

func NewOrderItemRepo(q query.Mongo) order.OrderItemRepo {
	return &impl{q}
}

func (im *impl) makeQuery(options order.OrderItemFindAllOptions) (bson.M, error) {
	query := bson.M{}

	if options.OrderHash != nil {
		query["orderHash"] = *options.OrderHash
	}

	if options.OrderItemHash != nil {
		query["orderItemHash"] = *options.OrderItemHash
	}

	if options.NftitemId != nil {
		query["chainId"] = options.NftitemId.ChainId
		query["collection"] = options.NftitemId.ContractAddress.ToLower()
		query["tokenID"] = options.NftitemId.TokenId
	}

	if options.IsValid != nil {
		query["isValid"] = *options.IsValid
	}

	if options.IsAsk != nil {
		query["isAsk"] = *options.IsAsk
	}

	if options.Signer != nil {
		query["signer"] = *options.Signer
	}

	if options.NonceLT != nil {
		nonce, ok := new(big.Int).SetString(*options.NonceLT, 10)
		if !ok {
			return nil, domain.ErrInvalidNumberFormat
		}
		query["hexNonce"] = bson.M{"$lt": hexutil.Encode(math.U256Bytes(nonce))}
	}

	startTimeQuery := bson.M{}
	if options.StartTimeGT != nil {
		startTimeQuery["$gt"] = *options.StartTimeGT
	}

	if options.StartTimeLT != nil {
		startTimeQuery["$lt"] = *options.StartTimeLT
	}

	if len(startTimeQuery) > 0 {
		query["startTime"] = startTimeQuery
	}

	endTimeQuery := bson.M{}
	if options.EndTimeGT != nil {
		endTimeQuery["$gt"] = *options.EndTimeGT
	}

	if options.EndTimeLT != nil {
		endTimeQuery["$lt"] = *options.EndTimeLT
	}

	if len(endTimeQuery) > 0 {
		query["endTime"] = endTimeQuery
	}

	if options.IsUsed != nil {
		query["isUsed"] = *options.IsUsed
	}

	if options.ChainId != nil {
		query["chainId"] = *options.ChainId
	}

	if options.Collection != nil {
		query["collection"] = *options.Collection
	}

	if options.Strategy != nil {
		query["strategy"] = *options.Strategy
	}

	return query, nil
}

func (im *impl) FindAll(ctx ctx.Ctx, opts ...order.OrderItemFindAllOptionsFunc) ([]*order.OrderItem, error) {
	options, err := order.GetOrderItemFindAllOptions(opts...)
	if err != nil {
		return nil, err
	}
	query, err := im.makeQuery(options)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("im.makeQuery")
		return nil, err
	}

	offset := 0
	if options.Offset != nil {
		offset = int(*options.Offset)
	}

	limit := 0
	if options.Limit != nil {
		limit = int(*options.Limit)
	}

	sort := "itemIdx"
	if options.Sort != nil {
		sort = *options.Sort
	}

	res := []*order.OrderItem{}
	err = im.q.Search(ctx, domain.TableOrderItems, offset, limit, sort, query, &res)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"query": query,
		}).Error("failed to q.Search")
		return nil, err
	}

	return res, nil
}

func (im *impl) FindOne(ctx ctx.Ctx, itemId order.OrderItemId) (*order.OrderItem, error) {
	query, err := mongoclient.MakeBsonM(itemId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":    err,
			"itemId": itemId,
		}).Error("failed to mongoclient.MakeBsonM")
		return nil, err
	}

	res := order.OrderItem{}
	err = im.q.FindOne(ctx, domain.TableOrderItems, query, &res)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"query": query,
		}).Error("failed to q.FindOne")
		return nil, err
	}

	return &res, nil
}

func (im *impl) Upsert(ctx ctx.Ctx, orderItem *order.OrderItem) error {
	id := orderItem.ToId()
	selector, err := mongoclient.MakeBsonM(id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to mongoclient.MakeBsonM")
		return err
	}

	err = im.q.Upsert(ctx, domain.TableOrderItems, selector, orderItem)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"selector":  selector,
			"orderItem": *orderItem,
		}).Error("failed to q.Upsert")
		return err
	}
	return nil
}

func (im *impl) Update(ctx ctx.Ctx, itemId order.OrderItemId, patchalbe order.OrderItemPatchable) error {
	selector, err := mongoclient.MakeBsonM(itemId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":    err,
			"itemId": itemId,
		}).Error("failed to mongoclient.MakeBsonM")
		return err
	}

	updater, err := mongoclient.MakeBsonM(patchalbe)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"patchable": patchalbe,
		}).Error("failed to mongoclient.MakeBsonM")
		return err
	}

	err = im.q.Patch(ctx, domain.TableOrderItems, selector, updater)
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

func (im *impl) RemoveAll(ctx ctx.Ctx, opts ...order.OrderItemFindAllOptionsFunc) error {
	options, err := order.GetOrderItemFindAllOptions(opts...)
	if err != nil {
		return err
	}
	query, err := im.makeQuery(options)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to im.makeQuery")
		return err
	}

	fmt.Println(query)

	_, err = im.q.RemoveAll(ctx, domain.TableOrderItems, query)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"query": query,
		}).Error("failed to q.RemoveAll")
		return err
	}

	return nil
}

func (im *impl) FindOneOrder(ctx ctx.Ctx, orderHash string) (*order.Order, error) {
	res := order.Order{}
	err := im.q.FindOne(ctx, domain.TableOrders, bson.M{
		"orderHash": orderHash,
	}, &res)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to q.FindOne")
		return nil, err
	}
	return &res, nil
}
