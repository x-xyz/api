package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection_promotion"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type collPromotionImpl struct {
	q query.Mongo
}

func NewCollPromotion(q query.Mongo) collection_promotion.CollPromotionRepo {
	return &collPromotionImpl{q}
}

func (im *collPromotionImpl) Create(c ctx.Ctx, collPromos []collection_promotion.CollPromotion, promotionId *string) error {
	ops := []query.UpsertOp{}
	for _, id := range collPromos {
		ops = append(ops, query.UpsertOp{
			Selector: bson.M{
				"chainId":       id.ChainId,
				"erc721Address": id.Address,
				"promotionId":   promotionId,
			},
			Updater: bson.M{
				"chainId":       id.ChainId,
				"erc721Address": id.Address,
				"reward":        id.Reward,
				"subReward":     id.SubReward,
				"multiplier":    id.Multiplier,
				"promotionId":   promotionId,
			},
		})
	}
	if _, _, err := im.q.BulkUpsert(c, domain.TableCollPromotions, ops); err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to BulkUpsert")
		return err
	}
	return nil
}

func (im *collPromotionImpl) FindAll(c ctx.Ctx, optFns ...collection_promotion.SelectOptions) ([]*collection_promotion.CollPromotion, error) {
	opts, err := collection_promotion.GetSelectOptions(optFns...)
	if err != nil {
		c.WithField("err", err).Error("collection_promotion.GetSelectOptions failed")
		return nil, err
	}

	qry, err := mongoclient.MakeBsonM(opts)
	if err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return nil, err
	}

	if len(opts.PromotionIds) > 1 {
		qry["promotionId"] = bson.M{"$in": opts.PromotionIds}
	} else if len(opts.PromotionIds) > 0 {
		qry["promotionId"] = opts.PromotionIds[0]
	}
	res := []*collection_promotion.CollPromotion{}

	if err := im.q.Search(c, domain.TableCollPromotions, 0, 0, "_id", qry, &res); err != nil {
		c.WithField("err", err).Error("q.Search failed")
		return nil, err
	} else {
		return res, nil
	}
}
