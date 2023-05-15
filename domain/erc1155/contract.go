package erc1155

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
)

type Contract struct {
	ChainId       domain.ChainId `json:"chainId" bson:"chainId"`
	Address       domain.Address `json:"address" bson:"address"`
	Name          string         `json:"name" bson:"name"`
	Symbol        string         `json:"symbol" bson:"symbol"`
	IsVerified    bool           `bson:"isVerified"`
	IsAppropriate bool           `bson:"isAppropriate"`
}

func (c *Contract) ToCollection() *collection.Collection {
	col := &collection.Collection{
		ChainId:       c.ChainId,
		Erc721Address: c.Address,
		TokenType:     domain.TokenType1155,
		IsVerified:    c.IsVerified,
		IsAppropriate: c.IsAppropriate,
		IsRegistered:  false,
	}

	if c.Name != "name" {
		col.CollectionName = c.Name
	}

	return col
}

type UpdatePayload struct {
	IsVerified    *bool `json:"isVerified" bson:"isVerified"`
	IsAppropriate *bool `json:"isAppropriate" bson:"isAppropriate"`
}

type findOptions struct {
	SortBy        *string
	SortDir       *domain.SortDir
	Offset        *int32
	Limit         *int32
	ChainId       *domain.ChainId
	Address       *domain.Address
	IsAppropriate *bool
}

type FindOptions func(*findOptions) error

func GetFindOptions(opts ...FindOptions) (findOptions, error) {
	res := findOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func WithSort(sortby string, sortdir domain.SortDir) FindOptions {
	return func(options *findOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func WithPagination(offset int32, limit int32) FindOptions {
	return func(options *findOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func WithChainId(chainId domain.ChainId) FindOptions {
	return func(options *findOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func WithAddress(address domain.Address) FindOptions {
	return func(options *findOptions) error {
		options.Address = &address
		return nil
	}
}

func WithIsAppropriate(isAppropriate bool) FindOptions {
	return func(options *findOptions) error {
		options.IsAppropriate = &isAppropriate
		return nil
	}
}

type Repo interface {
	FindAll(c ctx.Ctx, opts ...FindOptions) ([]*Contract, error)
	FindOne(c ctx.Ctx, opts ...FindOptions) (*Contract, error)
	Update(c ctx.Ctx, value UpdatePayload, opts ...FindOptions) error
	Create(c ctx.Ctx, value Contract) error
}

type UseCase interface {
	FindAll(c ctx.Ctx, opts ...FindOptions) ([]*Contract, error)
}
