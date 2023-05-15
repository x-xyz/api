package chainlink

import (
	"math/big"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type Chainlink interface {
	GetLatestAnswer(c ctx.Ctx, chainId domain.ChainId, feedAddress domain.Address) (*big.Int, error)
	GetLatestAnswerAt(c ctx.Ctx, chainId domain.ChainId, feedAddress domain.Address, blk *big.Int) (*big.Int, error)
}
