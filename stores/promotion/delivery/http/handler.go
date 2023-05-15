package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain/promotion"
)

type handler struct {
	promotion promotion.PromotionUsecase
}

func New(
	e *echo.Echo,
	promotion promotion.PromotionUsecase) {
	h := &handler{promotion}

	gs := e.Group("/promotions")

	gs.GET("/activated", h.getActivatedPromotions)
	//gs.POST("", h.createPromotion)
}

func (h *handler) createPromotion(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type payload struct {
		StartTime int64 `json:"startTime"`
		EndTime   int64 `json:"endTime"`
	}

	p := payload{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}
	startTs := time.Unix(p.StartTime, 0)
	endTs := time.Unix(p.EndTime, 0)
	promotion := promotion.Promotion{
		StartTime: &startTs,
		EndTime:   &endTs,
	}

	res, err := h.promotion.CreatePromotion(ctx, &promotion)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}
	return delivery.MakeJsonResp(c, http.StatusOK, res)
}

func (h *handler) getActivatedPromotions(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	p := &promotion.SearchParams{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}
	timestamp, _ := strconv.ParseInt(*p.TS, 10, 64)
	ts := time.Unix(timestamp, 0)
	res, err := h.promotion.GetActivatedPromotions(ctx, &ts)

	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}
