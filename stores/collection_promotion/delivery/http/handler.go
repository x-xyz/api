package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/collection_promotion"
)

type handler struct {
	collPromotion collection_promotion.CollPromotionUsecase
	collection    collection.Usecase
}

func New(
	e *echo.Echo,
	collPromotion collection_promotion.CollPromotionUsecase,
	collection collection.Usecase) {
	h := &handler{collPromotion, collection}

	gs := e.Group("/collection-promotions")

	gs.GET("", h.GetCollPromotions)
	gs.GET("/activated", h.GetPromotedCollections)
	gs.GET("/rewards", h.CalculateListingRewards)
	gs.GET("/last-hour-reward-per-listing", h.GetLastHourAverageRewardPerListing)
	//gs.POST("", h.createCollPromotion)

}

// func (h *handler) createCollPromotion(c echo.Context) error {
// 	ctx := c.Get("ctx").(ctx.Ctx)

// 	type payload struct {
// 		CollectionIds []string `json:"collectionIds"`
// 		PromotionId   string   `json:"promotionId"`
// 	}

// 	p := payload{}

// 	if err := c.Bind(&p); err != nil {
// 		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
// 	}

// 	err := h.collPromotion.CreateCollPromotion(ctx, &p.CollectionIds, &p.PromotionId)
// 	if err != nil {
// 		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
// 	}
// 	return delivery.MakeJsonResp(c, http.StatusOK, p)
// }

func (h *handler) GetCollPromotions(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		PromotionIds []string `query:"promotionIds"`
	}

	p := params{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}
	res, err := h.collPromotion.GetCollPromotions(ctx, &p.PromotionIds)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) GetPromotedCollections(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		TS int64 `query:"ts"`
	}

	p := params{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}
	ts := time.Now()
	if p.TS != 0 {
		ts = time.Unix(p.TS, 0)
	}
	_, promoCollections, err := h.collPromotion.GetPromotedCollections(ctx, &ts)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}
	var addresses []domain.Address
	for _, promoCollection := range promoCollections {
		addresses = append(addresses, promoCollection.Address)
	}
	opts := []collection.FindAllOptions{
		collection.WithIsAppropriate(true),
	}
	opts = append(opts, collection.WithChainId(1))
	opts = append(opts, collection.WithAddresses(addresses))

	collections, err := h.collection.FindAll(ctx, opts...)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, collections)
	}
}

func (h *handler) CalculateListingRewards(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		BeginEpoch int64  `query:"begin"`
		EndEpoch   int64  `query:"end"`
		RewardType string `query:"rewardType"`
	}

	p := params{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}
	begin := time.Unix(p.BeginEpoch, 0)
	end := time.Unix(p.EndEpoch, 0)

	var (
		res *collection_promotion.ListingRewardDistribution
		err error
	)
	if p.RewardType == "flat" {
		res, err = h.collPromotion.CalculateListingRewardsFlat(ctx, begin, end)
	} else {
		res, err = h.collPromotion.CalculateListingRewardsFixedTotal(ctx, begin, end)
	}
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) GetLastHourAverageRewardPerListing(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	rewards, err := h.collPromotion.CalculateLastHourAverageRewardPerListing(ctx)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}
	res := struct {
		Rewards string `json:"rewards"`
	}{
		Rewards: rewards,
	}
	return delivery.MakeJsonResp(c, http.StatusOK, res)
}
