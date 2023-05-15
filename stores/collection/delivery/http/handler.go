package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/like"
	"github.com/x-xyz/goapi/middleware"
	authMiddleware "github.com/x-xyz/goapi/stores/auth/delivery/http/middleware"
)

var met metrics.Service

type handler struct {
	account        account.Usecase
	collection     collection.Usecase
	authMiddleware *authMiddleware.AuthMiddleware
	like           like.CollectionLikeUsecase
	tradingVolume  collection.TradingVolumeUseCase
}

func New(
	e *echo.Echo,
	account account.Usecase,
	collection collection.Usecase,
	authMiddleware *authMiddleware.AuthMiddleware,
	like like.CollectionLikeUsecase,
	tradingVolume collection.TradingVolumeUseCase) {
	met = metrics.New("collection")

	h := &handler{account, collection, authMiddleware, like, tradingVolume}

	gs := e.Group("/collections")

	gs.GET("", h.getAll, middleware.CacheHttp(30*time.Second))

	gs.GET("/mintable", h.getMintable, authMiddleware.Auth())

	gs.GET("/unreviewed", h.getUnreviewed, authMiddleware.Auth(), authMiddleware.IsAdmin())

	gs.POST("", h.register, authMiddleware.Auth())

	gs.GET("/top", h.getTopCollections, middleware.CacheHttp(1*time.Minute))

	gs.GET("/editable", h.getEditable, authMiddleware.Auth())

	g := e.Group("/collection/:chainId/:contract")

	g.GET("", h.get, h.collectionRequestCount(), middleware.CacheHttp(1*time.Minute))

	g.GET("/view-count", h.getViewCount)

	g.POST("/like", h.likeCollection, authMiddleware.Auth())

	g.DELETE("/like", h.unlikeCollection, authMiddleware.Auth())

	g.POST("/review", h.review, authMiddleware.Auth(), authMiddleware.IsAdmin())

	g.POST("/ban", h.ban, authMiddleware.Auth(), authMiddleware.IsModerator())

	g.POST("/unban", h.unban, authMiddleware.Auth(), authMiddleware.IsAdmin())

	g.GET("/volume", h.getVolume)

	g.PUT("/info", h.updateInfo, authMiddleware.Auth())

	g.PUT("/trait-floor", h.updateTraitFloor)

	g.GET("/activities", h.getActivities)

	g.GET("/globalofferstat", h.getGlobalOfferStat)
}

func (h *handler) getAll(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	p := &collection.SearchParams{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}

	sortBy, sortDir := collection.ParseSearchSortOption(p.SortBy)

	opts := []collection.FindAllOptions{
		collection.WithIsAppropriate(true),
		collection.WithSort(sortBy, sortDir),
	}

	if p.Offset != 0 || p.Limit != 0 {
		opts = append(opts, collection.WithPagination(p.Offset, p.Limit))
	} else {
		// for backward compatible
		opts = append(opts, collection.WithPagination(0, 5000))
	}

	if p.ChainId != nil {
		opts = append(opts, collection.WithChainId(*p.ChainId))
	}

	if p.Category != nil {
		opts = append(opts, collection.WithCategory(*p.Category))
	}

	if p.BelongsTo != nil {
		opts = append(opts, collection.WithOwner(*p.BelongsTo))
	}

	if p.FloorPriceGTE != nil {
		opts = append(opts, collection.WithFloorPriceGTE(*p.FloorPriceGTE))
	}

	if p.FloorPriceLTE != nil {
		opts = append(opts, collection.WithFloorPriceLTE(*p.FloorPriceLTE))
	}

	if p.UsdFloorPriceGTE != nil {
		opts = append(opts, collection.WithUsdFloorPriceGTE(*p.UsdFloorPriceGTE))
	}

	if p.UsdFloorPriceLTE != nil {
		opts = append(opts, collection.WithUsdFloorPriceLTE(*p.UsdFloorPriceLTE))
	}

	if p.YugaLab {
		opts = append(opts, collection.WithAddresses(domain.YugaLabCollectionAddresses))
	}

	var (
		res             []*collection.CollectionWithHoldingCount
		pagingRes       *collection.SearchResult
		err             error
		userCollections *account.AccountCollectionHoldings
	)

	if p.Holder != nil {
		userCollections, err = h.account.GetAccountCollectionHoldings(ctx, *p.Holder)
		if err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		}
		var addresses []domain.Address
		for cId := range userCollections.Collections {
			addresses = append(addresses, cId.Address)
		}
		opts = append(opts, collection.WithAddresses(addresses))
	}

	if p.LikedBy != nil {
		opts = append(opts, collection.WithLikedBy(*p.LikedBy))
	}

	if p.ListedBy != nil {
		opts = append(opts, collection.WithListedBy(*p.ListedBy))
	}

	if p.OfferedBy != nil {
		opts = append(opts, collection.WithOfferedBy(*p.OfferedBy))
	}

	if p.IncludeUnregistered != nil && *p.IncludeUnregistered {
		res, err = h.collection.FindAllIncludingUnregistered(ctx, opts...)
	} else {
		pagingRes, err = h.collection.FindAll(ctx, opts...)
		res = pagingRes.Items
	}

	if p.Holder != nil {
		for _, i := range res {
			cId := i.Collection.ToId()
			id := account.CollectionId{ChainId: cId.ChainId, Address: cId.Address}
			i.HoldingCount = userCollections.Collections[id]
			i.HoldingBalance = userCollections.CollectionsHoldingBalance[id]
		}
		collection.SortCollectionWithHoldingCount(res, p.SortBy)
	}

	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else if p.IsPaging != nil && *p.IsPaging {
		return delivery.MakeJsonResp(c, http.StatusOK, pagingRes)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) getMintable(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	type params struct {
		SortBy   *string         `query:"sortBy"`
		SortDir  *domain.SortDir `query:"sortDir"`
		ChainId  *domain.ChainId `query:"chainId"`
		Category *string         `query:"category"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}

	opts := []collection.FindAllOptions{
		//	@todo	pagination
		collection.WithPagination(0, 5000),
	}

	if p.SortBy != nil && p.SortDir != nil {
		opts = append(opts, collection.WithSort(*p.SortBy, *p.SortDir))
	}

	if p.ChainId != nil {
		opts = append(opts, collection.WithChainId(*p.ChainId))
	}

	if p.Category != nil {
		opts = append(opts, collection.WithCategory(*p.Category))
	}

	if res, err := h.collection.FindAllMintable(ctx, address, opts...); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) getUnreviewed(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	if res, err := h.collection.FindAllUnreviewd(ctx, address); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) collectionRequestCount() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			type params struct {
				ChainId  domain.ChainId `param:"chainId"`
				Contract domain.Address `param:"contract"`
			}

			p := params{}
			c.Bind(&p)
			met.BumpSum("get.count", 1, "chainId", fmt.Sprint(p.ChainId), "contract", p.Contract.ToLowerStr())
			return next(c)
		}
	}
}

// get
//
//	@Summary		Get collection info
//	@Description	Retrieve information for a collection
//	@Tags			collections
//	@Produce		json
//	@Param			chainId	path		int		true	"chain id"				example(1)
//	@Param			address	path		string	true	"collection address"	example(0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d)
//	@Success		200		{object}	collection.CollectionWithStat
//	@Failure		400
//	@Failure		500
//	@Router			/collection/{chainId}/{address} [get]
func (h *handler) get(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
	}

	p := params{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	collectionId := collection.CollectionId{ChainId: p.ChainId, Address: p.Contract}

	if res, err := h.collection.FindOneWithStat(ctx, collectionId); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) getViewCount(c echo.Context) error {
	id := collection.CollectionId{}

	if err := c.Bind(&id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	if res, err := h.collection.GetViewCount(ctx, id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) register(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	type payload struct {
		collection.Registration
		Signature string `json:"signature"`
	}

	p := payload{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid body")
	}

	if err := h.account.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	p.Owner = address

	if res, err := h.collection.Register(ctx, p.Registration); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, res)
	}
}

func (h *handler) review(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		Accept   bool           `json:"accept"`
		Reason   string         `json:"reason"`
	}

	p := params{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	collectionId := collection.CollectionId{ChainId: p.ChainId, Address: p.Contract}

	if p.Accept {
		res, err := h.collection.Accept(ctx, collectionId)

		if err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		} else {
			return delivery.MakeJsonResp(c, http.StatusCreated, res)
		}
	} else {
		res, err := h.collection.Reject(ctx, collectionId, p.Reason)

		if err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		} else {
			return delivery.MakeJsonResp(c, http.StatusCreated, res)
		}
	}
}

func (h *handler) ban(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	signer := c.Get("address").(domain.Address)

	type params struct {
		ChainId   domain.ChainId `param:"chainId"`
		Contract  domain.Address `param:"contract"`
		Signature string         `json:"signature"`
	}

	p := params{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.account.ValidateSignature(ctx, signer, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	collectionId := collection.CollectionId{ChainId: p.ChainId, Address: p.Contract}

	if res, err := h.collection.Ban(ctx, collectionId, true); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, res)
	}
}

func (h *handler) unban(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	signer := c.Get("address").(domain.Address)

	type params struct {
		ChainId   domain.ChainId `param:"chainId"`
		Contract  domain.Address `param:"contract"`
		Signature string         `json:"signature"`
	}

	p := params{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.account.ValidateSignature(ctx, signer, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	collectionId := collection.CollectionId{ChainId: p.ChainId, Address: p.Contract}
	if res, err := h.collection.Ban(ctx, collectionId, false); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, res)
	}
}

func (h *handler) getTopCollections(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	type params struct {
		PeriodType collection.PeriodType `query:"periodType"`
		Limit      int32                 `query:"limit"`
		Offset     int32                 `query:"offset"`
		YugaLab    bool                  `query:"yugaLab"`
	}
	p := params{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}
	if p.Limit == 0 {
		p.Limit = 15
	}

	opts := []domain.OpenseaDataFindAllOptions{
		domain.OpenseaDataWithPagination(p.Offset, p.Limit),
	}
	if p.YugaLab {
		opts = append(opts, domain.OpenseaDataWithAddresses(domain.YugaLabCollectionAddresses))
	}

	if res, err := h.collection.GetTopCollections(ctx, p.PeriodType, opts...); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		if p.YugaLab {
			m := map[domain.Address]collection.CollectionWithTradingVolume{}
			for _, col := range res {
				m[col.Erc721Address] = col
			}
			cols := []collection.CollectionWithTradingVolume{}
			for _, address := range domain.YugaLabCollectionAddresses {
				if _, ok := m[address]; ok {
					cols = append(cols, m[address])
				}
			}
			res = cols
		}

		return delivery.MakeJsonResp(c, http.StatusCreated, res)
	}
}

func (h *handler) getEditable(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	result, err := h.collection.FindAll(ctx, collection.WithAccountEditable(address))
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, result)
}

func (h *handler) updateInfo(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	type params struct {
		ChainId   domain.ChainId `param:"chainId"`
		Contract  domain.Address `param:"contract"`
		Signature string         `json:"signature"`
		collection.UpdateInfoPayload
	}

	p := params{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.account.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	id := collection.CollectionId{ChainId: p.ChainId, Address: p.Contract}

	collection, err := h.collection.FindOne(ctx, id)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	editableAccounts := append(collection.EditableAccounts, collection.Owner)
	accountMap := map[domain.Address]struct{}{}
	for _, ac := range editableAccounts {
		accountMap[ac.ToLower()] = struct{}{}
	}

	if _, ok := accountMap[address.ToLower()]; !ok {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, "you don't have permission to edit this collection")
	}

	if err := h.collection.UpdateInfo(ctx, id, p.UpdateInfoPayload); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusAccepted, "ok")
}

func (h *handler) likeCollection(c echo.Context) error {
	id := collection.CollectionId{}

	if err := c.Bind(&id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	liker := c.Get("address").(domain.Address)

	if res, err := h.like.Like(ctx, id, liker); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, res)
	}
}

func (h *handler) unlikeCollection(c echo.Context) error {
	id := collection.CollectionId{}

	if err := c.Bind(&id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	liker := c.Get("address").(domain.Address)

	if res, err := h.like.Unlike(ctx, id, liker); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, res)
	}

}

func (h *handler) getVolume(c echo.Context) error {
	p := struct {
		collection.CollectionId
		Period collection.PeriodType `query:"period"`
		Date   time.Time             `query:"date"` // utc 00:00
	}{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	id := collection.TradingVolumeId{
		ChainId: p.ChainId,
		Address: p.Address,
		Period:  p.Period,
		Date:    p.Date,
	}

	if res, err := h.tradingVolume.FindOne(ctx, id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) updateTraitFloor(c echo.Context) error {
	p := struct {
		collection.CollectionId
		TraitName  string  `json:"traitName" validate:"required"`
		TraitValue string  `json:"traitValue" validate:"required"`
		Price      float64 `json:"price" validate:"required"`
		Key        string  `json:"key"`
	}{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if p.Key != "7d4aIcbdJf" {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, "")
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	id := collection.CollectionId{
		ChainId: p.ChainId,
		Address: p.Address.ToLower(),
	}

	if err := h.collection.UpdateTraitFloorPrice(ctx, id, p.TraitName, p.TraitValue, p.Price); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, "")
}

// getActivities
//
//	@Summary		List collection activities
//	@Description	Retrieve a list of activities for a collection
//	@Tags			collections
//	@Produce		json
//	@Param			chainId	path		int			true	"chain id"															example(1)
//	@Param			address	path		string		true	"collection address"												example(0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d)
//	@Param			offset	query		int			false	"paging offset"														example(0)
//	@Param			limit	query		int			false	"paging size"														example(100)
//	@Param			types	query		[]string	false	"activity types (sold is only used in legacy on-chain marketplace)"	enums(sale, list, createOffer, sold, cancelListing, cancelOffer)	collectionFormat(multi)	example(sale)
//	@Success		200		{object}	collection.ActivityResult
//	@Failure		400
//	@Failure		500
//	@Router			/collection/{chainId}/{address}/activities [get]
func (h *handler) getActivities(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		collection.CollectionId
		Offset int                           `query:"offset"`
		Limit  int                           `query:"limit"`
		Types  []account.ActivityHistoryType `query:"types"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	opts := []account.FindActivityHistoryOptions{
		account.ActivityHistoryWithPagination(p.Offset, p.Limit),
		account.ActivityHistoryWithSource(account.SourceX),
	}

	if len(p.Types) > 0 {
		opts = append(opts, account.ActivityHistoryWithTypes(p.Types...))
	}

	id := collection.CollectionId{ChainId: p.ChainId, Address: p.Address}
	if res, err := h.collection.GetActivities(ctx, id, opts...); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

// getGlobalOfferStat
//
//	@Summary	Get global offer stat of a collection
//	@Tags		collections
//	@Produce	json
//	@Param		chainId	path		int		true	"chain id"				example(1)
//	@Param		address	path		string	true	"collection address"	example(0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d)
//	@Success	200		{object}	collection.GlobalOfferStatResult
//	@Failure	400
//	@Failure	500
//	@Router		/collection/{chainId}/{address}/globalofferstat [get]
func (h *handler) getGlobalOfferStat(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		collection.CollectionId
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	id := collection.CollectionId{ChainId: p.ChainId, Address: p.Address}
	if res, err := h.collection.GetGlobalOfferStats(ctx, id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}
