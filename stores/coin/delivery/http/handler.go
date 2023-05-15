package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/service/coingecko"
)

type handler struct {
	client coingecko.Client
}

func New(e *echo.Echo, coingeckoClient coingecko.Client) {
	h := &handler{
		client: coingeckoClient,
	}

	g := e.Group("/coin")
	g.GET("/:coinId", h.getCoin)
}

func (h *handler) getCoin(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	p := struct {
		CoinId string `param:"coinId" validate:"required"`
	}{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	val, err := h.client.GetPrice(ctx, p.CoinId)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, val)
}
