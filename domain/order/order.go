package order

import (
	"math/big"
	"strings"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type Item struct {
	Collection domain.Address `json:"collection" bson:"collection"`
	TokenId    domain.TokenId `json:"tokenId" bson:"tokenID"`
	Amount     string         `json:"amount" bson:"amount"`
	Price      string         `json:"price" bson:"price"`
}

func (i *Item) LowerCase() {
	i.Collection = i.Collection.ToLower()
}

type OrderItem struct {
	ChainId            domain.ChainId `json:"chainId" bson:"chainId"`
	Item               `bson:"inline"`
	ItemIdx            int              `json:"itemIdx" bson:"itemIdx"`
	OrderHash          domain.OrderHash `json:"orderHash" bson:"orderHash"`
	OrderItemHash      domain.OrderHash `json:"orderItemHash" bson:"orderItemHash"`
	IsAsk              bool             `json:"isAsk" bson:"isAsk"`
	Signer             domain.Address   `json:"signer" bson:"signer"`
	Nonce              string           `json:"nonce" bson:"nonce"`
	HexNonce           string           `json:"hexNonce" bson:"hexNonce"`
	Currency           domain.Address   `json:"currency" bson:"currency"`
	StartTime          time.Time        `json:"startTime" bson:"startTime"`
	EndTime            time.Time        `json:"endTime" bson:"endTime"`
	MinPercentageToAsk string           `json:"minPercentageToAsk" bson:"minPercentageToAsk"`
	Marketplace        string           `json:"marketplace" bson:"marketplace"`
	Strategy           Strategy         `json:"strategy" bson:"strategy"`
	ReservedBuyer      domain.Address   `json:"reservedBuyer" bson:"reservedBuyer"`
	PriceInUsd         float64          `json:"priceInUsd" bson:"priceInUsd"`
	PriceInNative      float64          `json:"priceInNative" bson:"priceInNative"`
	DisplayPrice       string           `json:"displayPrice" bson:"displayPrice"`

	// valid if:
	// - when IsAsk:
	//  1. Signer equals nftitem owner when token type == 721
	//  2. Signer's nftitem balance > amount when token type == 1155
	IsValid bool `json:"isValid" bson:"isValid"`

	// true if order is canceled or order is taken
	IsUsed bool `json:"isUsed" bson:"isUsed"`
}

type OrderItemPatchable struct {
	IsValid       *bool    `json:"isValid" bson:"isValid,omitempty"`
	IsUsed        *bool    `json:"isUsed" bson:"isUsed,omitempty"`
	PriceInUsd    *float64 `json:"priceInUsd" bson:"priceInUsd,omitempty"`
	PriceInNative *float64 `json:"priceInNative" bson:"priceInNative,omitempty"`
	DisplayPrice  *string  `json:"displayPrice" bson:"displayPrice,omitempty"`
}

func (o OrderItem) ToId() OrderItemId {
	return OrderItemId{
		ChainId:   o.ChainId,
		OrderHash: o.OrderHash,
		ItemIdx:   o.ItemIdx,
	}
}

type OrderItemId struct {
	ChainId   domain.ChainId   `json:"chainId" bson:"chainId"`
	OrderHash domain.OrderHash `json:"orderHash" bson:"orderHash"`
	ItemIdx   int              `json:"itemIdx" bson:"itemIdx"`
}

type FeeDistType string

const (
	FeeDistTypeBurn   FeeDistType = "burn"
	FeeDistTypeDonate FeeDistType = "donate"
)

func ToFeeDistType(name FeeDistType) FeeDistType {
	switch name {
	case FeeDistTypeBurn:
		return FeeDistTypeBurn
	case FeeDistTypeDonate:
		return FeeDistTypeDonate
	}
	return FeeDistTypeBurn
}

type Order struct {
	ChainId   domain.ChainId   `json:"chainId" bson:"chainId"`
	OrderHash domain.OrderHash `json:"orderHash" bson:"orderHash"`
	IsAsk     bool             `json:"isAsk" bson:"isAsk"`
	Signer    domain.Address   `json:"signer" bson:"signer"`
	Items     []Item           `json:"items" bson:"items"`
	Strategy  domain.Address   `json:"strategy" bson:"strategy"`
	Currency  domain.Address   `json:"currency" bson:"currency"`
	Nonce     string           `json:"nonce" bson:"nonce"`
	// string format in unix timestamp
	StartTime string `json:"startTime" bson:"startTime"`
	// string format in unix timestamp
	EndTime            string      `json:"endTime" bson:"endTime"`
	MinPercentageToAsk string      `json:"minPercentageToAsk" bson:"minPercentageToAsk"`
	Marketplace        string      `json:"marketplace" bson:"marketplace"`
	Params             string      `json:"params" bson:"params"`
	V                  int         `json:"v" bson:"v"`
	R                  string      `json:"r" bson:"r"`
	S                  string      `json:"s" bson:"s"`
	FeeDistType        FeeDistType `json:"feeDistType" bson:"feeDistType"`
}

func (o *Order) ToId() OrderId {
	return OrderId{
		ChainId:   o.ChainId,
		OrderHash: o.OrderHash,
	}
}

func (o *Order) LowerCase() {
	o.OrderHash = o.OrderHash.ToLower()
	o.Signer = o.Signer.ToLower()
	o.Strategy = o.Strategy.ToLower()
	o.Currency = o.Currency.ToLower()
	o.Marketplace = strings.ToLower(o.Marketplace)
	for i := range o.Items {
		o.Items[i].LowerCase()
	}
}

type OrderId struct {
	ChainId   domain.ChainId   `json:"chainId" bson:"chainId"`
	OrderHash domain.OrderHash `json:"orderHash" bson:"orderHash"`
}

type OrderPatchable struct {
}

type OrderItemFindAllOptions struct {
	OrderHash     *domain.OrderHash
	OrderItemHash *domain.OrderHash
	NftitemId     *nftitem.Id
	Signer        *domain.Address
	NonceLT       *string
	IsValid       *bool
	IsAsk         *bool
	StartTimeGT   *time.Time
	StartTimeLT   *time.Time
	EndTimeLT     *time.Time
	EndTimeGT     *time.Time
	IsUsed        *bool
	Offset        *int32
	Limit         *int32
	ChainId       *domain.ChainId
	Collection    *domain.Address
	Sort          *string
	Strategy      *Strategy
}

type OrderItemFindAllOptionsFunc func(*OrderItemFindAllOptions) error

func GetOrderItemFindAllOptions(opts ...OrderItemFindAllOptionsFunc) (OrderItemFindAllOptions, error) {
	res := OrderItemFindAllOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func WithOrderHash(orderHash domain.OrderHash) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.OrderHash = &orderHash
		return nil
	}
}

func WithOrderItemHash(orderItemHash domain.OrderHash) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.OrderItemHash = &orderItemHash
		return nil
	}
}

func WithNftItemId(nftitemId nftitem.Id) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.NftitemId = &nftitemId
		return nil
	}
}

func WithIsValid(isValid bool) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.IsValid = &isValid
		return nil
	}
}

func WithIsAsk(isAsk bool) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.IsAsk = &isAsk
		return nil
	}
}

func WithSigner(signer domain.Address) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.Signer = &signer
		return nil
	}
}

func WithNonceLT(nonce string) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.NonceLT = &nonce
		return nil
	}
}

func WithStartTimeGT(t time.Time) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.StartTimeGT = &t
		return nil
	}
}

func WithStartTimeLT(t time.Time) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.StartTimeLT = &t
		return nil
	}
}

func WithEndTimeGT(t time.Time) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.EndTimeGT = &t
		return nil
	}
}

func WithEndTimeLT(t time.Time) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.EndTimeLT = &t
		return nil
	}
}

func WithIsUsed(isUsed bool) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.IsUsed = &isUsed
		return nil
	}
}

func WithPagination(offset int32, limit int32) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func WithChainId(chainId domain.ChainId) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func WithContractAddress(address domain.Address) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.Collection = &address
		return nil
	}
}

func WithSort(sort string) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.Sort = &sort
		return nil
	}
}

func WithStrategy(s Strategy) OrderItemFindAllOptionsFunc {
	return func(options *OrderItemFindAllOptions) error {
		options.Strategy = &s
		return nil
	}
}

type OrderFindAllOptions struct {
	ChainId     *domain.ChainId
	OrderHash   *domain.OrderHash
	IsAsk       *bool
	Signer      *domain.Address
	Nonce       *string
	StartTimeGT *time.Time
	StartTimeLT *time.Time
	EndTimeLT   *time.Time
	EndTimeGT   *time.Time
}

type OrderFindAllOptionsFunc func(*OrderFindAllOptions) error

func GetOrderFindAllOptions(opts ...OrderFindAllOptionsFunc) (OrderFindAllOptions, error) {
	res := OrderFindAllOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func OrderWithChainId(chainId domain.ChainId) OrderFindAllOptionsFunc {
	return func(options *OrderFindAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func OrderWithOrderHash(orderHash domain.OrderHash) OrderFindAllOptionsFunc {
	return func(options *OrderFindAllOptions) error {
		options.OrderHash = &orderHash
		return nil
	}
}

func OrderWithIsAsk(isAsk bool) OrderFindAllOptionsFunc {
	return func(options *OrderFindAllOptions) error {
		options.IsAsk = &isAsk
		return nil
	}
}

func OrderWithSigner(signer domain.Address) OrderFindAllOptionsFunc {
	return func(options *OrderFindAllOptions) error {
		options.Signer = &signer
		return nil
	}
}

func OrderWithNonce(nonce string) OrderFindAllOptionsFunc {
	return func(options *OrderFindAllOptions) error {
		options.Nonce = &nonce
		return nil
	}
}

func OrderWithStartTimeGT(t time.Time) OrderFindAllOptionsFunc {
	return func(options *OrderFindAllOptions) error {
		options.StartTimeGT = &t
		return nil
	}
}

func OrderWithStartTimeLT(t time.Time) OrderFindAllOptionsFunc {
	return func(options *OrderFindAllOptions) error {
		options.StartTimeLT = &t
		return nil
	}
}

func OrderWithEndTimeGT(t time.Time) OrderFindAllOptionsFunc {
	return func(options *OrderFindAllOptions) error {
		options.EndTimeGT = &t
		return nil
	}
}

func OrderWithEndTimeLT(t time.Time) OrderFindAllOptionsFunc {
	return func(options *OrderFindAllOptions) error {
		options.EndTimeLT = &t
		return nil
	}
}

type OrderItemRepo interface {
	FindAll(ctx ctx.Ctx, opts ...OrderItemFindAllOptionsFunc) ([]*OrderItem, error)
	FindOne(ctx ctx.Ctx, itemId OrderItemId) (*OrderItem, error)
	Upsert(ctx ctx.Ctx, orderItem *OrderItem) error
	Update(ctx ctx.Ctx, itemId OrderItemId, patchalbe OrderItemPatchable) error
	RemoveAll(ctx ctx.Ctx, opts ...OrderItemFindAllOptionsFunc) error

	FindOneOrder(ctx ctx.Ctx, orderHash string) (*Order, error)
}

type OrderRepo interface {
	FindAll(ctx ctx.Ctx, opts ...OrderFindAllOptionsFunc) ([]*Order, error)
	Count(ctx ctx.Ctx, opts ...OrderFindAllOptionsFunc) (int, error)
	FindOne(ctx ctx.Ctx, id OrderId) (*Order, error)
	Upsert(ctx ctx.Ctx, order *Order) error
	Update(ctx ctx.Ctx, id OrderId, patchable OrderPatchable) error
	RemoveAll(ctx ctx.Ctx, opts ...OrderFindAllOptionsFunc) error
}

type UseCase interface {
	FindAll(ctx ctx.Ctx, opts ...OrderItemFindAllOptionsFunc) ([]*OrderItem, error)
	GetOrder(ctx ctx.Ctx, id OrderId) (*Order, error)
	MakeOrder(ctx ctx.Ctx, order Order) error
	CancelOrderItemByOrderItemHash(ctx ctx.Ctx, chainId domain.ChainId, orderItemHash domain.OrderHash, logCancelActivity bool, lMeta *domain.LogMeta) error
	CancelOrderItemByNonce(ctx ctx.Ctx, chainId domain.ChainId, signer domain.Address, nonce *big.Int, lMeta *domain.LogMeta) error
	RefreshOrders(ctx ctx.Ctx, nftitemId nftitem.Id) error
}
