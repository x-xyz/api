package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain/twelvefold"
)

type handler struct {
	twelvefold twelvefold.TwelvefoldUsecase
}

func New(
	e *echo.Echo,
	twelvefold twelvefold.TwelvefoldUsecase) {
	h := &handler{twelvefold}

	e.GET("/twelvefold", h.getTwelvefold)
}

func (h *handler) getTwelvefold(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		Edition string `query:"edition"`
		Series  string `query:"series"`
		Season  string `query:"season"`
		Limit   int32  `query:"limit"`
		Offset  int32  `query:"offset"`
	}

	p := params{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}
	if p.Limit == 0 {
		p.Limit = 20
	}
	opts := []twelvefold.TwelvefoldFindAllOptionsFunc{
		twelvefold.TwelvefoldWithPagination(p.Offset, p.Limit),
	}

	if p.Edition != "" {
		opts = append(opts, twelvefold.TwelvefoldWithEdition(*&p.Edition))
	}
	if p.Series != "" {
		opts = append(opts, twelvefold.TwelvefoldWithSeries(*&p.Series))
	}
	if p.Season != "" {
		opts = append(opts, twelvefold.TwelvefoldWithSeason(*&p.Season))
	}
	if p.Edition == "" && p.Series == "" && p.Season == "" {
		opts = append(opts, twelvefold.TwelvefoldWithDummy(true))
	}

	res, err := h.twelvefold.FindAll(ctx, opts...)

	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}
