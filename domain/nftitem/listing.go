package nftitem

import (
	"time"

	"github.com/x-xyz/goapi/domain"
)

type Listing struct {
	// raw data from contract
	Owner        domain.Address `json:"owner" bson:"owner"`
	Quantity     string         `json:"quantity" bson:"quantity"`
	PayToken     domain.Address `json:"payToken" bson:"payToken"`
	PricePerItem string         `json:"pricePerItem" bson:"pricePerItem"`
	StartingTime *time.Time     `json:"startTime" bson:"startTime"`
	Deadline     *time.Time     `json:"deadline" bson:"deadline"`

	// additional info
	BlockNumber   domain.BlockNumber `json:"blockNumber" bson:"blockNumber"`
	DisplayPrice  string             `json:"displayPrice" bson:"displayPrice"` // payment token, exact
	PriceInUsd    float64            `json:"priceInUsd" bson:"priceInUsd"`
	PriceInNative float64            `json:"priceInNative" bson:"priceInNative"`
}
