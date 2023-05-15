package airdrop

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type AirdropType string

const (
	AirdropTypeOnce  = "once"
	AirdropTypeRound = "round"
)

type Airdrop struct {
	Name               string         `json:"name" bson:"name"`
	Image              string         `json:"image" bson:"image"`
	ChainId            domain.ChainId `json:"chainId" bson:"chainId"`
	ContractAddress    domain.Address `json:"contractAddress" bson:"contractAddress"`
	RewardTokenAddress domain.Address `json:"rewardTokenAddress" bson:"rewardTokenAddress"`
	Type               AirdropType    `json:"type" bson:"type"`
	Deadline           time.Time      `json:"deadline" bson:"deadline"`
}

type AirdropFindAllOptions struct {
	SortBy        *string         `bson:"-"`
	SortDir       *domain.SortDir `bson:"-"`
	Offset        *int32          `bson:"-"`
	Limit         *int32          `bson:"-"`
	DeadlineAfter *time.Time      `bson:"-"`
}

type AirdropFindAllOptionsFunc func(*AirdropFindAllOptions) error

func GetAirdropFindAllOptions(opts ...AirdropFindAllOptionsFunc) (AirdropFindAllOptions, error) {
	res := AirdropFindAllOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func AirdropWithSort(sortby string, sortdir domain.SortDir) AirdropFindAllOptionsFunc {
	return func(options *AirdropFindAllOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func AirdropWithPagination(offset int32, limit int32) AirdropFindAllOptionsFunc {
	return func(options *AirdropFindAllOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func AirdropWithDeadlineAfter(deadline time.Time) AirdropFindAllOptionsFunc {
	return func(options *AirdropFindAllOptions) error {
		options.DeadlineAfter = &deadline
		return nil
	}
}

type AirdropRepo interface {
	FindAll(ctx.Ctx, ...AirdropFindAllOptionsFunc) ([]Airdrop, error)
	Create(ctx.Ctx, *Airdrop) error
}

type AirdropUseCase interface {
	FindAll(ctx.Ctx, ...AirdropFindAllOptionsFunc) ([]Airdrop, error)
	Create(ctx.Ctx, *Airdrop) error
}
