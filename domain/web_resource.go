package domain

import (
	"github.com/x-xyz/goapi/base/ctx"
)

type WebResourceReaderRepository interface {
	Get(ctx.Ctx, string) ([]byte, error)
}

type WebResourceWriterRepository interface {
	Store(ctx.Ctx, string, []byte, string) (string, error)
}

type WebResourceUseCase interface {
	Get(ctx.Ctx, string) ([]byte, error)
	GetJson(ctx.Ctx, string) ([]byte, error)
	Store(ctx.Ctx, ChainId, Address, TokenId, string, string, []byte, string) (string, error)
}
