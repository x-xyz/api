package usecase

import (
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type transferEventUseCase struct {
	transferEventRepo domain.TransferEventRepo
	ctxTimeout        time.Duration
}

func NewTransferEventUseCase(r domain.TransferEventRepo, ctxTimeout time.Duration) domain.TransferEventUseCase {
	return &transferEventUseCase{
		transferEventRepo: r,
		ctxTimeout:        ctxTimeout,
	}
}
func (u *transferEventUseCase) Store(c bCtx.Ctx, e *domain.TransferEvent) error {
	ctx, cancel := bCtx.WithTimeout(c, u.ctxTimeout)
	defer cancel()
	return u.transferEventRepo.Store(ctx, e)
}
