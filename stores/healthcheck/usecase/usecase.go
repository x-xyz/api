package usecase

import (
	"github.com/x-xyz/goapi/base/ctx"
	hcdomain "github.com/x-xyz/goapi/domain/healthcheck"
)

type impl struct {
	repo hcdomain.HealthCheckRepo
}

// New creates new healthCheckUsecase object representation of HealthCheckUsecase interface
func New(repo hcdomain.HealthCheckRepo) hcdomain.HealthCheckUsecase {
	return &impl{
		repo: repo,
	}
}

func (im *impl) Check(context ctx.Ctx) error {
	return im.repo.PingDB(context)
}
