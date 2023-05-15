package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/promotion"
)

type promotionImpl struct {
	promotion promotion.PromotionRepo
}

func NewPromotion(promotion promotion.PromotionRepo) promotion.PromotionUsecase {
	return &promotionImpl{promotion}
}

func (im *promotionImpl) CreatePromotion(c ctx.Ctx, p *promotion.Promotion) (*promotion.Promotion, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		c.WithField("err", err).Error("failed to uuid.NewRandom")
		return nil, err
	}

	id := uuid.String()
	p.Id = id
	if err := im.promotion.Create(c, p); err != nil {
		c.WithField("err", err).Error("promotion.Create failed")
		return nil, err
	}
	return p, nil
}

func (im *promotionImpl) GetActivatedPromotions(c ctx.Ctx, ts *time.Time) ([]*promotion.Promotion, error) {
	items, err := im.promotion.FindAll(c, promotion.WithTargetTime(ts))
	if err != nil {
		c.WithField("err", err).Error("FindAll failed")
		return nil, err
	}
	return items, nil
}
