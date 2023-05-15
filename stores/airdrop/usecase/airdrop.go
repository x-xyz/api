package usecase

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/airdrop"
)

type airdropUseCaseImpl struct {
	repo airdrop.AirdropRepo
}

func NewAirdropUseCase(repo airdrop.AirdropRepo) airdrop.AirdropUseCase {
	return &airdropUseCaseImpl{repo: repo}
}

func (r *airdropUseCaseImpl) FindAll(ctx bCtx.Ctx, optFns ...airdrop.AirdropFindAllOptionsFunc) ([]airdrop.Airdrop, error) {
	airdrops, err := r.repo.FindAll(ctx, optFns...)
	if err != nil {
		ctx.WithField("err", err).Error("repo.FindAll failed")
		return nil, err
	}
	return airdrops, nil
}

func (r *airdropUseCaseImpl) Create(ctx bCtx.Ctx, a *airdrop.Airdrop) error {
	if err := r.repo.Create(ctx, a); err != nil {
		ctx.WithField("err", err).Error("repo.Create failed")
		return err
	}
	return nil
}
