package order

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type PriceSource struct {
	Price        float64
	PaymentToken domain.Address
	PriceInUsd   float64
	Source       nftitem.PriceSource
	CreatedAt    time.Time
}

func ResolveLatestPrice(orders []*OrderItem) PriceSource {
	sources := []PriceSource{}
	listings := []*OrderItem{}
	offers := []*OrderItem{}

	for _, od := range orders {
		if od.IsAsk && od.IsValid {
			listings = append(listings, od)
		} else if !od.IsAsk {
			offers = append(offers, od)
		}
	}

	if len(listings) > 0 {
		sort.Slice(listings, func(i, j int) bool {
			return listings[i].PriceInUsd < listings[j].PriceInUsd
		})

		cheapest := listings[0]
		sources = append(sources, PriceSource{
			Source:       nftitem.PriceSourceListing,
			Price:        decimal.RequireFromString(cheapest.DisplayPrice).InexactFloat64(),
			PaymentToken: cheapest.Currency,
			PriceInUsd:   cheapest.PriceInUsd,
			CreatedAt:    cheapest.StartTime,
		})
	}

	// listings have higher precedence over offers, so skip this if we already have a source
	if len(sources) == 0 && len(offers) > 0 {
		sort.Slice(offers, func(i, j int) bool {
			return offers[i].PriceInUsd > offers[j].PriceInUsd
		})

		mostExpensive := offers[0]
		sources = append(sources, PriceSource{
			Source:       nftitem.PriceSourceOffer,
			Price:        decimal.RequireFromString(mostExpensive.DisplayPrice).InexactFloat64(),
			PaymentToken: mostExpensive.Currency,
			PriceInUsd:   mostExpensive.PriceInUsd,
			CreatedAt:    mostExpensive.StartTime,
		})
	}

	if len(sources) == 0 {
		return PriceSource{}
	}

	sort.Slice(sources, func(i, j int) bool {
		return sources[i].CreatedAt.After(sources[j].CreatedAt)
	})

	return sources[0]
}
