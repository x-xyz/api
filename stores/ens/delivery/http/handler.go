package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/ens"
)

type handler struct {
	ens ens.ENS
}

func New(e *echo.Echo, ens ens.ENS) {
	h := &handler{
		ens,
	}

	g := e.Group("ens")

	g.GET("/resolve/:name", h.Resolve)

	g.GET("/reverse-resolve/:address", h.ReverseResolve)
}

func (h *handler) Resolve(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type payload struct {
		Name string `param:"name" validate:"required"`
	}

	p := payload{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	address, err := h.ens.Resolve(ctx, p.Name)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, address)
}

func (h *handler) ReverseResolve(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type payload struct {
		Address domain.Address `param:"address" validate:"required"`
	}

	p := payload{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	name, err := h.ens.ReverseResolve(ctx, p.Address)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, name)
}
