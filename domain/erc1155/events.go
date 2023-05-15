package erc1155

import (
	"math/big"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type Transfer struct {
	Operator domain.Address
	From     domain.Address
	To       domain.Address
	Id       domain.TokenId
	Value    *big.Int
}

type Erc1155EventUseCase interface {
	Transfer(ctx.Ctx, domain.ChainId, *Transfer, *domain.LogMeta) error
}
