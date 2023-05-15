package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
)

type authHandler struct {
	auth               domain.AuthUsecase
	signingMsgTemplate string
}

func New(e *echo.Echo, auth domain.AuthUsecase, template string) {
	handler := &authHandler{
		auth:               auth,
		signingMsgTemplate: template,
	}
	g := e.Group("/auth")
	g.POST("/sign", handler.sign)
	g.GET("/signingMsgTemplate", handler.getSigningMsgTemplate)
}

// sign
//
//	@Summary		Get access token
//	@Description	Create access token for given address
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			params	body		http.sign.params	true	"params"
//	@Success		201		{object}	object{data=string}
//	@Failure		400
//	@Failure		500
//	@Router			/auth/sign [post]
func (h *authHandler) sign(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		Address domain.Address `json:"address" binding:"address" description:"account address" example:"0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"` // account address
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		ctx.WithField("err", err).Error("bind failed")
		return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if tkn, err := h.auth.SignToken(ctx, p.Address); err != nil {
		ctx.WithField("err", err).Error("auth.SignToken failed")
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, tkn)
	}
}

// getSigningMsgTemplate
//
//	@Summary		Get signature template
//	@Description	Replace %s with nonce fetched from /account/nonce to build signing message
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	object{msg=string}	"signing message template"
//	@Router			/auth/signingMsgTemplate [get]
func (h *authHandler) getSigningMsgTemplate(c echo.Context) error {
	res := struct {
		Msg string `json:"template"`
	}{
		Msg: h.signingMsgTemplate,
	}
	return delivery.MakeJsonResp(c, http.StatusOK, res)
}
