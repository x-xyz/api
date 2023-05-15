package domain

import (
	"github.com/x-xyz/goapi/base/ctx"
)

type Id struct {
	ChainId ChainId `bson:"chainId"`
	Address Address `bson:"address"`
}

type PayToken struct {
	Name                  string  `bson:"name"`
	Symbol                string  `bson:"symbol"`
	Decimals              int32   `bson:"decimals"` // decimals for chainlink pricefeed
	TokenDecimals         int32   `bson:"tokenDecimals"`
	ChainId               ChainId `bson:"chainId"`
	Address               Address `bson:"address"`
	ChainlinkProxyAddress Address `bson:"chainlinkProxyAddress"`
	IsMainnet             bool    `bson:"isMainnet"`
	CoinGeckoId           string  `bson:"coinGeckoId"`
}

func (t *PayToken) ToId() *Id {
	return &Id{
		ChainId: t.ChainId,
		Address: t.Address,
	}
}

type PayTokenRepo interface {
	FindOne(ctx.Ctx, ChainId, Address) (*PayToken, error)
	Create(ctx.Ctx, *PayToken) error
	Upsert(ctx.Ctx, *PayToken) error
}
