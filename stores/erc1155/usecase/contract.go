package usecase

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/erc1155"
)

type erc1155UseCase struct {
	repo erc1155.Repo
}

func NewErc1155UseCase(r erc1155.Repo) erc1155.UseCase {
	return &erc1155UseCase{
		repo: r,
	}
}

func (u *erc1155UseCase) FindAll(ctx bCtx.Ctx, optFns ...erc1155.FindOptions) ([]*erc1155.Contract, error) {
	return u.repo.FindAll(ctx, optFns...)
}
