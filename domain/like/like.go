package like

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type Like struct {
	ChainId         domain.ChainId `bson:"chainId"`
	ContractAddress domain.Address `bson:"contractAddress"`
	TokenId         domain.TokenId `bson:"tokenID"`
	Liker           domain.Address `bson:"follower"`
}

func (l *Like) ToNftItemId() *nftitem.Id {
	return &nftitem.Id{
		ChainId:         l.ChainId,
		ContractAddress: l.ContractAddress,
		TokenId:         l.TokenId,
	}
}

type selectOptions struct {
	Offset            *int32           `bson:"-"`
	Limit             *int32           `bson:"-"`
	ChainId           *domain.ChainId  `bson:"chainId"`
	ContractAddress   *domain.Address  `bson:"contractAddress"`
	ContractAddresses []domain.Address `bson:"-"`
	TokenId           *domain.TokenId  `bson:"tokenID"`
	Liker             *domain.Address  `bson:"follower"`
}

type SelectOptions func(*selectOptions) error

func GetSelectOptions(opts ...SelectOptions) (selectOptions, error) {
	res := selectOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func WithPagination(offset int32, limit int32) SelectOptions {
	return func(options *selectOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func WithToken(chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) SelectOptions {
	return func(options *selectOptions) error {
		if err := WithChainId(chainId)(options); err != nil {
			return err
		}
		if err := WithContract(contract)(options); err != nil {
			return err
		}
		if err := WithTokenId(tokenId)(options); err != nil {
			return err
		}
		return nil
	}
}

func WithChainId(chainId domain.ChainId) SelectOptions {
	return func(options *selectOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func WithContract(contract domain.Address) SelectOptions {
	return func(options *selectOptions) error {
		options.ContractAddress = contract.ToLowerPtr()
		return nil
	}
}

func WithContracts(contracts []domain.Address) SelectOptions {
	return func(options *selectOptions) error {
		options.ContractAddresses = contracts
		return nil
	}
}

func WithTokenId(tokenId domain.TokenId) SelectOptions {
	return func(options *selectOptions) error {
		options.TokenId = &tokenId
		return nil
	}
}

func WithLiker(liker domain.Address) SelectOptions {
	return func(options *selectOptions) error {
		options.Liker = liker.ToLowerPtr()
		return nil
	}
}

type Repo interface {
	FindAll(c ctx.Ctx, opts ...SelectOptions) ([]*Like, error)
	Count(c ctx.Ctx, opts ...SelectOptions) (int, error)
	Create(c ctx.Ctx, value Like) error
	Delete(c ctx.Ctx, opts ...SelectOptions) error
}

type Usecase interface {
	GetLikers(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) ([]domain.Address, error)
	GetLikerCount(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) (int, error)
	GetLikeds(c ctx.Ctx, liker domain.Address) ([]*Like, error)
	GetLikedCount(c ctx.Ctx, liker domain.Address) (int, error)
	IsLiked(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId, liker domain.Address) (bool, error)
	Like(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId, liker domain.Address) (int, error)
	Unlike(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId, liker domain.Address) (int, error)
}
