package external_listing

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type ExternalListing struct {
	Owner           domain.Address `json:"owner" bson:"owner"`
	ChainId         domain.ChainId `json:"chainId" bson:"chainId"`
	Minter          domain.Address `json:"minter" bson:"minter"`
	ContractAddress domain.Address `json:"contractAddress" bson:"contractAddress"`
	TokenId         domain.TokenId `json:"tokenId" bson:"tokenId"`
	Quantity        int64          `json:"quantity" bson:"quantity"`
	PaymentToken    domain.Address `json:"paymentToken" bson:"paymentToken"`
	Price           string         `json:"price" bson:"price"`
	PriceInUsd      string         `json:"priceInUsd" bson:"priceInUSD"`
	StartTime       time.Time      `json:"startTime" bson:"startTime"`
	Deadline        time.Time      `json:"deadline" bson:"deadline"`
	Source          string         `json:"source" bson:"source"`
	UpdatedTime     time.Time      `json:"updatedTime" bson:"updatedTime"`
}

type ExternalListingId struct {
	Owner           domain.Address
	ChainId         domain.ChainId
	ContractAddress domain.Address
	TokenId         domain.TokenId
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

type RemoveAllOptions struct {
	Owner   *domain.Address `bson:"owner"`
	ChainId *domain.ChainId `bson:"chainId"`
}

type RemoveAllOptionsFunc func(*RemoveAllOptions) error

func GetRemoveAllOptions(opts ...RemoveAllOptionsFunc) (RemoveAllOptions, error) {
	res := RemoveAllOptions{}
	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}
	return res, nil
}

func WithExternalListing(owner domain.Address, chainId domain.ChainId) RemoveAllOptionsFunc {
	return func(options *RemoveAllOptions) error {
		options.Owner = &owner
		options.ChainId = &chainId
		return nil
	}
}

func WithOwner(owner domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Owner = &owner
		return nil
	}
}

func WithChainId(chainId domain.ChainId) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

type FindAllOptions struct {
	SortBy  *string
	SortDir *domain.SortDir
	Offset  *int32
	Limit   *int32
	ChainId *domain.ChainId `bson:"chainId"`
	Owner   *domain.Address
}

type ExternalListingRepo interface {
	FindAll(c ctx.Ctx, opts ...FindAllOptionsFunc) ([]ExternalListing, error)
	BulkUpsert(ctx.Ctx, []ExternalListing) error
	RemoveAll(c ctx.Ctx, opts ...RemoveAllOptionsFunc) error
}

type ExternalListingUseCase interface {
	GetListings(c ctx.Ctx, account domain.Address, chainId domain.ChainId) ([]ExternalListing, error)
	FetchOpenseaListings(c ctx.Ctx, account domain.Address, chainId domain.ChainId) ([]ExternalListing, error)
	BulkUpsert(ctx.Ctx, []ExternalListing) error
	DeleteListing(ctx.Ctx, ExternalListingId) error
}
