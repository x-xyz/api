package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
	dAirdrop "github.com/x-xyz/goapi/domain/airdrop"
)

type handler struct {
	airdrop dAirdrop.AirdropUseCase
	proof   dAirdrop.ProofUseCase
}

func New(e *echo.Echo, _airdrop dAirdrop.AirdropUseCase, _proof dAirdrop.ProofUseCase) {
	h := &handler{_airdrop, _proof}
	e.GET("/airdrops", h.getAirdrops)
	e.GET("/proofs", h.getProofs)
}

func (h *handler) getAirdrops(_ctx echo.Context) error {
	ctx := _ctx.Get("ctx").(bCtx.Ctx)
	type params struct {
		SortBy  *string         `query:"sortBy"`
		SortDir *domain.SortDir `query:"sortDir"`
		Offset  int32           `query:"offset"`
		Limit   int32           `query:"limit"`
	}

	p := &params{}
	if err := _ctx.Bind(p); err != nil {
		return delivery.MakeJsonResp(_ctx, http.StatusBadRequest, "invalid params")
	}

	opts := []dAirdrop.AirdropFindAllOptionsFunc{
		dAirdrop.AirdropWithPagination(p.Offset, p.Limit),
	}
	if p.SortBy != nil && p.SortDir != nil {
		opts = append(opts, dAirdrop.AirdropWithSort(*p.SortBy, *p.SortDir))
	}

	res, err := h.airdrop.FindAll(ctx, opts...)
	if err != nil {
		return delivery.MakeJsonResp(_ctx, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(_ctx, http.StatusOK, res)
	}
}

func (h *handler) getProofs(_ctx echo.Context) error {
	ctx := _ctx.Get("ctx").(bCtx.Ctx)
	type params struct {
		SortBy          *string         `query:"sortBy"`
		SortDir         *domain.SortDir `query:"sortDir"`
		Offset          int32           `query:"offset"`
		Limit           int32           `query:"limit"`
		ChainId         *domain.ChainId `query:"chainId"`
		ContractAddress *domain.Address `query:"contractAddress"`
		Claimer         *domain.Address `query:"claimer"`
	}

	p := &params{}
	if err := _ctx.Bind(p); err != nil {
		return delivery.MakeJsonResp(_ctx, http.StatusBadRequest, "invalid params")
	}

	opts := []dAirdrop.ProofFindAllOptionsFunc{
		dAirdrop.ProofWithPagination(p.Offset, p.Limit),
	}
	if p.SortBy != nil && p.SortDir != nil {
		opts = append(opts, dAirdrop.ProofWithSort(*p.SortBy, *p.SortDir))
	}
	if p.ChainId != nil {
		opts = append(opts, dAirdrop.ProofWithChainId(*p.ChainId))
	}
	if p.ContractAddress != nil {
		opts = append(opts, dAirdrop.ProofWithContractAddress(*p.ContractAddress))
	}
	if p.Claimer != nil {
		opts = append(opts, dAirdrop.ProofWithClaimer(*p.Claimer))
	}

	res, err := h.proof.FindAll(ctx, opts...)
	if err != nil {
		return delivery.MakeJsonResp(_ctx, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(_ctx, http.StatusOK, res)
	}
}
