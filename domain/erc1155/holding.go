package erc1155

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type Holding struct {
	ChainId domain.ChainId `json:"chainId" bson:"chainId"`
	Address domain.Address `json:"address" bson:"address"`
	TokenId domain.TokenId `json:"tokenId" bson:"tokenId"`
	Owner   domain.Address `json:"owner" bson:"owner"`
	Balance int64          `json:"balance" bson:"balance"`
}

type HoldingId struct {
	ChainId domain.ChainId `bson:"chainId"`
	Address domain.Address `bson:"address"`
	TokenId domain.TokenId `bson:"tokenId"`
	Owner   domain.Address `bson:"owner"`
}

type FindAllOptions struct {
	Owner     *domain.Address
	Address   *domain.Address
	NftitemId *nftitem.Id
}

type FindAllOptionsFunc func(*FindAllOptions) error

func GetFindAllOptions(opts ...FindAllOptionsFunc) (FindAllOptions, error) {
	res := FindAllOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func WithOwner(owner domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Owner = &owner
		return nil
	}
}

func WithHoldingAddress(address domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Address = address.ToLowerPtr()
		return nil
	}
}

func WithNftitemId(id nftitem.Id) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.NftitemId = &id
		return nil
	}
}

type HoldingRepo interface {
	FindOne(c ctx.Ctx, id HoldingId) (*Holding, error)
	FindAll(c ctx.Ctx, opts ...FindAllOptionsFunc) ([]*Holding, error)
	Create(c ctx.Ctx, value Holding) error
	Delete(c ctx.Ctx, id HoldingId) error
	Increment(c ctx.Ctx, id HoldingId, value int64) (*Holding, error)
	CountUniqueOwner(c ctx.Ctx, chainId domain.ChainId, address domain.Address) (int64, error)
}
