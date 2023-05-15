package punk

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type Transfer struct {
	From    domain.Address
	To      domain.Address
	TokenId domain.TokenId
}

type PunkEventUseCase interface {
	Transfer(ctx.Ctx, domain.ChainId, *Transfer, *domain.LogMeta) error
}
