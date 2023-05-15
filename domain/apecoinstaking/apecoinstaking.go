package apecoinstaking

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type ApecoinStaking struct {
	ChainId         domain.ChainId `json:"chainId" bson:"chainId"`
	ContractAddress domain.Address `json:"contractAddress" bson:"contractAddress"`
	TokenId         domain.TokenId `json:"tokenId" bson:"tokenId"`
	Staked          bool           `json:"staked" bson:"staked"`
}

type Id struct {
	ChainId         domain.ChainId `json:"chainId" bson:"chainId"`
	ContractAddress domain.Address `json:"contractAddress" bson:"contractAddress"`
	TokenId         domain.TokenId `json:"tokenId" bson:"tokenId"`
}

func (s *ApecoinStaking) ToId() Id {
	return Id{
		ChainId:         s.ChainId,
		ContractAddress: s.ContractAddress,
		TokenId:         s.TokenId,
	}
}

type Repo interface {
	FindOne(ctx bCtx.Ctx, id Id) (*ApecoinStaking, error)
	Upsert(ctx bCtx.Ctx, s *ApecoinStaking) error
}

type UseCase interface {
	Get(ctx bCtx.Ctx, id Id) (*ApecoinStaking, error)
	Upsert(ctx bCtx.Ctx, s *ApecoinStaking) error
}
