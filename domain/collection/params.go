package collection

import (
	"sort"

	"github.com/x-xyz/goapi/domain"
)

type SearchParams struct {
	// required
	SortBy    SearchSortOption `query:"sortBy"`
	Offset    int32            `query:"offset"`
	Limit     int32            `query:"limit"`
	ChainId   *domain.ChainId  `query:"chainId"`
	Category  *string          `query:"category"`
	BelongsTo *domain.Address  `query:"belongsTo"`
	// floor price greater than or equal
	FloorPriceGTE *float64 `query:"floorPriceGTE"`
	// floor price less than or equal
	FloorPriceLTE *float64 `query:"floorPriceLTE"`
	// floor price greater than or equal
	UsdFloorPriceGTE *float64 `query:"usdFloorPriceGTE"`
	// floor price less than or equal
	UsdFloorPriceLTE    *float64 `query:"usdFloorPriceLTE"`
	IncludeUnregistered *bool    `query:"includeUnregistered"`
	// for backward compatible. with this flag, handler with return result with total count
	IsPaging  *bool           `query:"isPaging"`
	Holder    *domain.Address `query:"holder"`
	YugaLab   bool            `query:"yugaLab"`
	LikedBy   *domain.Address `query:"likedBy"`
	ListedBy  *domain.Address `query:"listedBy"`
	OfferedBy *domain.Address `query:"offeredBy"`
}

type SearchSortOption = string

const (
	SearchSortOptionListedAtDesc   = "listed_at_high_to_low"
	SearchSortOptionSoldAtDesc     = "sold_at_high_to_low"
	SearchSortOptionFloorPriceAsc  = "floor_price_low_to_high"
	SearchSortOptionFloorPriceDesc = "floor_price_high_to_low"
	SearchSortOptionViewedDesc     = "viewed_high_to_low"
	SearchSortOptionLikedDesc      = "liked_high_to_low"
	SearchSortOptionCreatedAtAsc   = "created_at_low_to_high"
	SearchSortOptionCreatedAtDesc  = "created_at_high_to_low"
	SearchSortOptionHoldingAsc     = "holding_low_to_high"
	SearchSortOptionHoldingDesc    = "holding_high_to_low"
	SearchSortOptionNameAsc        = "name_low_to_high"
	SearchSortOptionNameDesc       = "name_high_to_low"
	// unsupported options
	// SearchSortOptionTradingVolumeDesc = "trading_volume_high_to_low"
	// SearchSortOptionAuctionEndingSoon = "auction_ending_soon"
	// SearchSortOptionLastSalePriceAsc  = "last_sale_low_to_high"
	// SearchSortOptionLastSalePriceDesc = "last_sale_high_to_low"
)

func ParseSearchSortOption(value SearchSortOption) (string, domain.SortDir) {
	switch value {
	case SearchSortOptionListedAtDesc:
		return "lastListedAt", domain.SortDirDesc
	case SearchSortOptionSoldAtDesc:
		return "lastSoldAt", domain.SortDirDesc
	case SearchSortOptionFloorPriceAsc:
		return "usdFloorPrice", domain.SortDirAsc
	case SearchSortOptionFloorPriceDesc:
		return "usdFloorPrice", domain.SortDirDesc
	case SearchSortOptionViewedDesc:
		return "viewCount", domain.SortDirDesc
	case SearchSortOptionLikedDesc:
		return "liked", domain.SortDirDesc
	case SearchSortOptionCreatedAtAsc:
		return "_id", domain.SortDirAsc
	case SearchSortOptionCreatedAtDesc:
		return "_id", domain.SortDirDesc
	case SearchSortOptionNameAsc:
		return "collectionName", domain.SortDirAsc
	case SearchSortOptionNameDesc:
		return "collectionName", domain.SortDirDesc
	}
	return "lastListedAt", domain.SortDirDesc
}

func SortCollectionWithHoldingCount(arr []*CollectionWithHoldingCount, sortOption SearchSortOption) {
	var less func(i, j int) bool
	switch sortOption {
	case SearchSortOptionHoldingAsc:
		less = func(i, j int) bool {
			return arr[i].HoldingBalance < arr[j].HoldingBalance
		}
	case SearchSortOptionHoldingDesc:
		less = func(i, j int) bool {
			return arr[i].HoldingBalance > arr[j].HoldingBalance
		}
	default:
		return
	}
	sort.Slice(arr, less)
}
