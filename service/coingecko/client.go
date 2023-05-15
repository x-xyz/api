package coingecko

import (
	"errors"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

var (
	ErrStatusCodeNotOk = errors.New("http.status != 200")
	ErrMarketsLen      = errors.New("len(markets) != 1")
)

type Client interface {
	GetPrice(bCtx.Ctx, string) (decimal.Decimal, error)
	// GetPriceAtDate returns token price on certain date
	// example:
	//   id(coingecko token id, ex: ethereum)
	//	 date(format: dd-mm-yyyy)
	GetPriceAtDate(ctx bCtx.Ctx, id string, date string) (decimal.Decimal, error)
}

type ClientCfg struct {
	HttpClient http.Client
	Timeout    time.Duration
}

type Markets []Market

type Market struct {
	Id           string  `json:"id"`
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Image        string  `json:"image"`
	CurrentPrice float64 `json:"current_price"`
}

type PriceData struct {
	Usd float64 `json:"usd"`
}

type MarketData struct {
	CurrentPrice PriceData `json:"current_price"`
}

type History struct {
	Id         string     `json:"id"`
	Symbol     string     `json:"symbol"`
	Name       string     `json:"name"`
	MarketData MarketData `json:"market_data"`
}
