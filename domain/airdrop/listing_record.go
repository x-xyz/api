package airdrop

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type ListingRecord struct {
	Owner           domain.Address `bson:"owner"`
	ChainId         domain.ChainId `bson:"chainId"`
	ContractAddress domain.Address `bson:"contractAddress"`
	Count           int            `bson:"count"`
	SubCount        int            `bson:"subcount"`
	SnapshotTime    *time.Time     `bson:"snapshotTime"`
	UpdatedAt       *time.Time     `bson:"updatedAt"`
}

type ListingRecordId struct {
	Owner           domain.Address `bson:"owner"`
	ChainId         domain.ChainId `bson:"chainId"`
	ContractAddress domain.Address `bson:"contractAddress"`
	SnapshotTime    *time.Time     `bson:"snapshotTime"`
}

func (r *ListingRecord) ToId() ListingRecordId {
	return ListingRecordId{
		Owner:           r.Owner,
		ChainId:         r.ChainId,
		ContractAddress: r.ContractAddress,
		SnapshotTime:    r.SnapshotTime,
	}
}

type ListingRecordFindAllOptions struct {
	SortBy          *string         `bson:"-"`
	SortDir         *domain.SortDir `bson:"-"`
	Offset          *int32          `bson:"-"`
	Limit           *int32          `bson:"-"`
	ChainId         *domain.ChainId `bson:"chainId"`
	ContractAddress *domain.Address `bson:"contractAddress"`
	Owner           *domain.Address `bson:"owner"`
	SnapshotTimeLT  *time.Time      `bson:"-"`
	SnapshotTimeGTE *time.Time      `bson:"-"`
}

type ListingRecordFindAllOptionsFunc func(*ListingRecordFindAllOptions) error

func GetListingRecordFindAllOptions(opts ...ListingRecordFindAllOptionsFunc) (ListingRecordFindAllOptions, error) {
	res := ListingRecordFindAllOptions{}
	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}
	return res, nil
}

func ListingRecordWithSort(sortby string, sortdir domain.SortDir) ListingRecordFindAllOptionsFunc {
	return func(options *ListingRecordFindAllOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func ListingRecordWithPagination(offset int32, limit int32) ListingRecordFindAllOptionsFunc {
	return func(options *ListingRecordFindAllOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func ListingRecordWithChainId(chainId domain.ChainId) ListingRecordFindAllOptionsFunc {
	return func(options *ListingRecordFindAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func ListingRecordWithContractAddress(address domain.Address) ListingRecordFindAllOptionsFunc {
	return func(options *ListingRecordFindAllOptions) error {
		options.ContractAddress = address.ToLowerPtr()
		return nil
	}
}

func ListingRecordWithOwner(address domain.Address) ListingRecordFindAllOptionsFunc {
	return func(options *ListingRecordFindAllOptions) error {
		options.Owner = address.ToLowerPtr()
		return nil
	}
}

func ListingRecordWithSnapshotTime(begin time.Time, end time.Time) ListingRecordFindAllOptionsFunc {
	return func(options *ListingRecordFindAllOptions) error {
		options.SnapshotTimeGTE = &begin
		options.SnapshotTimeLT = &end
		return nil
	}
}

type ListingRecordRepo interface {
	Upsert(ctx.Ctx, *ListingRecord) error
	FindAll(ctx.Ctx, ...ListingRecordFindAllOptionsFunc) ([]ListingRecord, error)
}

type ListingRecordUseCase interface {
	SnapshotCollectionListings(ctx.Ctx, domain.ChainId, domain.Address, time.Time) error
}
