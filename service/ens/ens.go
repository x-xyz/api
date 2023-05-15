package ens

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type ENS interface {
	Resolve(ctx ctx.Ctx, name string) (domain.Address, error)
	ReverseResolve(ctx ctx.Ctx, address domain.Address) (string, error)
}
