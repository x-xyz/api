package usecase

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/erc721/contract"
)

type erc721UseCase struct {
	repo contract.Repo
}

func NewErc721UseCase(r contract.Repo) contract.UseCase {
	return &erc721UseCase{
		repo: r,
	}
}

func (u *erc721UseCase) FindAll(ctx bCtx.Ctx, optFns ...contract.FindOptions) ([]*contract.Contract, error) {
	return u.repo.FindAll(ctx, optFns...)
}
