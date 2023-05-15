package ip

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

var YugaLabIpCollectionAddresses = []domain.Address{
	"0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
	"0x60e4d786628fea6478f785a6d7e704777c86a7c6",
	"0xba30e5f9bb24caa003e9f2f0497ad287fdf95623",
	"0x7bd29408f11d2bfc23c34f18275bbf23bb716bc7",
	"0xe785e82358879f061bc3dcac6f0444462d4b5330",
	"0xf61f24c2d93bf2de187546b14425bf631f28d6dc",
	"0xb47e3cd837ddf8e4c57f05d70ab865de6e193bbb",
}

type IPListing struct {
	ID              string           `json:"id" bson:"id"`
	Username        string           `json:"username" bson:"username"`
	Twitter         string           `json:"twitter" bson:"twitter" validate:"required_without=ContactEmail"`
	ContactEmail    string           `json:"contactEmail" bson:"contactEmail" validate:"required_without=Twitter,omitempty,email"`
	ChainId         domain.ChainId   `json:"chainId" bson:"chainId" validate:"required_if=IsIpOwner true"`
	ContractAddress domain.Address   `json:"contractAddress" bson:"contractAddress" validate:"required_if=IsIpOwner true"`
	TokenIds        []domain.TokenId `json:"tokenIds" bson:"tokenIds" validate:"required_if=IsIpOwner true"`
	Title           string           `json:"title" bson:"title" validate:"required"`
	ListingDetail   string           `json:"listingDetail" bson:"listingDetail" validate:"required"`
	LicensingPeriod string           `json:"licensingPeriod" bson:"licensingPeriod"`
	LicensingFee    string           `json:"licensingFee" bson:"licensingFee"`
	Exclusivity     string           `json:"exclusivity" bson:"exclusivity"`
	IsIpOwner       bool             `json:"isIpOwner" bson:"isIpOwner"`
	Owner           domain.Address   `json:"owner" bson:"owner"`
	CreatedAt       time.Time        `json:"createdAt" bson:"createdAt"`
}

type FindAllOptions struct {
	IsIpOwner         *bool
	ChainId           *domain.ChainId
	ContractAddresses []domain.Address
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

func WithIsIpOwner(isIpOwner bool) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.IsIpOwner = &isIpOwner
		return nil
	}
}

func WithChainId(chainId domain.ChainId) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func WithContractAddresses(addresses []domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		for _, address := range addresses {
			options.ContractAddresses = append(options.ContractAddresses, address.ToLower())
		}
		return nil
	}
}

type UseCase interface {
	Insert(ctx ctx.Ctx, listing *IPListing) error
	FindAll(ctx ctx.Ctx, opts ...FindAllOptionsFunc) ([]*IPListing, error)
	FindOne(ctx ctx.Ctx, id string) (*IPListing, error)
	Delete(ctx ctx.Ctx, id string) error
}

type Repo interface {
	Insert(ctx ctx.Ctx, listing *IPListing) error
	FindAll(ctx ctx.Ctx, opts ...FindAllOptionsFunc) ([]*IPListing, error)
	FindOne(ctx ctx.Ctx, id string) (*IPListing, error)
	Delete(ctx ctx.Ctx, id string) error
}
