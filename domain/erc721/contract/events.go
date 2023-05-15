package contract

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type TransferEvent struct {
	From    domain.Address
	To      domain.Address
	TokenId domain.TokenId
}

type Erc721EventUseCase interface {
	Transfer(ctx.Ctx, domain.ChainId, *TransferEvent, *domain.LogMeta) error
}
