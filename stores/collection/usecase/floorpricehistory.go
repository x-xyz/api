package usecase

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/collection"
)

type floorPriceHistoryUseCase struct {
	repo collection.FloorPriceHistoryRepo
}

func NewFloorPriceHistoryUseCase(repo collection.FloorPriceHistoryRepo) collection.FloorPriceHistoryUseCase {
	return &floorPriceHistoryUseCase{repo: repo}
}

func (u *floorPriceHistoryUseCase) FindOne(ctx bCtx.Ctx, id collection.FloorPriceId) (*collection.FloorPriceHistory, error) {
	id.Date = id.Date.Truncate(day)
	return u.repo.FindOne(ctx, id)
}

func (u *floorPriceHistoryUseCase) Upsert(ctx bCtx.Ctx, h collection.FloorPriceHistory) error {
	h.Date = h.Date.Truncate(day)
	return u.repo.Upsert(ctx, h)
}
