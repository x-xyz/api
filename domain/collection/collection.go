package collection

import (
	"strconv"
	"strings"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
)

type CollectionId struct {
	domain.ChainId `json:"chainId" bson:"chainId" param:"chainId"`
	domain.Address `json:"erc721Address" bson:"erc721Address" param:"contract"`
}

type Collection struct {
	ChainId       domain.ChainId   `json:"chainId" bson:"chainId"`
	Erc721Address domain.Address   `json:"erc721Address" bson:"erc721Address"`
	TokenType     domain.TokenType `json:"tokenType" bson:"tokenType"`
	// contract register
	Owner           domain.Address `json:"owner" bson:"owner"`
	Email           string         `json:"email" bson:"email"`
	CollectionName  string         `json:"collectionName" bson:"collectionName"`
	Description     string         `json:"description" bson:"description"`
	Categories      []string       `json:"categories" bson:"categories"`
	LogoImageHash   string         `json:"logoImageHash" bson:"logoImageHash"`
	LogoImageUrl    string         `json:"logoImageUrl" bson:"logoImageUrl"`
	CoverImageHash  string         `json:"coverImageHash" bson:"coverImageHash"`
	CoverImageURL   string         `json:"coverImageUrl" bson:"coverImageUrl"`
	SiteUrl         string         `json:"siteUrl" bson:"siteUrl"`
	Discord         string         `json:"discord" bson:"discord"`
	TwitterHandle   string         `json:"twitterHandle" bson:"twitterHandle"`
	InstagramHandle string         `json:"instagramHandle" bson:"instagramHandle"`
	MediumHandle    string         `json:"mediumHandle" bson:"mediumHandle"`
	Telegram        string         `json:"telegram" bson:"telegram"`
	// false for unreviewed collections
	Status bool `json:"-" bson:"status"`
	// is create by public or private factory
	IsInternal bool `json:"-" bson:"isInternal"`
	// is created by private factory
	IsOwnerble bool `json:"-" bson:"isOwnerble"`
	// banned or not
	IsAppropriate bool    `json:"-" bson:"isAppropriate"`
	IsVerified    bool    `json:"isVerified" bson:"isVerified"`
	Royalty       float64 `json:"royalty" bson:"royalty"`
	FeeRecipient  string  `json:"feeRecipient" bson:"feeRecipient"`
	// inject in usecase
	IsRegistered bool `json:"isRegistered" bson:"-"`
	// total supply, will be updated by indexer
	Supply int64 `json:"supply" bson:"supply"`
	// attribute supply, for calculating rarity
	// trait_type -> trait_name -> count
	// not available for nested attributes
	// this field will be updated by indexer
	Attributes                map[string]map[string]int64 `json:"attributes" bson:"attributes"`
	AttributesHash            string                      `json:"-" bson:"attributesHash"`
	ShouldCalculateOpenrarity bool                        `json:"-" bson:"shouldCalculateOpenrarity"`
	// num of owners, will be updated by indexer
	NumOwners         int64   `json:"numOwners" bson:"numOwners"`
	NumOwnersMovement float64 `json:"numOwnersMovement" bson:"numOwnersMovement"`
	// total trading volume in native token of corresponding chain
	TotalVolume float64 `json:"totalVolume" bson:"totalVolume"`
	// current floor price
	FloorPriceInNative        float64 `json:"floorPrice" bson:"floorPrice"`
	OpenseaFloorPriceInNative float64 `json:"openseaFloorPriceInNative" bson:"openseaFloorPriceInNative"`
	OpenseaFloorPriceInUsd    float64 `json:"openseaFloorPriceInUsd" bson:"openseaFloorPriceInUsd"`
	OpenseaFloorPriceMovement float64 `json:"openseaFloorPriceMovement" bson:"openseaFloorPriceMovement"`
	// current floor price in USD
	FloorPriceInUsd    float64 `json:"usdFloorPrice" bson:"usdFloorPrice"`
	FloorPriceMovement float64 `json:"floorPriceMovement" bson:"floorPriceMovement"`
	HasFloorPrice      bool    `json:"hasFloorPrice" bson:"hasFloorPrice"`
	// highest sale price in native token
	HighestSale             float64                       `json:"highestSale" bson:"highestSale"`
	HighestSaleInUsd        float64                       `json:"highestSaleInUsd" bson:"highestSaleInUsd"`
	LastSoldAt              time.Time                     `json:"lastSoldAt" bson:"lastSoldAt"`
	HasBeenSold             bool                          `json:"hasBeenSold" bson:"hasBeenSold"`
	LastListedAt            time.Time                     `json:"lastListedAt" bson:"lastListedAt"`
	HasBeenListed           bool                          `json:"hasBeenListed" bson:"hasBeenListed"`
	ViewCount               int32                         `json:"viewCount" bson:"viewCount"`
	Liked                   int32                         `json:"liked" bson:"liked"`
	EditableAccounts        []domain.Address              `json:"editableAccounts" bson:"editableAccounts"`
	LastOpenseaEventIndexAt time.Time                     `json:"-" bson:"lastOpenseaEventIndexAt"`
	TraitFloorPrice         map[string]map[string]float64 `json:"traitFloorPrice" bson:"traitFloorPrice"`
}

func (c *Collection) ToId() CollectionId {
	return CollectionId{
		ChainId: c.ChainId,
		Address: c.Erc721Address,
	}
}

type CreatePayload struct {
	ChainId         domain.ChainId   `json:"chainId" bson:"chainId"`
	Erc721Address   domain.Address   `json:"erc721Address" bson:"erc721Address"`
	TokenType       domain.TokenType `json:"tokenType" bson:"tokenType"`
	Owner           domain.Address   `json:"owner" bson:"owner"`
	Email           string           `json:"email" bson:"email"`
	CollectionName  string           `json:"collectionName" bson:"collectionName"`
	Description     string           `json:"description" bson:"description"`
	Categories      []string         `json:"categories" bson:"categories"`
	LogoImage       string           `json:"logoImage" bson:"-"`
	LogoImageHash   string           `json:"-" bson:"logoImageHash"`
	LogoImageUrl    string           `json:"logoImageUrl" bson:"logoImageUrl"`
	CoverImage      string           `json:"coverImage" bson:"-"`
	CoverImageHash  string           `json:"-" bson:"coverImageHash"`
	CoverImageURL   string           `json:"coverImageUrl" bson:"coverImageUrl"`
	SiteUrl         string           `json:"siteUrl" bson:"siteUrl"`
	Discord         string           `json:"discord" bson:"discord"`
	TwitterHandle   string           `json:"twitterHandle" bson:"twitterHandle"`
	InstagramHandle string           `json:"instagramHandle" bson:"instagramHandle"`
	MediumHandle    string           `json:"mediumHandle" bson:"mediumHandle"`
	Telegram        string           `json:"telegram" bson:"telegram"`
	Royalty         float64          `json:"royalty" bson:"royalty"`
	FeeRecipient    string           `json:"feeRecipient" bson:"feeRecipient"`
	// determined by whether contract address existed in factory or not
	IsInternal bool `json:"-" bson:"isInternal"`
	// is internal private contract
	IsOwnerble bool `json:"-" bson:"isOwnerble"`
	// default false
	Status        bool `json:"-" bson:"status"`
	IsAppropriate bool `json:"-" bson:"isAppropriate"`
	IsVerified    bool `json:"isVerified" bson:"isVerified"`
}

type UpdatePayload struct {
	Owner           domain.Address `json:"-" bson:"owner"`
	Email           string         `json:"-" bson:"email"`
	CollectionName  string         `json:"-" bson:"collectionName"`
	Description     string         `json:"-" bson:"description"`
	Categories      []string       `json:"-" bson:"categories"`
	LogoImageHash   string         `json:"-" bson:"logoImageHash"`
	LogoImageUrl    string         `json:"-" bson:"logoImageUrl"`
	SiteUrl         string         `json:"-" bson:"siteUrl"`
	Discord         string         `json:"-" bson:"discord"`
	TwitterHandle   string         `json:"-" bson:"twitterHandle"`
	InstagramHandle string         `json:"-" bson:"instagramHandle"`
	MediumHandle    string         `json:"-" bson:"mediumHandle"`
	Telegram        string         `json:"-" bson:"telegram"`
	Royalty         float64        `json:"-" bson:"royalty"`
	FeeRecipient    string         `json:"-" bson:"feeRecipient"`
	Status          bool           `json:"-" bson:"status"`
	// use pointer to prevent be ignored when making bson
	IsAppropriate *bool `json:"-" bson:"isAppropriate"`
	IsVerified    bool  `json:"-" bson:"isVerified"`
	// supply and attributes will be updated by indexer
	Supply         int64                       `json:"supply" bson:"supply"`
	Attributes     map[string]map[string]int64 `json:"attributes" bson:"attributes"`
	AttributesHash string                      `json:"-" bson:"attributesHash,omitempty"`
	// num of owners, will be updated by indexer
	NumOwners         int64    `bson:"numOwners,omitempty"`
	NumOwnersMovement *float64 `bson:"numOwnersMovement"`
	// highest sale price in native token
	HighestSale      float64   `bson:"highestSale,omitempty"`
	HighestSaleInUsd float64   `bson:"highestSaleInUsd,omitempty"`
	LastSoldAt       time.Time `bson:"lastSoldAt,omitempty"`
	HasBeenSold      bool      `bson:"hasBeenSold,omitempty"`
	LastListedAt     time.Time `bson:"lastListedAt,omitempty"`
	HasBeenListed    bool      `bson:"hasBeenListed,omitempty"`
	// floor price in native token
	FloorPriceInNative        *float64  `bson:"floorPrice,omitempty"`
	OpenseaFloorPriceInNative *float64  `bson:"openseaFloorPriceInNative,omitempty"`
	OpenseaFloorPriceInUsd    *float64  `bson:"openseaFloorPriceInUsd,omitempty"`
	OpenseaFloorPriceInApe    *float64  `bson:"openseaFloorPriceInApe,omitempty"`
	OpenseaFloorPriceMovement *float64  `bson:"openseaFloorPriceMovement,omitempty"`
	FloorPriceInUsd           *float64  `bson:"usdFloorPrice,omitempty"`
	FloorPriceMovement        *float64  `bson:"floorPriceMovement"`
	HasFloorPrice             *bool     `bson:"hasFloorPrice,omitempty"`
	LastOpenseaEventIndexAt   time.Time `bson:"lastOpenseaEventIndexAt,omitempty"`
	// price is display price, ie: 4.2 ETH = 4.2
	TraitFloorPrice map[string]map[string]float64 `bson:"traitFloorPrice,omitempty"`
}

type UpdateInfoPayload struct {
	Email           *string  `json:"email" bson:"email"`
	ColectionName   *string  `json:"collectionName" bson:"collectionName"`
	Description     *string  `json:"description" bson:"description"`
	Categories      []string `json:"categories" bson:"categories"`
	LogoImageHash   *string  `json:"logoImageHash" bson:"logoImageHash"`
	LogoImageUrl    *string  `json:"logoImageUrl" bson:"logoImageUrl"`
	SiteUrl         *string  `json:"siteUrl" bson:"siteUrl"`
	Discord         *string  `json:"discord" bson:"discord"`
	TwitterHandle   *string  `json:"twitterHandle" bson:"twitterHandle"`
	InstagramHandle *string  `json:"instagramHandle" bson:"instagramHandle"`
	MediumHandle    *string  `json:"mediumHandle" bson:"mediumHandle"`
	Telegram        *string  `json:"telegram" bson:"telegram"`
}

type CollectionWithTradingVolume struct {
	ChainId                   domain.ChainId `json:"chainId"`
	Erc721Address             domain.Address `json:"erc721Address"`
	CollectionName            string         `json:"collectionName"`
	LogoImageHash             string         `json:"logoImageHash"`
	LogoImageUrl              string         `json:"logoImageUrl"`
	Sales                     float64        `json:"sales"`
	Volume                    float64        `json:"volume"`
	VolumeInUsd               float64        `json:"volumeInUsd"`
	VolumeInApe               float64        `json:"volumeInApe"`
	ChangeRatio               float64        `json:"changeRatio"`
	OpenseaFloorPriceInNative float64        `json:"openseaFloorPriceInNative"`
	OpenseaFloorPriceInUsd    float64        `json:"openseaFloorPriceInUsd"`
	OpenseaFloorPriceInApe    float64        `json:"openseaFloorPriceInApe"`
	OpenseaFloorPriceMovement float64        `json:"openseaFloorPriceMovement"`
	Supply                    int64          `json:"supply"`
	NumOwners                 int64          `json:"numOwners"`
	EligibleForPromo          bool           `json:"eligibleForPromo"`
}

type CollectionWithHoldingCount struct {
	Collection
	HoldingCount   int32 `json:"holdingCount"`
	HoldingBalance int32 `json:"holdingBalance"`
}

type CollectionWithStat struct {
	Collection
	OpenseaSalesVolume       float64 `json:"openseaSalesVolume"`
	OpenseaSalesVolumeChange float64 `json:"openseaSalesVolumeChange"`
	OpenseaFloorPriceInApe   float64 `json:"openseaFloorPriceInApe"`
}

type CollectionWithStatByAccount struct {
	Collection
	OwnedNftCount            int     `json:"ownedNftCount"`
	TotalValue               float64 `json:"totalValue"`
	InstantLiquidityInUsd    float64 `json:"instantLiquidityInUsd"`
	InstantLiquidityRatio    float64 `json:"instantLiquidityRatio"`
	OpenseaSalesVolume       float64 `json:"openseaSalesVolume"`
	OpenseaSalesVolumeChange float64 `json:"openseaSalesVolumeChange"`
}

type ActivityResult struct {
	Items []account.ActivityHistory `json:"items"`
	Count int                       `json:"count"`
}

type GlobalOfferStatRow struct {
	DisplayPrice string  `json:"displayPrice"`
	PriceInUsd   float64 `json:"priceInUsd"`
	Size         int     `json:"size"`
	Sum          string  `json:"sum"`
	Bidders      int     `json:"bidders"`
}

type GlobalOfferStatResult struct {
	Rows []GlobalOfferStatRow `json:"row"`
}

type findAllOptions struct {
	SortBy           *string
	SortDir          *domain.SortDir
	Offset           *int32
	Limit            *int32
	ChainId          *domain.ChainId
	Addresses        *[]domain.Address
	Category         *string
	Status           *bool
	IsAppropriate    *bool
	IsInternal       *bool
	IsOwnerble       *bool
	Owner            *domain.Address
	FloorPriceGTE    *float64
	FloorPriceLTE    *float64
	UsdFloorPriceGTE *float64
	UsdFloorPriceLTE *float64
	AccountEditable  *domain.Address
	LikedBy          *domain.Address
	ListedBy         *domain.Address
	OfferedBy        *domain.Address
}

type FindAllOptions func(*findAllOptions) error

func GetFindAllOptions(opts ...FindAllOptions) (findAllOptions, error) {
	res := findAllOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func WithSort(sortby string, sortdir domain.SortDir) FindAllOptions {
	return func(options *findAllOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func WithPagination(offset int32, limit int32) FindAllOptions {
	return func(options *findAllOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func WithChainId(chainId domain.ChainId) FindAllOptions {
	return func(options *findAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func WithCategory(category string) FindAllOptions {
	return func(options *findAllOptions) error {
		options.Category = &category
		return nil
	}
}

func WithStatus(status bool) FindAllOptions {
	return func(options *findAllOptions) error {
		options.Status = &status
		return nil
	}
}

func WithIsAppropriate(isAppropriate bool) FindAllOptions {
	return func(options *findAllOptions) error {
		options.IsAppropriate = &isAppropriate
		return nil
	}
}

func WithIsInternal(isInternal bool) FindAllOptions {
	return func(options *findAllOptions) error {
		options.IsInternal = &isInternal
		return nil
	}
}

func WithIsOwnerble(isOwnerble bool) FindAllOptions {
	return func(options *findAllOptions) error {
		options.IsOwnerble = &isOwnerble
		return nil
	}
}

func WithOwner(owner domain.Address) FindAllOptions {
	return func(options *findAllOptions) error {
		options.Owner = &owner
		return nil
	}
}

func WithFloorPriceGTE(val float64) FindAllOptions {
	return func(options *findAllOptions) error {
		options.FloorPriceGTE = &val
		return nil
	}
}
func WithFloorPriceLTE(val float64) FindAllOptions {
	return func(options *findAllOptions) error {
		options.FloorPriceLTE = &val
		return nil
	}
}
func WithUsdFloorPriceGTE(val float64) FindAllOptions {
	return func(options *findAllOptions) error {
		options.UsdFloorPriceGTE = &val
		return nil
	}
}
func WithUsdFloorPriceLTE(val float64) FindAllOptions {
	return func(options *findAllOptions) error {
		options.UsdFloorPriceLTE = &val
		return nil
	}
}
func WithAccountEditable(account domain.Address) FindAllOptions {
	return func(options *findAllOptions) error {
		lower := account.ToLower()
		options.AccountEditable = &lower
		return nil
	}
}

func WithAddresses(addresses []domain.Address) FindAllOptions {
	return func(options *findAllOptions) error {
		_addresses := make([]domain.Address, len(addresses))
		for i, address := range addresses {
			_addresses[i] = address.ToLower()
		}
		options.Addresses = &_addresses
		return nil
	}
}

func WithLikedBy(address domain.Address) FindAllOptions {
	return func(options *findAllOptions) error {
		options.LikedBy = address.ToLowerPtr()
		return nil
	}
}

func WithListedBy(address domain.Address) FindAllOptions {
	return func(options *findAllOptions) error {
		options.ListedBy = &address
		return nil
	}
}

func WithOfferedBy(address domain.Address) FindAllOptions {
	return func(options *findAllOptions) error {
		options.OfferedBy = &address
		return nil
	}
}

type Repo interface {
	FindAll(c ctx.Ctx, opts ...FindAllOptions) ([]*Collection, error)
	Count(c ctx.Ctx, opts ...FindAllOptions) (int, error)
	FindOne(c ctx.Ctx, id CollectionId) (*Collection, error)
	Create(c ctx.Ctx, value CreatePayload) error
	Upsert(c ctx.Ctx, value CreatePayload) error
	Update(c ctx.Ctx, id CollectionId, value UpdatePayload) error
	IncreaseViewCount(c ctx.Ctx, id CollectionId, count int) (int32, error)
	IncreaseLikeCount(c ctx.Ctx, id CollectionId, count int) (int32, error)
}

type Usecase interface {
	FindAll(c ctx.Ctx, opts ...FindAllOptions) (*SearchResult, error)
	FindAllIncludingUnregistered(c ctx.Ctx, optFns ...FindAllOptions) ([]*CollectionWithHoldingCount, error)
	FindAllMintable(c ctx.Ctx, eoa domain.Address, opts ...FindAllOptions) ([]*Collection, error)
	FindAllUnreviewd(c ctx.Ctx, eoa domain.Address) ([]*Registration, error)
	FindOne(c ctx.Ctx, id CollectionId) (*Collection, error)
	FindOneWithStat(c ctx.Ctx, id CollectionId) (*CollectionWithStat, error)
	CreateErc1155(c ctx.Ctx, value CreatePayload) error
	CreateErc721(c ctx.Ctx, value CreatePayload) error
	Register(c ctx.Ctx, value Registration) (*Registration, error)
	Accept(c ctx.Ctx, id CollectionId) (*Collection, error)
	Reject(c ctx.Ctx, id CollectionId, reason string) (*Registration, error)
	Ban(c ctx.Ctx, id CollectionId, ban bool) (*Collection, error)
	RefreshStat(c ctx.Ctx, id CollectionId) error
	GetTopCollections(c ctx.Ctx, periodType PeriodType, opts ...domain.OpenseaDataFindAllOptions) ([]CollectionWithTradingVolume, error)
	GetViewCount(c ctx.Ctx, id CollectionId) (int32, error)
	UpdateSaleStat(c ctx.Ctx, id CollectionId, priceInNative, priceInUsd float64, blkTime time.Time) error
	UpdateLastListedAt(c ctx.Ctx, id CollectionId, blkTime time.Time) error
	UpdateInfo(c ctx.Ctx, id CollectionId, info UpdateInfoPayload) error
	UpdateLastOpenseaEventIndexAt(c ctx.Ctx, id CollectionId, t time.Time) error
	UpdateTraitFloorPrice(c ctx.Ctx, id CollectionId, traitName, traitValue string, price float64) error
	UpdateOpenseaFloorPrice(c ctx.Ctx, id CollectionId, price float64) error
	GetCollectionStatByAccount(c ctx.Ctx, id CollectionId, account domain.Address) (*CollectionWithStatByAccount, error)
	GetActivities(c ctx.Ctx, id CollectionId, optFns ...account.FindActivityHistoryOptions) (*ActivityResult, error)
	GetGlobalOfferStats(c ctx.Ctx, id CollectionId) (*GlobalOfferStatResult, error)
}

func ToCollectionKey(chainId domain.ChainId, address domain.Address) string {
	return strings.Join([]string{strconv.Itoa(int(chainId)), address.ToLowerStr()}, ":")
}

func FromCollectionKey(key string) (domain.ChainId, domain.Address, error) {
	parts := strings.Split(key, ":")
	chainId, err := strconv.Atoi(parts[0])
	if err != nil {
		return domain.ChainId(0), "", err
	}
	return domain.ChainId(chainId), domain.Address(parts[1]), nil
}

type SearchResult struct {
	Items []*CollectionWithHoldingCount `json:"items"`
	Count int                           `json:"count"`
}
