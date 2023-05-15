package account

import (
	"time"

	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type PendingOffer struct {
	Token        nftitem.SimpleNftItem `json:"token"`
	Offeror      SimpleAccount         `json:"offeror"`
	Quantity     int32                 `json:"quantity"`
	Price        float64               `json:"price"`
	PaymentToken domain.Address        `json:"paymentToken"`
	PriceInUsd   float64               `json:"priceInUsd"`
	CreatedAt    time.Time             `json:"createdAt"`
	Deadline     float64               `json:"deadline"`
}

type PendingOfferResult struct {
	Data  []*PendingOffer `json:"data"`
	Count int
}

type findPendingOfferOptions struct {
	Offset   *int32
	Limit    *int32
	ChainId  *domain.ChainId
	Contract *domain.Address
	TokenId  *domain.TokenId
}

type FindPendingOfferOptions func(*findPendingOfferOptions) error

func GetFindPendingPfferOptions(opts ...FindPendingOfferOptions) (findPendingOfferOptions, error) {
	res := findPendingOfferOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func PendingOfferWithPagination(offset, limit int32) FindPendingOfferOptions {
	return func(opts *findPendingOfferOptions) error {
		opts.Offset = &offset
		opts.Limit = &limit
		return nil
	}
}

func PendingOfferWithChainId(chainId domain.ChainId) FindPendingOfferOptions {
	return func(opts *findPendingOfferOptions) error {
		opts.ChainId = &chainId
		return nil
	}
}

func PendingOfferWithContract(contract domain.Address) FindPendingOfferOptions {
	return func(opts *findPendingOfferOptions) error {
		opts.Contract = contract.ToLowerPtr()
		return nil
	}
}

func PendingOfferWithTokenId(tokenId domain.TokenId) FindPendingOfferOptions {
	return func(opts *findPendingOfferOptions) error {
		opts.TokenId = &tokenId
		return nil
	}
}
