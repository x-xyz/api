package collection

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type FloorPriceHistory struct {
	ChainId              domain.ChainId `json:"chainId" bson:"chainId"`
	Address              domain.Address `json:"address" bson:"address"`
	Date                 time.Time      `json:"date" bson:"date"`                   // utc 00:00
	PriceInNative        float64        `json:"priceInNative" bson:"priceInNative"` // native
	PriceInUsd           float64        `json:"priceInUsd" bson:"priceInUsd"`
	OpenseaPriceInNative float64        `json:"openseaPriceInNative" bson:"openseaPriceInNative"`
	OpenseaPriceInUsd    float64        `json:"openseaPriceInUsd"`
	NumOwners            int64          `json:"numOwners" bson:"numOwners"`
}

func (p *FloorPriceHistory) ToId() FloorPriceId {
	return FloorPriceId{
		ChainId: p.ChainId,
		Address: p.Address,
		Date:    p.Date,
	}
}

type FloorPriceId struct {
	ChainId domain.ChainId `json:"chainId" bson:"chainId"`
	Address domain.Address `json:"address" bson:"address"`
	Date    time.Time      `json:"date" bson:"date"` // utc 00:00
}

type FloorPriceHistoryUseCase interface {
	FindOne(ctx.Ctx, FloorPriceId) (*FloorPriceHistory, error)
	Upsert(ctx.Ctx, FloorPriceHistory) error
}

type FloorPriceHistoryRepo interface {
	FindOne(ctx.Ctx, FloorPriceId) (*FloorPriceHistory, error)
	Upsert(ctx.Ctx, FloorPriceHistory) error
}
