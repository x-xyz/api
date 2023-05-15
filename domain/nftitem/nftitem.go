package nftitem

import (
	"fmt"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IndexerState string

const (
	IndexerStateNew                       = "new"                          // newly minted
	IndexerStateNewRefreshing             = "new_refreshing"               // newly refreshing
	IndexerStateHasTokenURI               = "has_token_uri"                // db has TokenUri
	IndexerStatePendingTokenURIRefreshing = "pending_token_uri_refreshing" // pending for doing refreshing
	IndexerStateHasTokenURIRefreshing     = "has_token_uri_refreshing"     // db has TokenUri in refreshing stage
	IndexerStateHasImageURL               = "has_image_url"                // (hosted) metadata, db has ImageUrl
	IndexerStateHasHostedImage            = "has_hosted_image"             // db has hosted image url
	IndexerStateParsingAttributes         = "parsing_attributes"           // waiting or parsing attributes from metadata
	IndexerStateFetchingAnimation         = "fetching_animation"           // attempt to get animation media
	IndexerStateDone                      = "done"                         // hosted metadata, hosted image
	IndexerStateInvalid                   = "invalid"                      // no image
	IndexerStateBeforeMigrate             = "before_migrate"               // db has TokenUri and ImageUrl
	IndexerStateBeforeMigrateMimeType     = "before_migrate_mimetype"
)

var ReadyToServeIndexerStates = []IndexerState{
	IndexerStateNewRefreshing,
	IndexerStateHasTokenURIRefreshing,
	IndexerStateHasTokenURI,
	IndexerStateHasImageURL,
	IndexerStateHasHostedImage,
	IndexerStateParsingAttributes,
	IndexerStateFetchingAnimation,
	IndexerStateDone,
	IndexerStateBeforeMigrate,
	IndexerStateBeforeMigrateMimeType,
}
var ReadyToServeIndexerStatesMap = map[IndexerState]struct{}{
	IndexerStateNewRefreshing:         {},
	IndexerStateHasTokenURIRefreshing: {},
	IndexerStateHasTokenURI:           {},
	IndexerStateHasImageURL:           {},
	IndexerStateHasHostedImage:        {},
	IndexerStateParsingAttributes:     {},
	IndexerStateFetchingAnimation:     {},
	IndexerStateDone:                  {},
	IndexerStateBeforeMigrate:         {},
	IndexerStateBeforeMigrateMimeType: {},
}

type Id struct {
	ChainId         domain.ChainId `json:"chainId" bson:"chainId"`
	ContractAddress domain.Address `json:"contractAddress" bson:"contractAddress"`
	TokenId         domain.TokenId `json:"tokenId" bson:"tokenID"`
}

type NftItem struct {
	ObjectId                  primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	ChainId                   domain.ChainId     `json:"chainId" bson:"chainId"`
	ContractAddress           domain.Address     `json:"contractAddress" bson:"contractAddress"`
	TokenId                   domain.TokenId     `json:"tokenId" bson:"tokenID"`
	TokenType                 domain.TokenType   `json:"tokenType" bson:"tokenType"`
	TokenUri                  string             `json:"tokenUri" bson:"tokenURI"`
	ImageUrl                  string             `json:"imageUrl" bson:"imageURL"`
	ThumbnailPath             string             `json:"thumbnailPath" bson:"thumbnailPath"`
	ImagePath                 string             `json:"imagePath" bson:"imagePath"`
	HostedTokenUri            string             `json:"hostedTokenUri" bson:"hostedTokenURI"`
	HostedImageUrl            string             `json:"hostedImageUrl" bson:"hostedImageURL"`
	AnimationUrl              string             `json:"animationUrl" bson:"animationUrl"`
	HostedAnimationUrl        string             `json:"hostedAnimationUrl" bson:"hostedAnimationUrl"`
	AnimationUrlContentType   string             `json:"animationUrlContentType" bson:"animationUrlContentType"`
	AnimationUrlMimeType      string             `json:"animationUrlMimeType" bson:"animationUrlMimeType"`
	Symbol                    string             `json:"symbol" bson:"symbol"`
	Name                      string             `json:"name" bson:"name"`
	Owner                     domain.Address     `json:"owner" bson:"owner"`
	NumOwners                 int64              `json:"numOwners" bson:"numOwners"`
	Supply                    int32              `json:"supply" bson:"supply"`
	Royalty                   int32              `json:"royalty" bson:"royalty"`
	Price                     *float64           `json:"price" bson:"price"`
	PaymentToken              *domain.Address    `json:"paymentToken" bson:"paymentToken"`
	PriceInUsd                *float64           `json:"priceInUsd" bson:"priceInUSD"`
	PriceSource               *PriceSource       `json:"priceSource" bson:"priceSource"`
	LastSalePrice             float64            `json:"lastSalePrice" bson:"lastSalePrice"`
	LastSalePricePaymentToken string             `json:"lastSalePricePaymentToken" bson:"lastSalePricePaymentToken"`
	LastSalePriceInUsd        float64            `json:"lastSalePriceInUsd" bson:"lastSalePriceInUSD"`
	LastSalePriceInNative     float64            `json:"lastSalePriceInNative" bson:"lastSalePriceInNative"`
	Viewed                    int32              `json:"viewed" bson:"viewed"`
	Liked                     int32              `json:"liked" bson:"liked"`
	ContentType               string             `json:"contentType" bson:"contentType"`
	MimeType                  string             `json:"mimeType" bson:"mimeType"`
	IsAppropriate             *bool              `json:"isAppropriate" bson:"isAppropriate"`
	IsFiltered                bool               `json:"isFiltered" bson:"isFiltered"`
	BlockNumber               domain.BlockNumber `json:"blockNumber" bson:"blockNumber"`
	ListedAt                  *time.Time         `json:"listedAt,omitempty" bson:"listedAt"`
	SoldAt                    *time.Time         `json:"soldAt,omitempty" bson:"soldAt"`
	SaleEndsAt                *time.Time         `json:"saleEndsAt,omitempty" bson:"saleEndsAt"`
	IndexerState              IndexerState       `json:"indexerState" bson:"indexerState"`
	IndexerRetryCount         int32              `json:"-" bson:"indexerRetryCount"`
	CreatedAt                 time.Time          `json:"createdAt,omitempty" bson:"createdAt"`
	UpdatedAt                 time.Time          `json:"updatedAt,omitempty" bson:"updatedAt"`
	Attributes                Attributes         `json:"attributes" bson:"attributes"`
	Creator                   domain.Address     `json:"creator" bson:"creator"`
	HasActiveListings         bool               `json:"-" bson:"hasActiveListings"`
	OpenrarityRank            int                `json:"openrarityRank" bson:"openrarityRank"`
	OpenrarityScore           float64            `json:"openrarityScore" bson:"openrarityScore"`

	// ListingEndsAt is calculated from orders, take the last end listing
	ListingEndsAt         *time.Time       `json:"listingEndsAt,omitempty" bson:"listingEndsAt"`
	ListingOwners         []domain.Address `json:"listingOwners" bson:"listingOwners"`
	InactiveListingOwners []domain.Address `json:"inactiveListingOwners" bson:"inactiveListingOwners"`
	// OfferEndsAt is calculated from orders, take the last end offer
	OfferEndsAt           *time.Time       `json:"offerEndsAt,omitempty" bson:"offerEndsAt"`
	OfferOwners           []domain.Address `json:"offerOwners" bson:"offerOwners"`
	OfferStartsAt         *time.Time       `json:"offerStartsAt,omitempty" bson:"offerStartsAt"`
	InstantLiquidityInUsd float64          `json:"instantLiquidityInUsd" bson:"instantLiquidityInUsd"`
	HasOrder              bool             `json:"-" bson:"hasOrder"`
}

type PatchableNftItem struct {
	Owner                     *domain.Address     `json:"owner" bson:"owner"`
	TokenUri                  *string             `json:"tokenUri" bson:"tokenURI"`
	ImageUrl                  *string             `json:"imageUrl" bson:"imageURL"`
	ThumbnailPath             *string             `json:"thumbnailPath" bson:"thumbnailPath"`
	HostedTokenUri            *string             `json:"hostedTokenUri" bson:"hostedTokenURI"`
	HostedImageUrl            *string             `json:"hostedImageUrl" bson:"hostedImageURL"`
	Name                      *string             `json:"name" bson:"name"`
	NumOwners                 *int64              `json:"numOwners" bson:"numOwners"`
	ContentType               *string             `json:"contentType" bson:"contentType"`
	MimeType                  *string             `json:"mimeType" bson:"mimeType"`
	AnimationUrl              *string             `json:"animationUrl" bson:"animationUrl"`
	HostedAnimationUrl        *string             `json:"hostedAnimationUrl" bson:"hostedAnimationUrl"`
	AnimationUrlContentType   *string             `json:"animationUrlContentType" bson:"animationUrlContentType"`
	AnimationUrlMimeType      *string             `json:"animationUrlMimeType" bson:"animationUrlMimeType"`
	BlockNumber               *domain.BlockNumber `json:"blockNumber" bson:"blockNumber"`
	ListedAt                  *time.Time          `json:"listedAt,omitempty" bson:"listedAt"`
	SoldAt                    *time.Time          `json:"soldAt,omitempty" bson:"soldAt"`
	SaleEndsAt                *time.Time          `json:"saleEndsAt,omitempty" bson:"saleEndsAt"`
	LastSalePrice             *float64            `json:"lastSalePrice" bson:"lastSalePrice"`
	LastSalePricePaymentToken *string             `json:"lastSalePricePaymentToken" bson:"lastSalePricePaymentToken"`
	LastSalePriceInUsd        *float64            `json:"lastSalePriceInUsd" bson:"lastSalePriceInUSD"`
	LastSalePriceInNative     *float64            `json:"lastSalePriceInNative" bson:"lastSalePriceInNative"`
	Price                     *float64            `json:"price" bson:"price"`
	PaymentToken              *domain.Address     `json:"paymentToken" bson:"paymentToken"`
	PriceInUsd                *float64            `json:"priceInUsd" bson:"priceInUSD"`
	PriceSource               *PriceSource        `json:"priceSource" bson:"priceSource"`
	IsAppropriate             *bool               `json:"isAppropriate" bson:"isAppropriate"`
	IndexerState              *IndexerState       `json:"indexerState" bson:"indexerState"`
	IndexerRetryCount         *int32              `json:"-" bson:"indexerRetryCount"`
	Attributes                Attributes          `json:"attributes" bson:"attributes"`
	HasActiveListings         *bool               `json:"-" bson:"hasActiveListings"`
	OpenrarityRank            *int                `json:"openrarityRank" bson:"openrarityRank"`
	OpenrarityScore           *float64            `json:"openrarityScore" bson:"openrarityScore"`

	// ListingEndsAt is calculated from orders, take the last end listing
	ListingEndsAt         *time.Time       `json:"listingEndsAt,omitempty" bson:"listingEndsAt"`
	ListingOwners         []domain.Address `json:"listingOwners" bson:"listingOwners"`
	InactiveListingOwners []domain.Address `json:"inactiveListingOwners" bson:"inactiveListingOwners"`
	// OfferEndsAt is calculated from orders, take the last end offer
	OfferEndsAt           *time.Time       `json:"offerEndsAt,omitempty" bson:"offerEndsAt"`
	OfferOwners           []domain.Address `json:"offerOwners" bson:"offerOwners"`
	OfferStartsAt         *time.Time       `json:"offerStartsAt,omitempty" bson:"offerStartsAt"`
	InstantLiquidityInUsd *float64         `json:"instantLiquidityInUsd" bson:"instantLiquidityInUsd"`
	HasOrder              *bool            `json:"-" bson:"hasOrder"`
}

func (i *NftItem) ToId() *Id {
	return &Id{
		ChainId:         i.ChainId,
		ContractAddress: i.ContractAddress,
		TokenId:         i.TokenId,
	}
}

func (i *Id) ToString() string {
	return fmt.Sprintf("%v_%s_%s", i.ChainId, i.ContractAddress, i.TokenId)
}

func (i *NftItem) ToSimpleNftItem() *SimpleNftItem {
	return &SimpleNftItem{
		ChainId:        i.ChainId,
		Contract:       i.ContractAddress,
		TokenId:        i.TokenId,
		Name:           i.Name,
		TokenUri:       i.TokenUri,
		ThumbnailPath:  i.ThumbnailPath,
		ImagePath:      i.ImagePath,
		ImageUrl:       i.ImageUrl,
		HostedImageUrl: i.HostedImageUrl,
		HostedTokenUri: i.HostedTokenUri,
	}
}

type SimpleNftItem struct {
	ChainId        domain.ChainId `json:"chainId"`
	Contract       domain.Address `json:"contract"`
	TokenId        domain.TokenId `json:"tokenId"`
	Name           string         `json:"name"`
	TokenUri       string         `json:"tokenUri"`
	ThumbnailPath  string         `json:"thumbnailPath"`
	ImagePath      string         `json:"imagePath"`
	ImageUrl       string         `json:"imageUrl"`
	HostedImageUrl string         `json:"hostedImageUrl"`
	HostedTokenUri string         `json:"hostedTokenUri"`
}

type NftitemWith1155Balance struct {
	NftItem
	Balance int `json:"balance"`
}

// AttributeFilter is used to filter nftitem list
type AttributeFilter struct {
	Name   string   `query:"name"`
	Values []string `query:"values"`
}

type FindAllOptions struct {
	SortBy              *string
	SortDir             *domain.SortDir
	Sorts               *[]string
	ChainId             *domain.ChainId
	ContractAddresses   []domain.Address
	Owner               *domain.Address
	NotOwner            *bool
	ListingFrom         *domain.Address
	InactiveListingFrom *domain.Address
	Offset              *int32
	Limit               *int32
	IsAppropriate       *bool
	IndexerStates       []IndexerState
	IndexerRetryCountLT *int
	Attributes          []AttributeFilter
	SaleStatus          SaleStatus
	PriceGTE            *float64
	PriceLTE            *float64
	PriceInUsdGTE       *float64
	PriceInUsdLTE       *float64
	OfferPriceInUsdGTE  *float64
	OfferPriceInUsdLTE  *float64
	Ids                 *[]Id
	Name                *string
	Search              *string
	OfferOwners         []domain.Address
	HoldingIds          *[]Id
	TokenType           *domain.TokenType
	BidOwner            *domain.Address
	ObjectIdLT          *primitive.ObjectID
	HasOrder            *bool
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

func WithSort(sortby string, sortdir domain.SortDir) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func WithSorts(sorts []string) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Sorts = &sorts
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

func WithChainId(chainId domain.ChainId) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func WithOwner(address domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Owner = address.ToLowerPtr()
		options.NotOwner = ptr.Bool(false)
		return nil
	}
}

func WithNotOwner(address domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Owner = address.ToLowerPtr()
		options.NotOwner = ptr.Bool(true)
		return nil
	}
}

func WithPagination(offset int32, limit int32) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func WithIsAppropriate(isAppropriate bool) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.IsAppropriate = &isAppropriate
		return nil
	}
}

func WithIndexerStates(indexerStates []IndexerState) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.IndexerStates = indexerStates
		return nil
	}
}

func WithIndexerRetryCountLT(retryCountLT int) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.IndexerRetryCountLT = &retryCountLT
		return nil
	}
}

func WithAttributeFilters(attributes []AttributeFilter) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Attributes = attributes
		return nil
	}
}

func WithBuyNow() FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.SaleStatus = SetSaleStatus(options.SaleStatus, SaleStatusBuyNow)
		return nil
	}
}

func WithHasBid() FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.SaleStatus = SetSaleStatus(options.SaleStatus, SaleStatusHasBid)
		return nil
	}
}

func WithHasOffer() FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.SaleStatus = SetSaleStatus(options.SaleStatus, SaleStatusHasOffer)
		return nil
	}
}

func WithOnAuction() FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.SaleStatus = SetSaleStatus(options.SaleStatus, SaleStatusOnAuction)
		return nil
	}
}

func WithHasTraded() FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.SaleStatus = SetSaleStatus(options.SaleStatus, SaleStatusHasTraded)
		return nil
	}
}

func WithPriceGTE(val float64) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.PriceGTE = &val
		return nil
	}
}

func WithPriceLTE(val float64) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.PriceLTE = &val
		return nil
	}
}

func WithPriceInUsdGTE(val float64) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.PriceInUsdGTE = &val
		return nil
	}
}

func WithPriceInUsdLTE(val float64) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.PriceInUsdLTE = &val
		return nil
	}
}

func WithOfferPriceInUsdGTE(val float64) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.OfferPriceInUsdGTE = &val
		return nil
	}
}

func WithOfferPriceInUsdLTE(val float64) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.OfferPriceInUsdLTE = &val
		return nil
	}
}

func WithNftitemIds(ids []Id) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Ids = &ids
		return nil
	}
}

func WithListingFrom(owner domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.ListingFrom = owner.ToLowerPtr()
		return nil
	}
}

func WithInactiveListingFrom(owner domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.InactiveListingFrom = owner.ToLowerPtr()
		return nil
	}
}

func WithName(name string) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Name = &name
		return nil
	}
}

func WithSearch(search string) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.Search = &search
		return nil
	}
}

func WithOfferOwners(owners []domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.OfferOwners = owners
		return nil
	}
}

func WithHoldingIds(ids []Id) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.HoldingIds = &ids
		return nil
	}
}

func WithTokenType(tokenType domain.TokenType) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.TokenType = &tokenType
		return nil
	}
}

func WithBidOwner(owner domain.Address) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.BidOwner = owner.ToLowerPtr()
		return nil
	}
}

func WithObjectIdLT(objectId primitive.ObjectID) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.ObjectIdLT = &objectId
		return nil
	}
}

func WithHasOrder(hasOrder bool) FindAllOptionsFunc {
	return func(options *FindAllOptions) error {
		options.HasOrder = &hasOrder
		return nil
	}
}

type Repo interface {
	FindAll(c ctx.Ctx, opts ...FindAllOptionsFunc) ([]*NftItem, error)
	Count(c ctx.Ctx, opts ...FindAllOptionsFunc) (int, error)
	FindOne(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) (*NftItem, error)
	Patch(c ctx.Ctx, id Id, value PatchableNftItem) error
	IncreaseViewCount(c ctx.Ctx, id Id, count int) (int32, error)
	IncreaseLikeCount(c ctx.Ctx, id Id, count int) (int32, error)
	//	@todo	remember set IsAppropriate to true as default value
	Create(ctx.Ctx, *NftItem) error
	IncreaseSupply(c ctx.Ctx, id Id, n int) error
	DecreaseSupply(c ctx.Ctx, id Id, n int) error
}
