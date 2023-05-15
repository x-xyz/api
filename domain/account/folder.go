package account

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type Folder struct {
	Id                    string         `json:"id" bson:"id"`
	Name                  string         `json:"name" bson:"name"`
	IsPrivate             bool           `json:"isPrivate" bson:"isPrivate"`
	IsBuiltIn             bool           `json:"isBuiltIn" bson:"isBuiltIn"` // built-in folder is uneditable
	FloorPriceInUsd       float64        `json:"floorPriceInUsd" bson:"floorPriceInUsd"`
	TotalValueInUsd       float64        `json:"totalValueInUsd" bson:"totalValueInUsd"`
	InstantLiquidityInUsd float64        `json:"instantLiquidityInUsd" bson:"instantLiquidityInUsd"`
	TotalValueMovement    float64        `json:"totalValueMovement" bson:"totalValueMovement"`
	Owner                 domain.Address `json:"owner" bson:"owner"`
	CreatedAt             time.Time      `json:"createdAt" bson:"createdAt,omitempty"`
	NftCount              int            `json:"nftCount" bson:"nftCount"`
	CollectionCount       int            `json:"collectionCount" bson:"collectionCount"`
	Cover                 nftitem.Id     `json:"cover,omitempty" bson:"cover,omitempty"`
}

type FolderUpdater struct {
	Name                  *string  `json:"name" bson:"name"`
	IsPrivate             *bool    `json:"isPrivate" bson:"isPrivate"`
	FloorPriceInUsd       *float64 `json:"floorPriceInUsd" bson:"floorPriceInUsd"`
	TotalValueInUsd       *float64 `json:"totalValueInUsd" bson:"totalValueInUsd"`
	InstantLiquidityInUsd *float64 `json:"instantLiquidityInUsd" bson:"instantLiquidityInUsd"`
	TotalValueMovement    *float64 `json:"totalValueMovement" bson:"totalValueMovement"`
	// update by usecase
	NftCount *int `json:"-" bson:"nftCount"`
	// udpate by usecase
	CollectionCount *int `json:"-" bson:"collectionCount"`
	// update by usecase, use first of nfts
	Cover *nftitem.Id `json:"cover" bson:"cover"`
}

type GetFoldersOptions struct {
	Owner     *domain.Address
	IsBuiltIn *bool
	IsPrivate *bool
	Offset    *int32
	Limit     *int32
}

func ParseGetFoldersOptionFunc(opts ...GetFoldersOptionsFunc) (GetFoldersOptions, error) {
	res := GetFoldersOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

type GetFoldersOptionsFunc func(*GetFoldersOptions) error

func WithOwner(owner domain.Address) GetFoldersOptionsFunc {
	return func(gfo *GetFoldersOptions) error {
		gfo.Owner = &owner
		return nil
	}
}

func WithBuiltIn(isBuiltIn bool) GetFoldersOptionsFunc {
	return func(gfo *GetFoldersOptions) error {
		gfo.IsBuiltIn = &isBuiltIn
		return nil
	}
}

func WithPrivate(isPrivate bool) GetFoldersOptionsFunc {
	return func(gfo *GetFoldersOptions) error {
		gfo.IsPrivate = &isPrivate
		return nil
	}
}

func WithPagination(offset int32, limit int32) GetFoldersOptionsFunc {
	return func(options *GetFoldersOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

type FolderRepo interface {
	Insert(ctx ctx.Ctx, folder *Folder) error
	Get(ctx ctx.Ctx, Id string) (*Folder, error)
	GetFolders(ctx ctx.Ctx, opts ...GetFoldersOptionsFunc) ([]*Folder, error)
	Update(ctx ctx.Ctx, Id string, updater *FolderUpdater) error
	Delete(ctx ctx.Ctx, Id string) error
}

type FolderUseCase interface {
	InitBuiltInFolders(ctx ctx.Ctx, owner domain.Address) error
	Create(ctx.Ctx, *Folder) (string, error)
	Update(ctx.Ctx, string, *FolderUpdater, []nftitem.Id) error
	GetFolder(ctx ctx.Ctx, folderId string) (*Folder, error)
	GetFolders(ctx ctx.Ctx, opts ...GetFoldersOptionsFunc) ([]*Folder, error)
	GetNFTsInFolder(ctx ctx.Ctx, folderId string) ([]*nftitem.NftitemWith1155Balance, error)
	RefreshStat(ctx.Ctx, string) error
	RefreshCount(ctx ctx.Ctx, folderId string) error
	Delete(c ctx.Ctx, folderId string) error
	MarkNftPrivate(c ctx.Ctx, owner domain.Address, marks []nftitem.Id, unmarks []nftitem.Id) error
	DeleteRelationFromAllFolders(c ctx.Ctx, owner domain.Address, nftitemId nftitem.Id) error
	AddNftToPublicFolder(c ctx.Ctx, owner domain.Address, nftitemId nftitem.Id) error
}
