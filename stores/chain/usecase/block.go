package usecase

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/chain"
)

type blockUseCase struct {
	repo chain.BlockRepo
}

func NewBlockUseCase(r chain.BlockRepo) chain.BlockUseCase {
	return &blockUseCase{repo: r}
}

func (u *blockUseCase) Create(ctx bCtx.Ctx, b *chain.Block) error {
	return u.repo.Create(ctx, b)
}

func (u *blockUseCase) Upsert(ctx bCtx.Ctx, b *chain.Block) error {
	return u.repo.Upsert(ctx, b)
}

func (u *blockUseCase) FindOne(ctx bCtx.Ctx, b *chain.BlockId) (*chain.Block, error) {
	return u.repo.FindOne(ctx, b)
}
