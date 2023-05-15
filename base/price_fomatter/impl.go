package pricefomatter

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/shopspring/decimal"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/coingecko"
)

type PriceFormatterCfg struct {
	Paytoken  domain.PayTokenRepo
	Chainlink domain.ChainlinkUsacase
	CoinGecko coingecko.Client
}

type impl struct {
	paytoken  domain.PayTokenRepo
	chainlink domain.ChainlinkUsacase
	coinGecko coingecko.Client

	// mutex protected members
	mutex         sync.Mutex
	payTokenCache map[string]*domain.PayToken
}

func NewPriceFormatter(cfg *PriceFormatterCfg) PriceFormatter {
	return &impl{
		paytoken:      cfg.Paytoken,
		chainlink:     cfg.Chainlink,
		coinGecko:     cfg.CoinGecko,
		payTokenCache: make(map[string]*domain.PayToken),
	}
}

func (f *impl) getPayToken(ctx bCtx.Ctx, chainId domain.ChainId, token domain.Address) (*domain.PayToken, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	key := fmt.Sprintf("%d%s", chainId, token)
	p, ok := f.payTokenCache[key]
	if ok {
		return p, nil
	}
	p, err := f.paytoken.FindOne(ctx, chainId, token)
	if err != nil {
		ctx.WithFields(log.Fields{
			"chainId": chainId,
			"token":   token,
			"err":     err,
		}).Error("paytoken.FindOne failed")
		return nil, err
	}
	f.payTokenCache[key] = p
	return p, nil
}

func (f *impl) formatToken(ctx bCtx.Ctx, chainId domain.ChainId, token domain.Address, value *big.Int) (decimal.Decimal, error) {
	p, err := f.getPayToken(ctx, chainId, token)
	if err != nil {
		ctx.WithFields(log.Fields{
			"chainId": chainId,
			"token":   token,
			"err":     err,
		}).Error("getPayToken failed")
		return decimal.Zero, err
	}
	return decimal.NewFromBigInt(value, -p.TokenDecimals), nil
}

// GetPrices returns displayPrice, priceInUsd, priceInNative, error
func (f *impl) GetPrices(ctx bCtx.Ctx, chainId domain.ChainId, token domain.Address, value *big.Int) (decimal.Decimal, float64, float64, error) {
	displayPrice, err := f.formatToken(ctx, chainId, token, value)
	if err != nil {
		ctx.WithField("err", err).Error("failed to parse price")
		return decimal.Zero, 0, 0, err
	}
	priceInUsd, priceInNative, err := f.GetPricesFromDisplayPrice(ctx, chainId, token, displayPrice)
	if err != nil {
		ctx.WithFields(log.Fields{
			"chainId":      chainId,
			"token":        token,
			"displayPrice": displayPrice,
			"err":          err,
		}).Error("f.GetPricesFromDisplayPrice failed")
	}
	return displayPrice, priceInUsd, priceInNative, nil
}

func (f *impl) GetPricesFromDisplayPriceString(ctx bCtx.Ctx, chainId domain.ChainId, token domain.Address, displayPriceString string) (float64, float64, error) {

	displayPrice, err := decimal.NewFromString(displayPriceString)
	if err != nil {
		ctx.WithFields(log.Fields{
			"displayPrice": displayPriceString,
			"err":          err,
		}).Error("decimal.NewFromString failed")
		return 0, 0, err
	}
	return f.GetPricesFromDisplayPrice(ctx, chainId, token, displayPrice)
}

func (f *impl) GetPricesFromDisplayPrice(ctx bCtx.Ctx, chainId domain.ChainId, token domain.Address, displayPrice decimal.Decimal) (float64, float64, error) {
	payTokenPrice, err := f.chainlink.GetLatestAnswer(ctx, chainId, token)
	if err == domain.ErrNoPriceFeed {
		// fall back to coingecko
		payToken, err := f.getPayToken(ctx, chainId, token)
		if err != nil {
			ctx.WithFields(log.Fields{
				"chainId": chainId,
				"token":   token,
				"err":     err,
			}).Error("getPayToken failed")
			return 0, 0, err
		}
		payTokenPrice, err = f.coinGecko.GetPrice(ctx, payToken.CoinGeckoId)
		if err != nil {
			ctx.WithField("err", err).Error("coinGecko.GetPrice failed")
			return 0, 0, err
		}
	} else if err != nil {
		ctx.WithField("err", err).Error("chainlink.GetLatestAnswer failed")
		return 0, 0, err
	}
	priceInUsd := displayPrice.Mul(payTokenPrice)
	if f.isNativeToken(token, chainId) {
		return priceInUsd.InexactFloat64(), displayPrice.InexactFloat64(), nil
	}
	priceInNative, err := f.usdToNative(ctx, chainId, priceInUsd)
	if err != nil {
		ctx.WithField("err", err).Error("u.usdToNative failed")
		return 0, 0, err
	}
	return priceInUsd.InexactFloat64(), priceInNative.InexactFloat64(), nil
}

func (f *impl) isNativeToken(token domain.Address, chainId domain.ChainId) bool {
	return token.Equals(zeroAddr) || token.Equals(domain.ChainIdWrappedNativeMap[chainId])
}

func (f *impl) usdToNative(ctx bCtx.Ctx, chainId domain.ChainId, usd decimal.Decimal) (decimal.Decimal, error) {
	nativeTokenPrice, err := f.chainlink.GetLatestAnswer(ctx, chainId, zeroAddr)
	if err != nil {
		ctx.WithField("err", err).Error("chainlink.GetLatestAnswer failed")
		return decimal.Zero, err
	}
	return usd.Div(nativeTokenPrice), nil
}
