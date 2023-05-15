package opensea

import (
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

var (
	ErrStatusCodeNotOk = errors.New("http.status != 200")
	ErrParseTotalPrice = errors.New("parse opensea event total price error")
)

type GetEventOptions struct {
	ContractAddress *domain.Address
	EventType       *EventType
	Before          *time.Time
	After           *time.Time
	Cursor          *string
}

type GetEventOptionsFunc func(*GetEventOptions) error

func ParseGetEventOptions(opts ...GetEventOptionsFunc) (GetEventOptions, error) {
	opt := GetEventOptions{}
	for _, f := range opts {
		err := f(&opt)
		if err != nil {
			return opt, err
		}
	}
	return opt, nil
}

func WithContractAddress(address domain.Address) GetEventOptionsFunc {
	return func(opt *GetEventOptions) error {
		opt.ContractAddress = &address
		return nil
	}
}

func WithEventType(eventType EventType) GetEventOptionsFunc {
	return func(opt *GetEventOptions) error {
		opt.EventType = &eventType
		return nil
	}
}

func WithBefore(t time.Time) GetEventOptionsFunc {
	return func(opt *GetEventOptions) error {
		opt.Before = &t
		return nil
	}
}

func WithAfter(t time.Time) GetEventOptionsFunc {
	return func(opt *GetEventOptions) error {
		opt.After = &t
		return nil
	}
}

func WithCursor(c string) GetEventOptionsFunc {
	return func(opt *GetEventOptions) error {
		opt.Cursor = &c
		return nil
	}
}

type Client interface {
	GetCollectionBySlug(bCtx.Ctx, string) (*CollectionResp, error)
	GetAssetContractByAddress(bCtx.Ctx, string) (*AssetContractResp, error)
	GetEvent(ctx bCtx.Ctx, option ...GetEventOptionsFunc) (*EventResp, error)
	GetAsset(ctx bCtx.Ctx, collectionSlug string, tokenId string) (*AssetsResp, error)
	GetAssetByOwner(ctx bCtx.Ctx, owner domain.Address, next string) (*AssetsResp, error)
}

type ClientCfg struct {
	HttpClient http.Client
	Timeout    time.Duration
	Apikey     string
}

type AssetContractResp struct {
	Collection struct {
		Description   string `json:"description"`
		Name          string `json:"name"`
		Image         string `json:"image_url"`
		Url           string `json:"external_url"`
		Discord       string `json:"discord_url"`
		Twitter       string `json:"twitter_username"`
		Instagram     string `json:"instagram_username"`
		Medium        string `json:"medium_username"`
		Telegram      string `json:"telegram_url"`
		PayoutAddress string `json:"payout_address"`
		Slug          string `json:"slug"`
	} `json:"collection"`
	Address string `json:"address"`
	Royalty uint16 `json:"dev_seller_fee_basis_points"`
	Symbol  string `json:"symbol"`
}

type CollectionResp struct {
	Collection struct {
		Name  string `json:"name"`
		Stats struct {
			OneHourVolume   float64 `json:"one_hour_volume"`
			OneHourChange   float64 `json:"one_hour_change"`
			OneHourSales    float64 `json:"one_hour_sales"`
			SixHourVolume   float64 `json:"six_hour_volume"`
			SixHourChange   float64 `json:"six_hour_change"`
			SixHourSales    float64 `json:"six_hour_sales"`
			OneDayVolume    float64 `json:"one_day_volume"`
			OneDayChange    float64 `json:"one_day_change"`
			OneDaySales     float64 `json:"one_day_sales"`
			SevenDayVolume  float64 `json:"seven_day_volume"`
			SevenDayChange  float64 `json:"seven_day_change"`
			SevenDaySales   float64 `json:"seven_day_sales"`
			ThirtyDayVolume float64 `json:"thirty_day_volume"`
			ThirtyDayChange float64 `json:"thirty_day_change"`
			ThirtyDaySales  float64 `json:"thirty_day_sales"`
			TotalVolume     float64 `json:"total_volume"`
			TotalSales      float64 `json:"total_sales"`
			FloorPrice      float64 `json:"floor_price"`
		} `json:"stats"`
	} `json:"collection"`
}

type EventResp struct {
	Next        string       `json:"next"`
	AssetEvents []AssetEvent `json:"asset_events"`
}

type AssetsResp struct {
	Next     string  `json:"next"`
	Assets   []Asset `json:"assets"`
	Previous string  `json:"previous"`
}

type OpenseaListingMaker struct {
	Address domain.Address `json:"address"`
}

type OpenseaListingTaker struct {
	Address domain.Address `json:"address"`
}

type OpenseaListingMetadata struct {
	Asset OpenseaListingAsset `json:"asset"`
}

type OpenseaListingAsset struct {
	TokenId domain.TokenId `json:"id"`
	Address domain.Address `json:"address"`
}

type AssetContract struct {
	Address domain.Address `json:"address"`
}

type PaymentTokenContract struct {
	Symbol   string         `json:"symbol"`
	Address  domain.Address `json:"address"`
	Decimals int32          `json:"decimals"`
	EthPrice string         `json:"eth_price"`
	UsdPrice string         `json:"usd_price"`
}

type OpenseaListingProtocolData struct {
	Parameters OpenseaListingParameters `json:"parameters"`
}

type OpenseaListingParameters struct {
	Offer         []OpenseaListingOffer  `json:"offer"`
	Consideration []OpenseaConsideration `json:"consideration"`
}

type OpenseaListingOffer struct {
	ItemType        int            `json:"itemType"`
	ContractAddress domain.Address `json:"token"`
	StartAmount     string         `json:"startAmount"`
	EndAmount       string         `json:"endAmount"`
}

type OpenseaConsideration struct {
	ItemType        int            `json:"itemType"`
	ContractAddress domain.Address `json:"token"`
	StartAmount     string         `json:"startAmount"`
	EndAmount       string         `json:"endAmount"`
}

type SellOrder struct {
	OrderHash            string                 `json:"order_hash"`
	StartTime            string                 `json:"created_date"`
	Deadline             string                 `json:"closing_date"`
	ExpirationTime       int64                  `json:"expiration_time"`
	ListingTime          int64                  `json:"listing_time"`
	CurrentPrice         string                 `json:"current_price"`
	Metadata             OpenseaListingMetadata `json:"metadata"`
	Maker                OpenseaListingMaker    `json:"maker"`
	Taker                OpenseaListingTaker    `json:"taker"`
	PaymentToken         string                 `json:"payment_token"`
	PaymentTokenContract PaymentTokenContract   `json:"payment_token_contract"`
	Quantity             string                 `json:"quantity"`
}

type SeaportSellOrder struct {
	OrderHash      string                     `json:"order_hash"`
	StartTime      string                     `json:"created_date"`
	Deadline       string                     `json:"closing_date"`
	ExpirationTime int64                      `json:"expiration_time"`
	ListingTime    int64                      `json:"listing_time"`
	CurrentPrice   string                     `json:"current_price"`
	ProtocolData   OpenseaListingProtocolData `json:"protocol_data"`
	Maker          OpenseaListingMaker        `json:"maker"`
	Taker          OpenseaListingTaker        `json:"taker"`
	TokenId        domain.TokenId             `json:"token_id"`
}

type Asset struct {
	TokenId           domain.TokenId     `json:"token_id"`
	Owner             Owner              `json:"owner"`
	AssetContract     AssetContract      `json:"asset_contract"`
	SellOrders        []SellOrder        `json:"sell_orders"`
	SeaportSellOrders []SeaportSellOrder `json:"seaport_sell_orders"`
}

type Owner struct {
	Address domain.Address `json:"address"`
}

type Listing struct {
	TokenId           domain.TokenId `json:"token_id"`
	AssetContract     AssetContract  `json:"asset_contract"`
	SellOrders        []SellOrder    `json:"sell_orders"`
	SeaportSellOrders []SellOrder    `json:"seaport_sell_orders"`
}

type Account struct {
	Address domain.Address `json:"address"`
}

type Transaction struct {
	TransactionHash domain.TxHash     `json:"transaction_hash"`
	FromAccount     Account           `json:"from_account"`
	ToAccount       Account           `json:"to_account"`
	BlockNumber     *string           `json:"block_number"`
	BlockHash       *domain.BlockHash `json:"block_hash"`
}

type PaymentToken struct {
	Address  domain.Address `json:"address"`
	UsdPrice string         `json:"usd_price"`
	EthPrice string         `json:"eth_price"`
	Decimals int            `json:"decimals"`
}

type EventType string

const (
	EventTypeCreated      EventType = "created"
	EventTypeSuccessful   EventType = "successful"
	EventTypeCancelled    EventType = "cancelled"
	EventTypeBidEntered   EventType = "bid_entered"
	EventTypeBidWithdrawn EventType = "bid_withdrawn"
	EventTypeTransfer     EventType = "transfer"
	EventTypeApprove      EventType = "approve"
)

type AssetEvent struct {
	Id    int64  `json:"id"`
	Asset *Asset `json:"asset"`
	// NOTE: This ContractAddress is not collection contract address
	ContractAddress domain.Address `json:"contract_address"`
	EventType       EventType      `json:"event_type"`
	Transaction     Transaction    `json:"transaction"`
	CreatedDate     string         `json:"created_date"`
	TotalPrice      string         `json:"total_price"`
	Quantity        string         `json:"quantity"`
	PaymentToken    PaymentToken   `json:"payment_token"`
	Seller          Account        `json:"account"`
	WinnerAccount   Account        `json:"winner_account"`
}

// GetPrices returns displayPrice, priceInUsd, priceInNative
func (ev AssetEvent) GetPrices() (decimal.Decimal, float64, float64, error) {
	n := new(big.Int)
	n, ok := n.SetString(ev.TotalPrice, 10)
	if !ok {
		return decimal.Zero, 0, 0, ErrParseTotalPrice
	}

	displayPrice := decimal.NewFromBigInt(n, -int32(ev.PaymentToken.Decimals))
	usdPrice, err := strconv.ParseFloat(ev.PaymentToken.UsdPrice, 64)
	if err != nil {
		return decimal.Zero, 0, 0, err
	}

	ethPrice, err := strconv.ParseFloat(ev.PaymentToken.EthPrice, 64)
	if err != nil {
		return decimal.Zero, 0, 0, err
	}

	return displayPrice, displayPrice.InexactFloat64() * usdPrice, displayPrice.InexactFloat64() * ethPrice, nil
}
