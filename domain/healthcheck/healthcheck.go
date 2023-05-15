package healthcheck

import (
	"github.com/x-xyz/goapi/base/ctx"
)

// HealthCheckUsecase represents the healthCheck's usecases
type HealthCheckUsecase interface {
	Check(context ctx.Ctx) error
}

// HealthCheckRepo is repository layer of healthCheck
type HealthCheckRepo interface {
	PingDB(context ctx.Ctx) error
}
