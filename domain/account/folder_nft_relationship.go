package account

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type FolderNftRelationship struct {
	FolderId        string         `json:"folderId" bson:"folderId"`
	ChainId         domain.ChainId `json:"chainId" bson:"chainId"`
	ContractAddress domain.Address `json:"contractAddress" bson:"contractAddress"`
	TokenId         domain.TokenId `json:"tokenId" bson:"tokenId"`
	Index           int            `json:"index" bson:"index"`
}

func (r *FolderNftRelationship) ToNftItemId() *nftitem.Id {
	return &nftitem.Id{
		ChainId:         r.ChainId,
		ContractAddress: r.ContractAddress,
		TokenId:         r.TokenId,
	}
}

type RelationsQueryOptions struct {
	FolderId  *string
	LastIndex *int
	Limit     *int
	FolderIds *[]string
	NftitemId *nftitem.Id
}

type RelationsQueryOptionsFunc func(*RelationsQueryOptions) error

func ParseRelationsQueryOptionFunc(opts ...RelationsQueryOptionsFunc) (RelationsQueryOptions, error) {
	res := RelationsQueryOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func WithFolderId(folderId string) RelationsQueryOptionsFunc {
	return func(options *RelationsQueryOptions) error {
		options.FolderId = &folderId
		return nil
	}
}

func WithLastIndex(ind int) RelationsQueryOptionsFunc {
	return func(options *RelationsQueryOptions) error {
		options.LastIndex = &ind
		return nil
	}
}

func WithLimit(l int) RelationsQueryOptionsFunc {
	return func(options *RelationsQueryOptions) error {
		options.Limit = &l
		return nil
	}
}

func WithFolderIds(ids []string) RelationsQueryOptionsFunc {
	return func(options *RelationsQueryOptions) error {
		options.FolderIds = &ids
		return nil
	}
}

func WithNftitemId(id nftitem.Id) RelationsQueryOptionsFunc {
	return func(options *RelationsQueryOptions) error {
		options.NftitemId = &id
		return nil
	}
}

type FolderNftRelationshipRepo interface {
	Insert(ctx ctx.Ctx, relation *FolderNftRelationship) error
	GetAllRelations(ctx ctx.Ctx, opts ...RelationsQueryOptionsFunc) ([]*FolderNftRelationship, error)
	Count(ctx ctx.Ctx, opts ...RelationsQueryOptionsFunc) (int, error)
	DeleteAllRelationsByFolderID(ctx ctx.Ctx, folderID string) error
	DeleteAllRelationsByNftitem(ctx ctx.Ctx, nftID nftitem.Id) error
	DeleteAll(ctx ctx.Ctx, opts ...RelationsQueryOptionsFunc) error
	AddNftitemsToFolder(ctx ctx.Ctx, items []nftitem.Id, folderId string) error
	// MoveNftitems will move or create nft from src folder to destination folder
	MoveNftitems(ctx ctx.Ctx, items []nftitem.Id, fromFolderId, toFolderId string) error
}
