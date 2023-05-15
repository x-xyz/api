package usecase

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type openseaDataUseCase struct {
	repo domain.OpenseaDataRepo
}

func NewOpenseaUseCase(repo domain.OpenseaDataRepo) domain.OpenseaDataUseCase {
	return &openseaDataUseCase{repo: repo}
}

func (u *openseaDataUseCase) FindOne(ctx bCtx.Ctx, id domain.OpenseaDataId) (*domain.OpenseaData, error) {
	return u.repo.FindOne(ctx, id)
}

func (u *openseaDataUseCase) Upsert(ctx bCtx.Ctx, v domain.OpenseaData) error {
	return u.repo.Upsert(ctx, v)
}
