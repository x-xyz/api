package domain

import (
	"encoding/json"

	"github.com/x-xyz/goapi/base/ctx"
)

type Metadata struct {
	json.RawMessage
}

type MetadataUseCase interface {
	GetFromUrl(ctx.Ctx, string) (*Metadata, error)
	Store(ctx.Ctx, ChainId, string, int32, *Metadata) (string, error)
}
