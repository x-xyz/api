package token

import (
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

// SearchParams
type SearchParams struct {
	Offset      int32                    `query:"offset"`
	Limit       int32                    `query:"limit"`
	SortBy      SearchSortOption         `query:"sortBy"`
	SaleStatus  []nftitem.SaleStatusType `query:"saleStatus"`
	Collections []domain.Address         `query:"collections"`
	Category    *string                  `query:"category"`
	ChainId     *domain.ChainId          `query:"chainId"`
	// `BelongsTo` and `NotBelongsTo` are exclusive options
	BelongsTo *domain.Address `query:"belongsTo"`
	// `BelongsTo` and `NotBelongsTo` are exclusive options
	NotBelongsTo          *domain.Address   `query:"notBelongsTo"`
	LikedBy               *domain.Address   `query:"likedBy"`
	AttrFilters           []string          `query:"attrFilters"`
	PriceGTE              *float64          `query:"priceGTE"`
	PriceLTE              *float64          `query:"priceLTE"`
	PriceInUsdGTE         *float64          `query:"priceInUsdGTE"`
	PriceInUsdLTE         *float64          `query:"priceInUsdLTE"`
	OfferPriceInUsdGTE    *float64          `query:"offerPriceInUsdGTE"`
	OfferPriceInUsdLTE    *float64          `query:"offerPriceInUsdLTE"`
	FolderId              *string           `query:"folderId"`
	ListingFrom           *domain.Address   `query:"listingFrom"`
	InactiveListingFrom   *domain.Address   `query:"inactiveListingFrom"`
	Name                  *string           `query:"name"`
	Search                *string           `query:"search"`
	OfferOwners           []domain.Address  `query:"offerOwners"`
	TokenType             *domain.TokenType `query:"tokenType"`
	BidOwner              *domain.Address   `query:"bidOwner"`
	IncludeOrders         *bool             `query:"includeOrders"`
	IncludeInactiveOrders *bool             `query:"includeInactiveOrders"`
	// if cursor != nil, search result will store in cache
	// if cursor == nil, search result will bypass cache
	Cursor *string `query:"cursor"`
	// Size will be ignored if Cursor == nil
	Size *int `query:"size"`
}

type SearchSortOption = string

const (
	SearchSortOptionListedAtDesc       = "listed_at_high_to_low"
	SearchSortOptionSoldAtDesc         = "sold_at_high_to_low"
	SearchSortOptionPriceAsc           = "price_low_to_high"
	SearchSortOptionPriceDesc          = "price_high_to_low"
	SearchSortOptionViewedDesc         = "viewed_high_to_low"
	SearchSortOptionLikedDesc          = "liked_high_to_low"
	SearchSortOptionCreatedAtAsc       = "created_at_low_to_high"
	SearchSortOptionCreatedAtDesc      = "created_at_high_to_low"
	SearchSortOptionAuctionEndingSoon  = "auction_ending_soon"
	SearchSortOptionLastSalePriceAsc   = "last_sale_low_to_high"
	SearchSortOptionLastSalePriceDesc  = "last_sale_high_to_low"
	SearchSortOptionOfferPriceAsc      = "offer_price_low_to_high"
	SearchSortOptionOfferPriceDesc     = "offer_price_high_to_low"
	SearchSortOptionOfferExpiredSoon   = "offer_deadline_expired_soon"
	SearchSortOptionOfferCreatedAtAsc  = "offer_created_at_low_to_high"
	SearchSortOptionOfferCreatedAtDesc = "offer_created_at_high_to_low"
)

func ParseSearchSortOption(value SearchSortOption) (string, domain.SortDir, []SearchOptionsFunc) {
	switch value {
	case SearchSortOptionSoldAtDesc:
		return "soldAt", domain.SortDirDesc, []SearchOptionsFunc{WithHasTraded()}
	case SearchSortOptionPriceAsc:
		return "priceInUSD", domain.SortDirAsc, nil
	case SearchSortOptionPriceDesc:
		return "priceInUSD", domain.SortDirDesc, nil
	case SearchSortOptionViewedDesc:
		return "viewed", domain.SortDirDesc, nil
	case SearchSortOptionLikedDesc:
		return "liked", domain.SortDirDesc, nil
	case SearchSortOptionCreatedAtAsc:
		return "_id", domain.SortDirAsc, nil
	case SearchSortOptionCreatedAtDesc:
		return "_id", domain.SortDirDesc, nil
	case SearchSortOptionAuctionEndingSoon:
		// return "saleEndsAt", domain.SortDirDesc, []SearchOptionsFunc{WithOnAuction()}
		//	@todo	guard with `WithOnAuction` once implemented
		return "saleEndsAt", domain.SortDirDesc, nil
	case SearchSortOptionLastSalePriceAsc:
		return "lastSalePriceInUSD", domain.SortDirAsc, []SearchOptionsFunc{WithHasTraded()}
	case SearchSortOptionLastSalePriceDesc:
		return "lastSalePriceInUSD", domain.SortDirDesc, []SearchOptionsFunc{WithHasTraded()}
	case SearchSortOptionOfferPriceAsc:
		return "instantLiquidityInUsd", domain.SortDirAsc, nil
	case SearchSortOptionOfferPriceDesc:
		return "instantLiquidityInUsd", domain.SortDirDesc, nil
	case SearchSortOptionOfferExpiredSoon:
		return "offerEndsAt", domain.SortDirAsc, nil
	case SearchSortOptionOfferCreatedAtAsc:
		return "offerStartsAt", domain.SortDirAsc, nil
	case SearchSortOptionOfferCreatedAtDesc:
		return "offerStartsAt", domain.SortDirDesc, nil
	case SearchSortOptionListedAtDesc:
		return "listedAt", domain.SortDirDesc, nil
	default:
		return "_id", domain.SortDirDesc, nil
	}
}
