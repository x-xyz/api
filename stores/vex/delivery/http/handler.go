package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
)

type handler struct {
	history domain.VexFeeDistributionHistoryUseCase
}

func New(e *echo.Echo, history domain.VexFeeDistributionHistoryUseCase) {
	h := &handler{history}
	gs := e.Group("/vex")
	gs.GET("/apr", h.getVexApr)
}

func (h *handler) getVexApr(_ctx echo.Context) error {
	ctx := _ctx.Get("ctx").(bCtx.Ctx)
	type params struct {
		Limit int `query:"limit"`
	}

	p := &params{}
	if err := _ctx.Bind(p); err != nil {
		return delivery.MakeJsonResp(_ctx, http.StatusBadRequest, "invalid params")
	}

	limit := p.Limit
	if limit == 0 {
		limit = 1
	}
	res, err := h.history.LatestApr(ctx, limit)
	if err != nil {
		return delivery.MakeJsonResp(_ctx, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(_ctx, http.StatusOK, res)
	}
}
