package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/moderator"
	authMiddleware "github.com/x-xyz/goapi/stores/auth/delivery/http/middleware"
)

type handler struct {
	moderator moderator.Usecase
	account   account.Usecase
}

func New(e *echo.Echo, moderator moderator.Usecase, account account.Usecase, authMiddleware *authMiddleware.AuthMiddleware) {
	h := &handler{moderator, account}

	e.GET("/moderators", h.getAll, authMiddleware.Auth(), authMiddleware.IsAdmin())

	e.POST("/moderators/add", h.add, authMiddleware.Auth(), authMiddleware.IsAdmin())

	e.POST("/moderators/remove", h.remove, authMiddleware.Auth(), authMiddleware.IsAdmin())
}

func (h *handler) getAll(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	if res, err := h.moderator.FindAll(ctx); err != nil {
		ctx.WithField("err", err).Error("moderator.FindAll failed")
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) add(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	signer := c.Get("address").(domain.Address)

	type payload struct {
		Name      string         `json:"name"`
		Address   domain.Address `json:"address"`
		Signature string         `json:"signature"`
	}

	p := &payload{}

	if err := c.Bind(p); err != nil {
		ctx.WithField("err", err).Error("c.Bind failed")
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.account.ValidateSignature(ctx, signer, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	if err := h.moderator.Add(ctx, p.Address, p.Name); err != nil {
		ctx.WithField("err", err).Error("moderator.Add failed")
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusCreated, nil)
}

func (h *handler) remove(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	signer := c.Get("address").(domain.Address)

	type payload struct {
		Address   domain.Address `param:"address"`
		Signature string         `json:"signature"`
	}

	p := &payload{}

	if err := c.Bind(p); err != nil {
		ctx.WithField("err", err).Error("c.Bind failed")
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.account.ValidateSignature(ctx, signer, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	if err := h.moderator.Remove(ctx, p.Address); err != nil {
		ctx.WithField("err", err).Error("moderator.Remove failed")
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusCreated, nil)
}
