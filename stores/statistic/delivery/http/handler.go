package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain/statistic"
)

type handler struct {
	statisticUC statistic.UseCase
}

func New(e *echo.Echo, statisticUC statistic.UseCase) {
	h := &handler{statisticUC}
	gs := e.Group("/statistics")
	gs.GET("/apeburned", h.getApeBurned)
}

func (h *handler) getApeBurned(_ctx echo.Context) error {
	ctx := _ctx.Get("ctx").(ctx.Ctx)
	value, err := h.statisticUC.Get(ctx, statistic.Apeburned)
	if err != nil {
		return delivery.MakeJsonResp(_ctx, http.StatusInternalServerError, err)
	}
	apeburned, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return delivery.MakeJsonResp(_ctx, http.StatusInternalServerError, err)
	}
	res := struct {
		Apeburned float64 `json:"apeburned"`
	}{
		Apeburned: apeburned,
	}
	return delivery.MakeJsonResp(_ctx, http.StatusOK, res)
}
