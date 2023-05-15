package domain

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
)

type VexFeeDistributionHistory struct {
	VexTotalSupply    float64     `json:"-" bson:"vexTotalSupply"`
	VexTotalSupplyBlk BlockNumber `json:"-" bson:"vexTotalSupplyBlk"`
	WethEmission      float64     `json:"-" bson:"wethEmiision"`
	XPrice            float64     `json:"-" bson:"xPrice"`
	EthPrice          float64     `json:"-" bson:"ethPrice"`
	PriceBlk          BlockNumber `json:"-" bson:"priceBlk"`
	Apr               float64     `json:"apr" bson:"apr"` // weth * ethPrice / vexSupply / XPrice * 52
	Time              *time.Time  `json:"-" bson:"time"`
}

type VexFeeDistributionHistoryRepo interface {
	FindLatest(ctx.Ctx, int) ([]VexFeeDistributionHistory, error)
	Create(ctx.Ctx, *VexFeeDistributionHistory) error
}

type VexFeeDistributionHistoryUseCase interface {
	LatestApr(ctx.Ctx, int) (float64, error)
	Create(ctx.Ctx, *VexFeeDistributionHistory) error
}
