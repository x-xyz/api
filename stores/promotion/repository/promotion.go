package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/promotion"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type promotionImpl struct {
	q query.Mongo
}

func NewPromotion(q query.Mongo) promotion.PromotionRepo {
	return &promotionImpl{q}
}

func (im *promotionImpl) Create(c ctx.Ctx, value *promotion.Promotion) error {
	if err := im.q.Insert(c, domain.TablePromotions, value); err != nil {
		return err
	}
	return nil
}

func (im *promotionImpl) FindAll(c ctx.Ctx, optFns ...promotion.SelectOptions) ([]*promotion.Promotion, error) {
	opts, err := promotion.GetSelectOptions(optFns...)
	if err != nil {
		c.WithField("err", err).Error("promotion.GetSelectOptions failed")
		return nil, err
	}

	qry, err := mongoclient.MakeBsonM(opts)

	if err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return nil, err
	}

	if opts.TargetTime != nil {
		qry = bson.M{
			"startTime": bson.M{"$lte": opts.TargetTime},
			"endTime":   bson.M{"$gt": opts.TargetTime},
		}
	}

	res := []*promotion.Promotion{}

	if err := im.q.Search(c, domain.TablePromotions, 0, 0, "_id", qry, &res); err != nil {
		c.WithField("err", err).Error("q.Search failed")
		return nil, err
	} else {
		return res, nil
	}
}
