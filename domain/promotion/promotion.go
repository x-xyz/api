package promotion

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
)

type SearchParams struct {
	// required
	TS *string `query:"ts"`
}

type Promotion struct {
	Id                    string     `bson:"id"`
	Name                  string     `bson:"name"`
	StartTime             *time.Time `bson:"startTime"`
	EndTime               *time.Time `bson:"endTime"`
	RewardPerDistribution string     `bson:"rewardPerDistribution"`
}

type selectOptions struct {
	TargetTime *time.Time `bson:"targetTime"`
}

type SelectOptions func(*selectOptions) error

func GetSelectOptions(opts ...SelectOptions) (selectOptions, error) {
	res := selectOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

type PromotionRepo interface {
	FindAll(c ctx.Ctx, opts ...SelectOptions) ([]*Promotion, error)
	Create(c ctx.Ctx, value *Promotion) error
}

type PromotionUsecase interface {
	CreatePromotion(c ctx.Ctx, value *Promotion) (*Promotion, error)
	GetActivatedPromotions(c ctx.Ctx, ts *time.Time) ([]*Promotion, error)
}

func WithTargetTime(targetTime *time.Time) SelectOptions {
	return func(options *selectOptions) error {
		options.TargetTime = targetTime
		return nil
	}
}
