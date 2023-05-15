package collection_promotion

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/promotion"
)

type CollPromotion struct {
	ChainId     domain.ChainId `json:"chainId" bson:"chainId"`
	Address     domain.Address `json:"erc721Address" bson:"erc721Address"`
	Reward      string         `json:"reward" bson:"reward"`         // flat rewards
	SubReward   string         `json:"subReward" bson:"subReward"`   // flat rewards
	Multiplier  int            `json:"multiplier" bson:"multiplier"` // fixed-total rewards
	PromotionId string         `json:"promotionId" bson:"promotionId"`
}

type selectOptions struct {
	PromotionIds []string       `bson:"promotionId"`
	ChainId      domain.ChainId `bson:"chainId"`
}

type SelectOptions func(*selectOptions) error

type ListingRewardDistribution struct {
	Rewards            map[domain.Address]string
	CollectionListings map[domain.Address]int
	AccountListings    map[domain.Address]map[domain.Address]int
}

type CollPromotionRepo interface {
	Create(c ctx.Ctx, collectionIds []CollPromotion, promotionId *string) error
	FindAll(c ctx.Ctx, optFns ...SelectOptions) ([]*CollPromotion, error)
}

type CollPromotionUsecase interface {
	CreateCollPromotion(c ctx.Ctx, collectionIds []CollPromotion, promotionId *string) error
	GetCollPromotions(c ctx.Ctx, promotionIds *[]string) ([]*CollPromotion, error)
	GetPromotedCollections(c ctx.Ctx, ts *time.Time) (*promotion.Promotion, []*CollPromotion, error)
	CalculateLastHourAverageRewardPerListing(c ctx.Ctx) (string, error)
	CalculateListingRewardsFlat(c ctx.Ctx, beginTime time.Time, endTime time.Time) (*ListingRewardDistribution, error)
	CalculateListingRewardsFixedTotal(c ctx.Ctx, beginTime time.Time, endTime time.Time) (*ListingRewardDistribution, error)
	CreateWeeklyPromotion(c ctx.Ctx, name string, startTime *time.Time, endTime *time.Time, topK int32, reward decimal.Decimal) ([]CollPromotion, error)
}

func GetSelectOptions(opts ...SelectOptions) (selectOptions, error) {
	res := selectOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func WithPromotionIds(promotionIds *[]string) SelectOptions {
	return func(options *selectOptions) error {
		options.PromotionIds = *promotionIds
		return nil
	}
}

func WithChainId(chainId *domain.ChainId) SelectOptions {
	return func(options *selectOptions) error {
		options.ChainId = *chainId
		return nil
	}
}
