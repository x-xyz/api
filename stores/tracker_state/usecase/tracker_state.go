package usecase

import (
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type trackerStateUseCase struct {
	trackerStatusRepo domain.TrackerStateRepo
	ctxTimeout        time.Duration
}

func NewTrackerStateUseCase(r domain.TrackerStateRepo, ctxTimeout time.Duration) domain.TrackerStateUseCase {
	return &trackerStateUseCase{
		trackerStatusRepo: r,
		ctxTimeout:        ctxTimeout,
	}
}

func (u *trackerStateUseCase) Get(c bCtx.Ctx, id *domain.TrackerStateId) (*domain.TrackerState, error) {
	ctx, cancel := bCtx.WithTimeout(c, u.ctxTimeout)
	defer cancel()
	return u.trackerStatusRepo.Get(ctx, id)
}

func (u *trackerStateUseCase) Update(c bCtx.Ctx, status *domain.TrackerState) error {
	ctx, cancel := bCtx.WithTimeout(c, u.ctxTimeout)
	defer cancel()
	return u.trackerStatusRepo.Update(ctx, status)
}

func (u *trackerStateUseCase) Store(c bCtx.Ctx, status *domain.TrackerState) error {
	ctx, cancel := bCtx.WithTimeout(c, u.ctxTimeout)
	defer cancel()
	return u.trackerStatusRepo.Store(ctx, status)
}
