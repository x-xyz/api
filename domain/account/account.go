package account

import (
	"errors"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

// Account is user's account stored in database
type Account struct {
	Address       domain.Address `bson:"address"`
	Alias         string         `bson:"alias"`
	Email         string         `bson:"email"`
	Bio           string         `bson:"bio"`
	ImageHash     string         `bson:"imageHash"`
	BannerHash    string         `bson:"bannerHash"`
	Nonce         int32          `bson:"nonce"`
	IsAppropriate bool           `bson:"isAppropriate"`
	CreatedAt     time.Time      `bson:"createdAt,omitempty"`
	UpdatedAt     time.Time      `bson:"updatedAt,omitempty"`
	Website       string         `bson:"website"`
	Twitter       string         `bson:"twitter"`
	Instagram     string         `bson:"instagram"`
	Discord       string         `bson:"discord"`
}

func (a *Account) ToSimpleAccount() *SimpleAccount {
	return &SimpleAccount{
		Address:   a.Address,
		Alias:     a.Alias,
		ImageHash: a.ImageHash,
	}
}

func unixMilli(t time.Time) int64 {
	return t.Unix()*1e3 + int64(t.Nanosecond())/1e6
}

func (a *Account) ToInfo() *Info {
	return &Info{
		Address:     a.Address,
		Alias:       a.Alias,
		Email:       a.Email,
		Bio:         a.Bio,
		ImageHash:   a.ImageHash,
		BannerHash:  a.BannerHash,
		CreatedAtMs: unixMilli(a.CreatedAt),
		UpdatedAtMs: unixMilli(a.UpdatedAt),
		Followers:   0,
		Followings:  0,
		Website:     a.Website,
		Twitter:     a.Twitter,
		Instagram:   a.Instagram,
		Discord:     a.Discord,
	}
}

type SimpleAccount struct {
	Address   domain.Address `json:"address"`
	Alias     string         `json:"alias"`
	ImageHash string         `json:"imageHash"`
}

// Info is account struct returns to client which contains public info and aggreates data from other usecases
type Info struct {
	Address     domain.Address `json:"address"`
	Alias       string         `json:"alias"`
	Email       string         `json:"email"`
	Bio         string         `json:"bio"`
	ImageHash   string         `json:"imageHash"`
	BannerHash  string         `json:"bannerHash"`
	CreatedAtMs int64          `json:"createdAtMs,omitempty"`
	UpdatedAtMs int64          `json:"updatedAtMs,omitempty"`
	Followers   int32          `json:"followers"`
	Followings  int32          `json:"followings"`
	IsModerator bool           `json:"isModerator"`
	Website     string         `json:"website"`
	Twitter     string         `json:"twitter"`
	Instagram   string         `json:"instagram"`
	Discord     string         `json:"discord"`
}

func (i *Info) Sanitized() *Info {
	return &Info{
		Address:     i.Address,
		Alias:       i.Alias,
		Bio:         i.Bio,
		ImageHash:   i.ImageHash,
		CreatedAtMs: i.CreatedAtMs,
		IsModerator: i.IsModerator,
	}
}

// Updater to update account info
type Updater struct {
	Alias         *string   `json:"alias" bson:"alias"`
	Email         *string   `json:"email" bson:"email"`
	Bio           *string   `json:"bio" bson:"bio"`
	ImageHash     *string   `json:"-" bson:"imageHash"`
	BannerHash    *string   `json:"-" bson:"bannerHash"`
	Nonce         int32     `json:"-" bson:"nonce"`
	UpdatedAt     time.Time `json:"-" bson:"updatedAt,omitempty"`
	IsAppropriate *bool     `json:"-" bson:"isAppropriate"`
	Website       *string   `json:"website" bson:"website"`
	Twitter       *string   `json:"twitter" bson:"twitter"`
	Instagram     *string   `json:"instagram" bson:"instagram"`
	Discord       *string   `json:"discord" bson:"discord"`
}

type AccountStat struct {
	Single             int32 `json:"single"`
	Bundle             int32 `json:"bundle"`
	Favorite           int32 `json:"favorite"`
	Collections        int32 `json:"collections"`
	CreatedNfts        int32 `json:"createdNfts"`
	CreatedCollections int32 `json:"createdCollections"`
}

type CollectionId struct {
	domain.ChainId `json:"chainId" bson:"chainId" param:"chainId"`
	domain.Address `json:"erc721Address" bson:"erc721Address" param:"contract"`
}

type AccountCollectionHoldings struct {
	Collections               map[CollectionId]int32 `json:"collections"`
	CollectionsHoldingBalance map[CollectionId]int32 `json:"collectionsHoldingBalance"`
}

var (
	// ErrInvalidNonce occured when validating a signature but the nonce of the address has not generated
	ErrInvalidNonce = errors.New("invalid nonce")
	// ErrSignature occured when a signature is invalid
	ErrInvalidSignature = errors.New("invalid signature")
)

// Usecase is account usecase
type Usecase interface {
	Create(c ctx.Ctx, address domain.Address) (*Info, error)
	Get(c ctx.Ctx, address domain.Address) (*Info, error)
	GenerateNonce(c ctx.Ctx, address domain.Address) (int32, error)
	ValidateSignature(c ctx.Ctx, address domain.Address, signature string) error
	Update(c ctx.Ctx, address domain.Address, updater *Updater) (*Info, error)
	// return ipfs cid
	UpdateAvatar(c ctx.Ctx, address domain.Address, imgData string) (string, error)
	// return ipfs cid
	UpdateBanner(c ctx.Ctx, address domain.Address, imgData string) (string, error)

	Follow(c ctx.Ctx, address, toAddress domain.Address) error
	Unfollow(c ctx.Ctx, address, toAddress domain.Address) error
	IsFollowing(c ctx.Ctx, address, toAddress domain.Address) (bool, error)
	GetFollowers(c ctx.Ctx, address domain.Address) ([]*Info, error)
	GetFollowings(c ctx.Ctx, address domain.Address) ([]*Info, error)

	GetNotificationSettings(c ctx.Ctx, address domain.Address) (*NotificationSettings, error)
	UpsertNotificationSettings(c ctx.Ctx, settings *NotificationSettings) (*NotificationSettings, error)

	Ban(c ctx.Ctx, address domain.Address) error
	Unban(c ctx.Ctx, address domain.Address) error

	GetActivities(c ctx.Ctx, address domain.Address, opts ...FindActivityHistoryOptions) (*ActivityResult, error)

	GetAccountStat(c ctx.Ctx, address domain.Address) (*AccountStat, error)
	GetAccountCollectionHoldings(c ctx.Ctx, address domain.Address) (*AccountCollectionHoldings, error)
}

// Repo is account repo
type Repo interface {
	Get(c ctx.Ctx, address domain.Address) (*Account, error)
	GetAccounts(c ctx.Ctx, addresses []domain.Address) ([]*Account, error)
	Insert(c ctx.Ctx, account *Account) error
	Update(c ctx.Ctx, address domain.Address, account *Updater) error
}
