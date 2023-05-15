package nftitem

import (
	"time"

	"github.com/x-xyz/goapi/domain"
)

type Offer struct {
	// raw data from contract
	Owner        domain.Address `json:"owner" bson:"owner"`
	PayToken     domain.Address `json:"payToken" bson:"payToken"`
	Quantity     string         `json:"quantity" bson:"quantity"`
	PricePerItem string         `json:"pricePerItem" bson:"pricePerItem"`
	Deadline     *time.Time     `json:"deadline" bson:"deadline"`
	CreatedAt    *time.Time     `json:"createdAt" bson:"createdAt"`

	// additional info
	BlockNumber   domain.BlockNumber `json:"blockNumber" bson:"blockNumber"`
	DisplayPrice  string             `json:"displayPrice" bson:"displayPrice"` // payment token, exact
	PriceInUsd    float64            `json:"priceInUsd" bson:"priceInUsd"`
	PriceInNative float64            `json:"priceInNative" bson:"priceInNative"`
}

func HighestOffer(offers []Offer) *Offer {
	now := time.Now()
	var highest *Offer
	for _, offer := range offers {
		if offer.Deadline.Before(now) {
			continue
		}
		if highest == nil || highest.PriceInUsd < offer.PriceInUsd {
			highest = &offer
		}
	}
	return highest
}
