package account

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type OrderNonce struct {
	Address            domain.Address `bson:"address"`
	ChainId            domain.ChainId `bson:"chainId"`
	NextAvailableNonce string         `bson:"nextAvailableNonce"`
	MinValidOrderNonce string         `bson:"minValidOrderNonce"`
}

type OrderNonceId struct {
	Address domain.Address `bson:"address"`
	ChainId domain.ChainId `bson:"chainId"`
}

func (on *OrderNonce) ToId() OrderNonceId {
	return OrderNonceId{
		Address: on.Address,
		ChainId: on.ChainId,
	}
}

type OrderNoncePatchable struct {
	NextAvailableNonce *string `bson:"nextAvailableNonce"`
	MinValidOrderNonce *string `bson:"minValidOrderNonce"`
}

type OrderNonceRepo interface {
	FindOne(ctx ctx.Ctx, id OrderNonceId) (*OrderNonce, error)
	Upsert(ctx ctx.Ctx, nonce *OrderNonce) error
	Update(ctx ctx.Ctx, id OrderNonceId, patchable OrderNoncePatchable) error
}

type OrderNonceUseCase interface {
	FindOne(ctx ctx.Ctx, id OrderNonceId) (*OrderNonce, error)
	UseAvailableNonce(ctx ctx.Ctx, id OrderNonceId) (string, error)
	UpdateMinValidOrderNonce(ctx ctx.Ctx, id OrderNonceId, nonce string) error
	UpdateAvailableNonceIfNeeded(ctx ctx.Ctx, id OrderNonceId, nonce string) error
}
