package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/external_listing"
	"github.com/x-xyz/goapi/middleware"
)

type handler struct {
	externalListing external_listing.ExternalListingUseCase
	collection      collection.Usecase
	cacheDuration   float64
}

type void struct{}

func New(
	e *echo.Echo,
	externalListing external_listing.ExternalListingUseCase,
	collection collection.Usecase,
	cacheDuration float64) {
	h := &handler{externalListing: externalListing, collection: collection, cacheDuration: cacheDuration}

	g := e.Group("/external-listings")
	g.GET("/:account", h.getListings, middleware.IsValidAddress("account"))
	g.POST("/:account/refresh", h.refreshListings, middleware.IsValidAddress("account"))
}

// getListings
//
//	@Description	Get opensea listings, should call refresh first
//	@Tags			external-listings
//	@Accept			json
//	@Produce		json
//	@Param			account	path		string	true	"address"
//	@Success		200		{object}	[]external_listing.ExternalListing
//	@Failure		500
//	@Router			/external-listings/{account} [get]
func (h *handler) getListings(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	ethereumChainId := domain.ChainId(1)
	account := domain.Address(c.Param("account"))
	listings, err := h.externalListing.GetListings(ctx, account, ethereumChainId)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, listings)
	}
}

// refreshListings
//
//	@Description	Refresh opensea listings
//	@Tags			external-listings
//	@Accept			json
//	@Produce		json
//	@Param			account	path	string	true	"address"
//	@Success		200
//	@Failure		500
//	@Router			/external-listings/{account}/refresh [post]
func (h *handler) refreshListings(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	ethereumChainId := domain.ChainId(1)
	account := domain.Address(c.Param("account"))
	oldListings, err := h.externalListing.GetListings(ctx, account, ethereumChainId)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}
	if len(oldListings) > 0 {
		t1 := time.Now()
		t2 := oldListings[0].UpdatedTime
		diff := t1.Sub(t2)
		if diff.Seconds() < h.cacheDuration {
			return delivery.MakeJsonResp(c, http.StatusOK, nil)
		}
		for _, oldListing := range oldListings {
			id := external_listing.ExternalListingId{
				Owner:   oldListing.Owner,
				ChainId: oldListing.ChainId,
			}
			err = h.externalListing.DeleteListing(ctx, id)
			if err != nil {
				return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
			}
		}
	}
	osListings, err := h.externalListing.FetchOpenseaListings(ctx, account, ethereumChainId)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}
	var collAddresses []domain.Address
	for _, osListing := range osListings {
		collAddresses = append(collAddresses, osListing.ContractAddress)
	}
	var opts []collection.FindAllOptions
	opts = append(opts, collection.WithChainId(1))
	opts = append(opts, collection.WithAddresses(collAddresses))
	collections, err := h.collection.FindAll(ctx, opts...)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}
	var exListings []external_listing.ExternalListing
	if collections.Count == 0 {
		return delivery.MakeJsonResp(c, http.StatusOK, nil)
	}
	collectionSet := make(map[domain.Address]void)
	for _, coll := range collections.Items {
		collectionSet[coll.Erc721Address] = void{}
	}
	for _, osListing := range osListings {
		if _, ok := collectionSet[osListing.ContractAddress]; ok {
			exListings = append(exListings, osListing)
		}
	}
	if len(exListings) > 0 {
		err = h.externalListing.BulkUpsert(ctx, exListings)
		if err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		}
	}
	return delivery.MakeJsonResp(c, http.StatusOK, nil)
}
