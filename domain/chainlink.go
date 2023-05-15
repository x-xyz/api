package domain

import (
	"errors"

	"github.com/shopspring/decimal"
	"github.com/x-xyz/goapi/base/ctx"
)

var ErrNoPriceFeed = errors.New("no price feed")

type ChainlinkUsacase interface {
	GetLatestAnswer(c ctx.Ctx, chain ChainId, token Address) (decimal.Decimal, error)
	GetLatestAnswerAt(c ctx.Ctx, chain ChainId, token Address, blk uint64) (decimal.Decimal, error)
}
