package account

import (
	"time"

	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type ActivityType string

const (
	ActivityTypeBid           ActivityType = "bid"
	ActivityTypeOffer         ActivityType = "offer"
	ActivityTypeCancelOffer   ActivityType = "cancelOffer"
	ActivityTypeList          ActivityType = "list"
	ActivityTypeCancelListing ActivityType = "cancelListing"
	ActivityTypeBuy           ActivityType = "buy"
	ActivityTypeSale          ActivityType = "sale"
	ActivityTypeTransfer      ActivityType = "transfer"
	ActivityTypeMint          ActivityType = "mint"
	ActivityTypeSale2         ActivityType = "sale"
)

type Activity struct {
	Type         ActivityType          `json:"type"`
	Token        nftitem.SimpleNftItem `json:"token"`
	Owner        SimpleAccount         `json:"owner"`
	To           SimpleAccount         `json:"to"`
	Quantity     int32                 `json:"quantity"`
	Price        float64               `json:"price"`
	PaymentToken domain.Address        `json:"paymentToken"`
	PriceInUsd   float64               `json:"priceInUsd"`
	TxHash       domain.TxHash         `json:"txHash"`
	CreatedAt    time.Time             `json:"createdAt"`
}

type ActivityResult struct {
	Activities []*Activity `json:"activities"`
	Count      int         `json:"count"`
}
