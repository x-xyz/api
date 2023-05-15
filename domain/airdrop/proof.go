package airdrop

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type Proof struct {
	ChainId         domain.ChainId `json:"chainId" bson:"chainId"`
	ContractAddress domain.Address `json:"contractAddress" bson:"contractAddress"`
	Claimer         domain.Address `json:"claimer" bson:"claimer"`
	Round           int            `json:"round" bson:"round"`
	Amount          string         `json:"amount" bson:"amount"`
	Proof           []string       `json:"proof" bson:"proof"`
}

type ProofFindAllOptions struct {
	SortBy          *string         `bson:"-"`
	SortDir         *domain.SortDir `bson:"-"`
	Offset          *int32          `bson:"-"`
	Limit           *int32          `bson:"-"`
	ChainId         *domain.ChainId `bson:"chainId"`
	ContractAddress *domain.Address `bson:"contractAddress"`
	Claimer         *domain.Address `bson:"claimer"`
}

type ProofFindAllOptionsFunc func(*ProofFindAllOptions) error

func GetProofFindAllOptions(opts ...ProofFindAllOptionsFunc) (ProofFindAllOptions, error) {
	res := ProofFindAllOptions{}
	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}
	return res, nil
}

func ProofWithSort(sortby string, sortdir domain.SortDir) ProofFindAllOptionsFunc {
	return func(options *ProofFindAllOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func ProofWithPagination(offset int32, limit int32) ProofFindAllOptionsFunc {
	return func(options *ProofFindAllOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func ProofWithChainId(chainId domain.ChainId) ProofFindAllOptionsFunc {
	return func(options *ProofFindAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func ProofWithContractAddress(address domain.Address) ProofFindAllOptionsFunc {
	return func(options *ProofFindAllOptions) error {
		options.ContractAddress = address.ToLowerPtr()
		return nil
	}
}

func ProofWithClaimer(claimer domain.Address) ProofFindAllOptionsFunc {
	return func(options *ProofFindAllOptions) error {
		options.Claimer = claimer.ToLowerPtr()
		return nil
	}
}

type ProofRepo interface {
	FindAll(ctx.Ctx, ...ProofFindAllOptionsFunc) ([]Proof, error)
	Create(ctx.Ctx, *Proof) error
}

type ProofUseCase interface {
	FindAll(ctx.Ctx, ...ProofFindAllOptionsFunc) ([]Proof, error)
	Create(ctx.Ctx, *Proof) error
}
