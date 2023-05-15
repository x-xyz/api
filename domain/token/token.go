package token

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SortBy string

const (
	// primitive columns
	SortByCreatedAt SortBy = "createdAt" // recently created, oldest
	SortByViewed    SortBy = "viewed"    // most viewed
	SortByLiked     SortBy = "liked"     // most liked
	SortByPrice     SortBy = "price"     // price will be updated after listed, sold, traded
	SortByLastPrice SortBy = "lastSalePrice"
	// aggrigated columns
	SortByListedAt         SortBy = "listedAt"         // recently listed
	SortBySoldAt           SortBy = "soldAt"           // recently sold
	SortByAuctionStartTime SortBy = "auctionStartTime" // recently started auction
	SortByAuctionEndTime   SortBy = "auctionEndTime"   // auction ending soon
)

type SearchOptions struct {
	SortBy                *string                   `json:"SortBy"`
	SortDir               *domain.SortDir           `json:"SortDir"`
	Sorts                 *[]string                 `json:"Sorts"`
	Offset                *int32                    `json:"Offset"`
	Limit                 *int32                    `json:"Limit"`
	SaleStatus            nftitem.SaleStatus        `json:"SaleStatus"`
	Collections           []domain.Address          `json:"Collections"`
	Category              *string                   `json:"Category"`
	ChainId               *domain.ChainId           `json:"ChainId"`
	BelongsTo             *domain.Address           `json:"BelongsTo"`
	NotBelongsTo          *domain.Address           `json:"NotBelongsTo"`
	LikedBy               *domain.Address           `json:"LikedBy"`
	Attributes            []nftitem.AttributeFilter `json:"Attributes"`
	PriceGTE              *float64                  `json:"PriceGTE"`
	PriceLTE              *float64                  `json:"PriceLTE"`
	PriceInUsdGTE         *float64                  `json:"PriceInUsdGTE"`
	PriceInUsdLTE         *float64                  `json:"PriceInUsdLTE"`
	OfferPriceInUsdGTE    *float64                  `json:"OfferPriceInUsdGTE"`
	OfferPriceInUsdLTE    *float64                  `json:"OfferPriceInUsdLTE"`
	FolderId              *string                   `json:"FolderId"`
	ListingFrom           *domain.Address           `json:"ListingFrom"`
	InactiveListingFrom   *domain.Address           `json:"InactiveListingFrom"`
	Name                  *string                   `json:"Name"`
	Search                *string                   `json:"search"`
	OfferOwners           []domain.Address          `json:"OfferOwners"`
	TokenType             *domain.TokenType         `json:"TokenType"`
	BidOwner              *domain.Address           `json:"BidOwner"`
	ObjectIdLT            *primitive.ObjectID       `json:"ObjectIdLT"`
	IncludeOrders         *bool                     `json:"IncludeOrders"`
	IncludeInactiveOrders *bool                     `json:"IncludeInactiveOrders"`
	HasOrder              *bool                     `json:"HasOrder"`
	Cursor                *string                   `json:"-"`
	Size                  *int                      `json:"-"`
}

func OptionsToKey(opts SearchOptions) string {
	key, _ := json.Marshal(opts)
	return string(key)
}

func ParseKeyToOptions(key string) (*SearchOptions, error) {
	opts := SearchOptions{}
	err := json.Unmarshal([]byte(key), &opts)
	if err != nil {
		return nil, err
	}
	return &opts, nil
}

type SearchOptionsFunc func(*SearchOptions) error

func GetSearchOptions(opts ...SearchOptionsFunc) (SearchOptions, error) {
	res := SearchOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func WithCursor(cursor string) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.Cursor = &cursor
		return nil
	}
}

func WithSize(size int) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.Size = &size
		return nil
	}
}

func WithSort(sortby string, sortdir domain.SortDir) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func WithSorts(sorts []string) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.Sorts = &sorts
		return nil
	}
}

func WithPagination(offset int32, limit int32) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func WithSaleStatus(saleStatus nftitem.SaleStatus) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.SaleStatus = saleStatus
		return nil
	}
}

func WithBuyNow() SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.SaleStatus = options.SaleStatus | nftitem.SaleStatusBuyNow
		return nil
	}
}

func WithHasBid() SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.SaleStatus = options.SaleStatus | nftitem.SaleStatusHasBid
		return nil
	}
}

func WithHasOffer() SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.SaleStatus = options.SaleStatus | nftitem.SaleStatusHasOffer
		return nil
	}
}

func WithOnAuction() SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.SaleStatus = options.SaleStatus | nftitem.SaleStatusOnAuction
		return nil
	}
}

func WithHasTraded() SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.SaleStatus = options.SaleStatus | nftitem.SaleStatusHasTraded
		return nil
	}
}

func WithCollections(collections ...domain.Address) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.Collections = collections
		return nil
	}
}

func WithCategory(category string) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.Category = &category
		return nil
	}
}

func WithChainId(chainId domain.ChainId) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func WithBelongsTo(address domain.Address) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.BelongsTo = address.ToLowerPtr()
		return nil
	}
}

func WithNotBelongsTo(address domain.Address) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.NotBelongsTo = address.ToLowerPtr()
		return nil
	}
}

func WithLikedBy(address domain.Address) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.LikedBy = &address
		return nil
	}
}

// WithAttributes ...
func WithAttributes(attributes []nftitem.AttributeFilter) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.Attributes = attributes
		return nil
	}
}

func WithPriceGTE(val float64) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.PriceGTE = &val
		return nil
	}
}

func WithPriceLTE(val float64) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.PriceLTE = &val
		return nil
	}
}

func WithPriceInUsdGTE(val float64) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.PriceInUsdGTE = &val
		return nil
	}
}

func WithPriceInUsdLTE(val float64) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.PriceInUsdLTE = &val
		return nil
	}
}

func WithOfferPriceInUsdGTE(val float64) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.OfferPriceInUsdGTE = &val
		return nil
	}
}

func WithOfferPriceInUsdLTE(val float64) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.OfferPriceInUsdLTE = &val
		return nil
	}
}

func WithFolderId(folderId string) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.FolderId = &folderId
		return nil
	}
}

func WithListingFrom(owner domain.Address) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.ListingFrom = owner.ToLowerPtr()
		return nil
	}
}

func WithInactiveListingFrom(owner domain.Address) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.InactiveListingFrom = owner.ToLowerPtr()
		return nil
	}
}

func WithName(name string) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.Name = &name
		return nil
	}
}

func WithSearch(search string) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.Search = &search
		return nil
	}
}

func WithOfferOwners(owners ...domain.Address) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.OfferOwners = owners
		return nil
	}
}

func WithTokenType(tokenType domain.TokenType) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.TokenType = &tokenType
		return nil
	}
}

func WithBidOwner(owner domain.Address) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.BidOwner = owner.ToLowerPtr()
		return nil
	}
}

func WithObjectIdLT(id primitive.ObjectID) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.ObjectIdLT = &id
		return nil
	}
}

func WithIncludeOrders(include bool) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.IncludeOrders = &include
		return nil
	}
}

func WithIncludeInactiveOrders(include bool) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.IncludeInactiveOrders = &include
		return nil
	}
}

func WithHasOrder(hasOrder bool) SearchOptionsFunc {
	return func(options *SearchOptions) error {
		options.HasOrder = &hasOrder
		return nil
	}
}

type SearchResult struct {
	Items      []*TokenWithDetail `json:"items"`
	Count      int                `json:"count"`
	NextCursor string             `json:"nextCursor"`
}

type ActivityResult struct {
	Items []account.ActivityHistory `json:"items"`
	Count int                       `json:"count"`
}

type TokenWithDetail struct {
	nftitem.NftItem
	// inject in get tokens, token handler only if query with auth token
	IsLiked *bool `json:"isLiked,omitempty"`
	// inject in get token handler
	HasUnlockable *bool `json:"hasUnlockable,omitempty"`
	// inject balance only if token is queried by belongsTo and token type if 1155
	Balance *int `json:"balance,omitempty"`

	Listings         []*order.OrderItem `json:"listings,omitempty"`
	InactiveListings []*order.OrderItem `json:"inactiveListings,omitempty"`
	Offers           []*order.OrderItem `json:"offers,omitempty"`
	ActiveListing    *order.OrderItem   `json:"activeListing,omitempty"`
}

type UploadPayload struct {
	Name           string        `json:"name" form:"name"`
	Description    string        `json:"description" form:"description"`
	XtraUrl        string        `json:"xtra" form:"xtra"`
	Image          string        `json:"image" form:"image"`
	Royalty        string        `json:"royalty" form:"royalty"`
	CollectionName string        `json:"collectionName" form:"collectionName"`
	Traits         []*TokenTrait `json:"traits,omitempty" form:"traits"`
}

type TokenTrait struct {
	Type  string `json:"trait_type" form:"type"`
	Value string `json:"value" form:"value"`
}

type UploadResult struct {
	FileHash string `json:"fileHash"`
	JsonHash string `json:"jsonHash"`
}

type PriceHistory struct {
	PriceInUsd    float64   `json:"priceInUsd"`
	PriceInNative float64   `json:"priceInNative"`
	Time          time.Time `json:"time"`
}

type Usecase interface {
	Search(c ctx.Ctx, opts ...SearchOptionsFunc) (*SearchResult, error)
	SearchV2(c ctx.Ctx, opts ...SearchOptionsFunc) (*SearchResult, error)
	FindOne(c ctx.Ctx, id nftitem.Id) (*TokenWithDetail, error)
	GetActivities(c ctx.Ctx, id nftitem.Id, offet, limit int) (*ActivityResult, error)
	GetPriceHistories(c ctx.Ctx, id nftitem.Id, period domain.TimePeriod) ([]PriceHistory, error)
	AddUnlockableContent(c ctx.Ctx, id nftitem.Id, content string) error
	GetUnlockableContent(c ctx.Ctx, id nftitem.Id) (string, error)
	BanNftItem(c ctx.Ctx, id nftitem.Id) error
	UnbanNftItem(c ctx.Ctx, id nftitem.Id) error
	Upload(c ctx.Ctx, account domain.Address, payload UploadPayload) (*UploadResult, error)
	PatchNft(ctx.Ctx, *nftitem.Id, *nftitem.PatchableNftItem) error
	SearchForIndexerState(c ctx.Ctx, indexerStates []nftitem.IndexerState, retryCountLimit int, opts ...SearchOptionsFunc) ([]*nftitem.NftItem, int, error)
	// also increase view count
	GetViewCount(c ctx.Ctx, id nftitem.Id) (int32, error)
	UpsertListing(c ctx.Ctx, id nftitem.Id, listing *nftitem.Listing, overrideActive bool) error
	RemoveListing(c ctx.Ctx, id nftitem.Id, owner *domain.Address) error
	SetActiveListingTo(c ctx.Ctx, id nftitem.Id, owner *domain.Address) error
	GetListing(c ctx.Ctx, id nftitem.Id, owner *domain.Address) (*nftitem.Listing, error)
	UpsertOffer(c ctx.Ctx, id nftitem.Id, offer *nftitem.Offer) error
	RemoveOffer(c ctx.Ctx, id nftitem.Id, offerer *domain.Address) error
	GetOffer(c ctx.Ctx, id nftitem.Id, offerer *domain.Address) (*nftitem.Offer, error)
	SetAuction(c ctx.Ctx, id nftitem.Id, auction *nftitem.Auction) error
	UpdateAuction(c ctx.Ctx, id nftitem.Id, auction *nftitem.Auction) error
	ClearAuction(c ctx.Ctx, id nftitem.Id) error
	SetHighestBid(c ctx.Ctx, id nftitem.Id, bid *nftitem.Bid) error
	ClearHighestBid(c ctx.Ctx, id nftitem.Id) error
	EnsureNftExists(c ctx.Ctx, id nftitem.Id) (*nftitem.NftItem, error)
	RefreshIndexerState(c ctx.Ctx, id nftitem.Id) error
	RefreshListingAndOfferState(ctx ctx.Ctx, id nftitem.Id) error
	GetOpenRararityScore(ctx ctx.Ctx, id nftitem.Id) (float64, error)
}

func ToTokenKey(chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) string {
	return strings.Join([]string{strconv.Itoa(int(chainId)), contract.ToLowerStr(), tokenId.String()}, ":")
}

func FromTokenKey(key string) (domain.ChainId, domain.Address, domain.TokenId, error) {
	parts := strings.Split(key, ":")
	chainId, err := strconv.Atoi(parts[0])
	if err != nil {
		return domain.ChainId(0), "", "", err
	}
	tokenId := parts[2]
	return domain.ChainId(chainId), domain.Address(parts[1]), domain.TokenId(tokenId), nil
}
