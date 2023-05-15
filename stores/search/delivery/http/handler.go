package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/search"
)

type handler struct {
	search search.Usecase
}

func New(e *echo.Echo, search search.Usecase) {
	h := &handler{search: search}

	g := e.Group("/search")

	g.GET("", h.searchAll)

	g.GET("/accounts", h.searchAccounts)

	g.GET("/collections", h.searchCollections)

	g.GET("/tokens", h.searchTokens)
}

func (h *handler) searchAll(c echo.Context) error {
	type params struct {
		Keyword     string           `query:"keyword"`
		Filter      []string         `query:"filter"`
		Collections []domain.Address `query:"collections"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	if res, err := h.search.Search(ctx, p.Keyword, p.Filter, p.Collections); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) searchAccounts(c echo.Context) error {
	type params struct {
		Keyword string `query:"keyword"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	if res, err := h.search.SearchAccounts(ctx, p.Keyword); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) searchCollections(c echo.Context) error {
	type params struct {
		Keyword     string           `query:"keyword"`
		Collections []domain.Address `query:"collections"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	if res, err := h.search.SearchCollections(ctx, p.Keyword, p.Collections); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) searchTokens(c echo.Context) error {
	type params struct {
		Keyword     string           `query:"keyword"`
		Collections []domain.Address `query:"collections"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	if res, err := h.search.SearchTokens(ctx, p.Keyword, p.Collections); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}
