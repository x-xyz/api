package pricefomatter

import (
	"math/big"

	"github.com/shopspring/decimal"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

const zeroAddr = domain.Address("0x0000000000000000000000000000000000000000")

type PriceFormatter interface {
	GetPrices(ctx bCtx.Ctx, chainId domain.ChainId, token domain.Address, value *big.Int) (decimal.Decimal, float64, float64, error)
	GetPricesFromDisplayPriceString(ctx bCtx.Ctx, chainId domain.ChainId, token domain.Address, displayPriceString string) (float64, float64, error)
	GetPricesFromDisplayPrice(ctx bCtx.Ctx, chainId domain.ChainId, token domain.Address, displayPrice decimal.Decimal) (float64, float64, error)
}
