package nftitem

import (
	"time"

	"github.com/x-xyz/goapi/domain"
)

type Auction struct {
	// raw data from contract
	Owner        domain.Address `json:"owner" bson:"owner"`
	PayToken     domain.Address `json:"payToken" bson:"payToken"`
	ReservePrice string         `json:"reservePrice" bson:"reservePrice"`
	StartTime    *time.Time     `json:"startTime" bson:"startTime"`
	EndTime      *time.Time     `json:"endTime" bson:"endTime"`

	// additional info
	BlockNumber   domain.BlockNumber `json:"blockNumber" bson:"blockNumber"`
	DisplayPrice  string             `json:"displayPrice" bson:"displayPrice"` // payment token, exact
	PriceInUsd    float64            `json:"priceInUsd" bson:"priceInUsd"`
	PriceInNative float64            `json:"priceInNative" bson:"priceInNative"`
}
