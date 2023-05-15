package exchange

import (
	"math/big"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type CancelAllOrdersEvent struct {
	User     domain.Address
	NewNonce *big.Int
}

type CancelMultipleOrdersEvent struct {
	OrderItemHashes []domain.OrderHash
}

type Fulfillment struct {
	Collection domain.Address
	TokenId    *big.Int
	Amount     *big.Int
	Currency   domain.Address
	Price      *big.Int
}

type TakerAskEvent struct {
	Taker         domain.Address
	Maker         domain.Address
	Strategy      domain.Address
	OrderHash     domain.OrderHash
	ItemIdx       *big.Int
	OrderItemHash domain.OrderHash
	Fulfillment   Fulfillment
	Marketplace   domain.MarketplaceHash
}

type TakerBidEvent struct {
	Taker         domain.Address
	Maker         domain.Address
	Strategy      domain.Address
	OrderHash     domain.OrderHash
	ItemIdx       *big.Int
	OrderItemHash domain.OrderHash
	Fulfillment   Fulfillment
	Marketplace   domain.MarketplaceHash
}

type UseCase interface {
	CancelAllOrders(ctx.Ctx, domain.ChainId, *CancelAllOrdersEvent, *domain.LogMeta) error
	CancelMultipleOrders(ctx.Ctx, domain.ChainId, *CancelMultipleOrdersEvent, *domain.LogMeta) error
	TakerAsk(ctx.Ctx, domain.ChainId, *TakerAskEvent, *domain.LogMeta) error
	TakerBid(ctx.Ctx, domain.ChainId, *TakerBidEvent, *domain.LogMeta) error
}
