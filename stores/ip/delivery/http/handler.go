package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/ip"
	authMiddleware "github.com/x-xyz/goapi/stores/auth/delivery/http/middleware"
)

type handler struct {
	ipUseCase ip.UseCase
	au        account.Usecase
}

func New(
	e *echo.Echo,
	ipUseCase ip.UseCase,
	au account.Usecase,
	authMiddleware *authMiddleware.AuthMiddleware,
) {
	h := &handler{ipUseCase, au}

	gs := e.Group("/ip/listings")

	gs.GET("", h.search)

	gs.POST("", h.createListing, authMiddleware.Auth())

	g := e.Group("/ip/listing/:id")

	g.DELETE("", h.deleteListing, authMiddleware.Auth())
}

// search
//
//	@Summary		Search for ip listings
//	@Description	Search for ip listings
//	@Tags			ips
//	@Accept			json
//	@Produce		json
//	@Param			chainId			query		int		true	"chain id. e.g: `1` for ethereum"	example(1)
//	@Param			contractAddress	query		string	true	"NFT collection contract address"	example(0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d)
//	@Param			isIpOwner		query		bool	false	"true for ip listings, false for offers"
//	@Success		200				{object}	[]ip.IPListing
//	@Failure		400
//	@Failure		500
//	@Router			/ip/listings [get]
func (h *handler) search(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		IsIpOwner       *bool           `query:"isIpOwner"`
		ChainId         *domain.ChainId `query:"chainId"`
		ContractAddress *domain.Address `query:"contractAddress"`
		YugaLab         bool            `query:"yugaLab"`
	}

	p := params{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	opts := []ip.FindAllOptionsFunc{}

	if p.IsIpOwner != nil {
		opts = append(opts, ip.WithIsIpOwner(*p.IsIpOwner))
	}

	if p.ChainId != nil && p.ContractAddress != nil {
		opts = append(opts, ip.WithChainId(*p.ChainId))
		opts = append(opts, ip.WithContractAddresses([]domain.Address{*p.ContractAddress}))
	} else if p.YugaLab {
		opts = append(opts, ip.WithChainId(1))
		opts = append(opts, ip.WithContractAddresses(ip.YugaLabIpCollectionAddresses))
	}

	res, err := h.ipUseCase.FindAll(ctx, opts...)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, res)
}

// createListing
//
//	@Summary		Create new ip listing
//	@Description	Create new ip listing
//	@Tags			ips
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			params	body	http.createListing.params	true	"params"
//	@Success		200
//	@Failure		400
//	@Failure		500
//	@Router			/ip/listings [post]
func (h *handler) createListing(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		ip.IPListing
		// signature of message retrieved from /auth/signingMsgTemplate
		Signature string `json:"signature"`
	}

	p := params{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	// inject address to owner, since only signer can create listing
	address := c.Get("address").(domain.Address)
	p.Owner = address.ToLower()

	if err := c.Validate(p.IPListing); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.au.ValidateSignature(ctx, p.Owner, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	if err := h.ipUseCase.Insert(ctx, &p.IPListing); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, p.IPListing)
}

// deleteListing
//
//	@Summary		Delete listing
//	@Description	Delete listing
//	@Tags			ips
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			params	body	http.deleteListing.params	true	"params"
//	@Success		200
//	@Failure		400
//	@Failure		500
//	@Router			/ip/listings/{id} [delete]
func (h *handler) deleteListing(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	address := c.Get("address").(domain.Address)

	type params struct {
		Id string `param:"id" validate:"required"`
		// signature of message retrieved from /auth/signingMsgTemplate
		Signature string `json:"signature"`
	}

	p := params{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	listing, err := h.ipUseCase.FindOne(ctx, p.Id)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if err := h.au.ValidateSignature(ctx, listing.Owner, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	if listing.Owner != address {
		return delivery.MakeJsonResp(c, http.StatusUnauthorized, nil)
	}
	if err := h.ipUseCase.Delete(ctx, p.Id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, p.Id)
}
