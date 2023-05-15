package nftitem

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
)

type PriceSource string

const (
	PriceSourceListing        = "listing"
	PriceSourceOffer          = "offer"
	PriceSourceAuctionReserve = "auction_reserve"
	PriceSourceAuctionBid     = "auction_bid"
)

func ResolveLatestPrice(listings []Listing, offers []Offer, auction *Auction, bid *Bid) (price *float64, paymentToken *domain.Address, priceInUsd *float64, priceSource *PriceSource) {
	type source struct {
		price        float64
		paymentToken domain.Address
		priceInUsd   float64
		source       PriceSource
		createdAt    time.Time
	}

	sources := []source{}

	if len(listings) > 0 {
		sort.Slice(listings, func(i, j int) bool {
			return listings[i].PriceInUsd < listings[j].PriceInUsd
		})

		cheapest := listings[0]
		sources = append(sources, source{
			source:       PriceSourceListing,
			price:        decimal.RequireFromString(cheapest.DisplayPrice).InexactFloat64(),
			paymentToken: cheapest.PayToken,
			priceInUsd:   cheapest.PriceInUsd,
			createdAt:    *cheapest.StartingTime,
		})
	}

	// listings have higher precedence over offers, so skip this if we already have a source
	if len(sources) == 0 && len(offers) > 0 {
		sort.Slice(offers, func(i, j int) bool {
			return offers[i].PriceInUsd > offers[j].PriceInUsd
		})

		mostExpensive := offers[0]
		sources = append(sources, source{
			source:       PriceSourceOffer,
			price:        decimal.RequireFromString(mostExpensive.DisplayPrice).InexactFloat64(),
			paymentToken: mostExpensive.PayToken,
			priceInUsd:   mostExpensive.PriceInUsd,
			createdAt:    *mostExpensive.CreatedAt,
		})
	}

	if auction != nil && auction.BlockNumber != 0 {
		sources = append(sources, source{
			source:       PriceSourceAuctionReserve,
			price:        decimal.RequireFromString(auction.DisplayPrice).InexactFloat64(),
			paymentToken: auction.PayToken,
			priceInUsd:   auction.PriceInUsd,
			createdAt:    *auction.StartTime,
		})

		if bid != nil && bid.BlockNumber != 0 {
			sources = append(sources, source{
				source:       PriceSourceAuctionBid,
				price:        decimal.RequireFromString(bid.DisplayPrice).InexactFloat64(),
				paymentToken: bid.PayToken,
				priceInUsd:   bid.PriceInUsd,
				createdAt:    *bid.BidTime,
			})
		}
	}

	if len(sources) == 0 {
		price = ptr.Float64(0)
		paymentToken = (*domain.Address)(ptr.String(""))
		priceInUsd = ptr.Float64(0)
		priceSource = (*PriceSource)(ptr.String(""))
		return
	}

	sort.Slice(sources, func(i, j int) bool {
		return sources[i].createdAt.After(sources[j].createdAt)
	})

	price = &sources[0].price
	paymentToken = &sources[0].paymentToken
	priceInUsd = &sources[0].priceInUsd
	priceSource = &sources[0].source
	return
}
