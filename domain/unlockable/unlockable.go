package unlockable

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type UnlockableId struct {
	ChainId         domain.ChainId `json:"chainId" bson:"chainId"`
	ContractAddress domain.Address `json:"contractAddress" bson:"contractAddress"`
	TokenId         domain.TokenId `json:"tokenId" bson:"tokenID"`
}

type Unlockable struct {
	ChainId         domain.ChainId `json:"chainId" bson:"chainId"`
	ContractAddress domain.Address `json:"contractAddress" bson:"contractAddress"`
	TokenId         domain.TokenId `json:"tokenId" bson:"tokenID"`
	Content         string         `json:"content" bson:"content"`
}

type Repo interface {
	FindOne(c ctx.Ctx, id UnlockableId) (*Unlockable, error)
	Create(c ctx.Ctx, val Unlockable) error
}
