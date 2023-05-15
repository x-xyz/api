package usecase

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
)

type vexUseCaseImpl struct {
	repo domain.VexFeeDistributionHistoryRepo
}

func NewVexFeeDistrubutionHistoryUseCase(repo domain.VexFeeDistributionHistoryRepo) domain.VexFeeDistributionHistoryUseCase {
	return &vexUseCaseImpl{repo: repo}
}

func (r *vexUseCaseImpl) Create(ctx bCtx.Ctx, h *domain.VexFeeDistributionHistory) error {
	if err := r.repo.Create(ctx, h); err != nil {
		ctx.WithFields(log.Fields{
			"history": h,
			"err":     err,
		}).Error("repo.Create failed")
		return err
	}
	return nil
}

func (r *vexUseCaseImpl) LatestApr(ctx bCtx.Ctx, latest int) (float64, error) {
	history, err := r.repo.FindLatest(ctx, latest)
	if err != nil {
		ctx.WithFields(log.Fields{
			"latest": latest,
			"err":    err,
		}).Error("repo.FindLatest failed")
		return 0, err
	}
	sum := float64(0)
	for _, h := range history {
		sum += h.Apr
	}
	average := sum / float64(len(history))
	return average, nil
}
