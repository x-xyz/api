package nftitem

import (
	"time"

	"github.com/x-xyz/goapi/domain"
)

type Bid struct {
	// raw data from contract
	Owner    domain.Address `json:"owner" bson:"owner"`
	PayToken domain.Address `json:"payToken" bson:"payToken"`
	Bid      string         `json:"bid" bson:"bid"`
	BidTime  *time.Time     `json:"bidTime" bson:"bidTime"`

	// additional info
	BlockNumber   domain.BlockNumber `json:"blockNumber" bson:"blockNumber"`
	DisplayPrice  string             `json:"displayPrice" bson:"displayPrice"` // payment token, exact
	PriceInUsd    float64            `json:"priceInUsd" bson:"priceInUsd"`
	PriceInNative float64            `json:"priceInNative" bson:"priceInNative"`
}
