package like

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
)

type CollectionLike struct {
	ChainId         domain.ChainId `bson:"chainId"`
	ContractAddress domain.Address `bson:"contractAddress"`
	Liker           domain.Address `bson:"follower"`
}

type CollectionLikeRepo interface {
	FindAll(c ctx.Ctx, opts ...SelectOptions) ([]*CollectionLike, error)
	Count(c ctx.Ctx, opts ...SelectOptions) (int, error)
	Create(c ctx.Ctx, value CollectionLike) error
	Delete(c ctx.Ctx, opts ...SelectOptions) error
}

type CollectionLikeUsecase interface {
	GetLikers(c ctx.Ctx, id collection.CollectionId) ([]domain.Address, error)
	GetLikerCount(c ctx.Ctx, id collection.CollectionId) (int, error)
	GetLikeds(c ctx.Ctx, liker domain.Address) ([]*CollectionLike, error)
	GetLikedCount(c ctx.Ctx, liker domain.Address) (int, error)
	IsLiked(c ctx.Ctx, id collection.CollectionId, liker domain.Address) (bool, error)
	Like(c ctx.Ctx, id collection.CollectionId, liker domain.Address) (int, error)
	Unlike(c ctx.Ctx, id collection.CollectionId, liker domain.Address) (int, error)
}
