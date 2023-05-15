package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/like"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/domain/token"
	"github.com/x-xyz/goapi/service/hyype"
	authMiddleware "github.com/x-xyz/goapi/stores/auth/delivery/http/middleware"
)

type handler struct {
	token       token.Usecase
	like        like.Usecase
	account     account.Usecase
	folder      account.FolderUseCase
	order       order.UseCase
	hyypeClient hyype.Client
}

func New(
	e *echo.Echo,
	token token.Usecase,
	like like.Usecase,
	account account.Usecase,
	folder account.FolderUseCase,
	order order.UseCase,
	authMiddleware *authMiddleware.AuthMiddleware,
	hyypeClient hyype.Client,
) {
	h := &handler{token, like, account, folder, order, hyypeClient}

	gs := e.Group("/tokens", authMiddleware.OptionalAuth())

	gs.GET("", h.Search)

	gs.GET("/v2", h.SearchV2, authMiddleware.OptionalAuth())

	gs.POST("/upload", h.upload, authMiddleware.Auth())

	gs.POST("/mark-private", h.markTokensPrivate, authMiddleware.Auth())

	gs.GET("/order/:chainId/:orderHash", h.getOrder)

	// NOTE: not sure need to auth or not
	gs.POST("/make-order", h.makeOrder)

	g := e.Group("/token/:chainId/:contract/:tokenId")

	g.GET("", h.get, authMiddleware.OptionalAuth())

	g.GET("/activities", h.getActivities)

	g.GET("/price-histories/:period", h.getPriceHistories)

	g.GET("/likers", h.getLikers)

	g.GET("/view-count", h.getViewCount)

	g.GET("/lores", h.getLores)

	g.POST("/like", h.likeToken, authMiddleware.Auth())

	g.DELETE("/like", h.unlikeToken, authMiddleware.Auth())

	g.POST("/ban", h.ban, authMiddleware.Auth(), authMiddleware.IsModerator())

	g.POST("/unban", h.unban, authMiddleware.Auth(), authMiddleware.IsModerator())

	g.POST("/unlockable-content", h.addUnlockableContent, authMiddleware.Auth())

	g.POST("/unlockable-content/reveal", h.getUnlockableContent, authMiddleware.Auth())

	g.POST("/refresh-metadata", h.refreshMetadata)

	g.GET("/score", h.getOpenrarityScore)
}

func (h *handler) Search(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address")

	user := domain.Address("")

	if address != nil {
		user = address.(domain.Address)
	}

	type params struct {
		Offset      int32            `query:"offset"`
		Limit       int32            `query:"limit"`
		SortBy      *token.SortBy    `query:"sortBy"`
		SortDir     *domain.SortDir  `query:"sortDir"`
		Status      string           `query:"status"`
		Collections []domain.Address `query:"collections"`
		Category    *string          `query:"category"`
		ChainId     *domain.ChainId  `query:"chainId"`
		BelongsTo   *domain.Address  `query:"belongsTo"`
		LikedBy     *domain.Address  `query:"likedBy"`
		AttrFilters []string         `query:"attrFilters"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}

	opts := []token.SearchOptionsFunc{
		token.WithPagination(p.Offset, p.Limit),
	}

	if p.SortBy != nil && p.SortDir != nil {
		opts = append(opts, token.WithSort(string(*p.SortBy), *p.SortDir))
	}

	if len(p.Status) > 0 {
		for _, status := range strings.Split(p.Status, ",") {
			switch strings.ToLower(strings.TrimSpace(status)) {
			case "buynow":
				opts = append(opts, token.WithBuyNow())
			case "hasbids":
				opts = append(opts, token.WithHasBid())
			case "hasoffers":
				opts = append(opts, token.WithHasOffer())
			case "onauction":
				opts = append(opts, token.WithOnAuction())
			}
		}
	}

	if len(p.Collections) > 0 {
		opts = append(opts, token.WithCollections(p.Collections...))
	}

	if p.Category != nil {
		opts = append(opts, token.WithCategory(*p.Category))
	}

	if p.ChainId != nil {
		opts = append(opts, token.WithChainId(*p.ChainId))
	}

	if p.BelongsTo != nil {
		opts = append(opts, token.WithBelongsTo(*p.BelongsTo))
	}

	if p.LikedBy != nil {
		opts = append(opts, token.WithLikedBy(*p.LikedBy))
	}

	// usage
	// attrFilters={"name":"address","values":["0x35bcf180358e74d09dfe6c96f6ddc74262be506e","b"]}&attrFilters={"name":"recipient","values":["0x35bcf180358e74d09dfe6c96f6ddc74262be506e","d"]}
	if len(p.AttrFilters) > 0 {
		attrs := []nftitem.AttributeFilter{}
		for _, af := range p.AttrFilters {
			attr := nftitem.AttributeFilter{}
			if err := json.Unmarshal([]byte(af), &attr); err != nil {
				return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
			}
			attrs = append(attrs, attr)
		}
		opts = append(opts, token.WithAttributes(attrs))
	}

	res, err := h.token.Search(ctx, opts...)

	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if !user.IsEmpty() {
		for _, item := range res.Items {
			if isLiked, err := h.like.IsLiked(ctx, item.ChainId, item.ContractAddress, item.TokenId, user); err != nil {
				ctx.WithField("err", err).WithField("nftitemId", item.NftItem.ToId()).Error("like.IsLiked")
				continue
			} else {
				item.IsLiked = &isLiked
			}
		}
	}

	return delivery.MakeJsonResp(c, http.StatusOK, res)
}

// SearchV2 godoc
//
//	@Description	This api returns a list of NFTs based on the specified query parameters.
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Param			offset			query		int			false	"paging offset"
//	@Param			limit			query		int			false	"paging size"						example(100)
//	@Param			sortBy			query		string		false	"NFT sorting rule"					Enums(price_low_to_high, price_high_to_low, offer_price_low_to_high, offer_price_high_to_low)	example(price_low_to_high)
//	@Param			saleStatus		query		[]string	false	"Filter with specific order type"	enums(buynow, hasoffer)																			example(buynow)	collectionFormat(multi)
//	@Param			chainId			query		int			false	"chain id. e.g: `1` for ethereum"	example(1)
//	@Param			collections		query		string		false	"NFT collection contract address"	example(0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d)
//	@Param			includeOrders	query		bool		false	"determining if order information should be included in the response."
//	@Param			belongsTo		query		string		false	"NFT belongs to owner address"	example(0xed2ab4948bA6A909a7751DEc4F34f303eB8c7236)
//	@Param			offerOwners		query		string		false	"Get NFT with offer owner"		example(0x020ca66c30bec2c4fe3861a94e4db4a498a35872)
//	@Success		200				{object}	token.SearchResult
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/tokens/v2 [get]
func (h *handler) SearchV2(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address")

	user := domain.Address("")

	if address != nil {
		user = address.(domain.Address)
	}

	p := &token.SearchParams{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid params")
	}

	sortBy, sortDir, requiredOpts := token.ParseSearchSortOption(p.SortBy)

	saleSatatus := nftitem.ParseSaleStatusType(p.SaleStatus...)

	// set limit to 1000 if limit == 0 or larger than 1000
	if p.Limit == 0 || p.Limit > 1000 {
		p.Limit = 1000
	}

	opts := []token.SearchOptionsFunc{
		token.WithPagination(p.Offset, p.Limit),
		token.WithSort(sortBy, sortDir),
		token.WithSaleStatus(saleSatatus),
	}

	opts = append(opts, requiredOpts...)

	if p.Cursor != nil {
		opts = append(opts, token.WithCursor(*p.Cursor))
	}

	if p.Size != nil {
		opts = append(opts, token.WithSize(*p.Size))
	}

	if p.PriceGTE != nil || p.PriceLTE != nil || p.PriceInUsdGTE != nil || p.PriceInUsdLTE != nil {
		opts = append(opts, token.WithBuyNow())
	}

	if p.ChainId != nil {
		opts = append(opts, token.WithChainId(*p.ChainId))

		if len(p.Collections) > 0 {
			opts = append(opts, token.WithCollections(p.Collections...))
		}
	}

	if p.Category != nil {
		opts = append(opts, token.WithCategory(*p.Category))
	}

	if p.BelongsTo != nil {
		opts = append(opts, token.WithBelongsTo(*p.BelongsTo))
	}

	if p.NotBelongsTo != nil {
		opts = append(opts, token.WithNotBelongsTo(*p.NotBelongsTo))
	}

	if p.ListingFrom != nil {
		opts = append(opts, token.WithListingFrom(*p.ListingFrom))
	}

	if p.InactiveListingFrom != nil {
		opts = append(opts, token.WithInactiveListingFrom(*p.InactiveListingFrom))
	}

	if p.LikedBy != nil {
		opts = append(opts, token.WithLikedBy(*p.LikedBy))
	}

	if p.PriceGTE != nil {
		opts = append(opts, token.WithPriceGTE(*p.PriceGTE))
	}

	if p.PriceLTE != nil {
		opts = append(opts, token.WithPriceLTE(*p.PriceLTE))
	}

	if p.PriceInUsdGTE != nil {
		opts = append(opts, token.WithPriceInUsdGTE(*p.PriceInUsdGTE))
	}

	if p.PriceInUsdLTE != nil {
		opts = append(opts, token.WithPriceInUsdLTE(*p.PriceInUsdLTE))
	}

	if p.OfferPriceInUsdGTE != nil {
		opts = append(opts, token.WithOfferPriceInUsdGTE(*p.OfferPriceInUsdGTE))
	}

	if p.OfferPriceInUsdLTE != nil {
		opts = append(opts, token.WithOfferPriceInUsdLTE(*p.OfferPriceInUsdLTE))
	}

	if p.Name != nil {
		opts = append(opts, token.WithName(*p.Name))
	}

	if p.Search != nil {
		opts = append(opts, token.WithSearch(*p.Search))
	}

	if p.TokenType != nil {
		opts = append(opts, token.WithTokenType(*p.TokenType))
	}

	if p.BidOwner != nil {
		opts = append(opts, token.WithBidOwner(p.BidOwner.ToLower()))
	}

	if p.IncludeOrders != nil {
		opts = append(opts, token.WithIncludeOrders(*p.IncludeOrders))
	}

	if p.IncludeInactiveOrders != nil {
		opts = append(opts, token.WithIncludeInactiveOrders(*p.IncludeInactiveOrders))
	}

	if len(p.OfferOwners) > 0 {
		opts = append(opts, token.WithOfferOwners(p.OfferOwners...))
	}

	// usage
	// attrFilters={"name":"address","values":["0x35bcf180358e74d09dfe6c96f6ddc74262be506e","b"]}&attrFilters={"name":"recipient","values":["0x35bcf180358e74d09dfe6c96f6ddc74262be506e","d"]}
	if len(p.AttrFilters) > 0 {
		attrs := []nftitem.AttributeFilter{}
		for _, af := range p.AttrFilters {
			attr := nftitem.AttributeFilter{}
			if err := json.Unmarshal([]byte(af), &attr); err != nil {
				return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
			}
			attrs = append(attrs, attr)
		}
		opts = append(opts, token.WithAttributes(attrs))
	}

	if p.FolderId != nil {
		// get folder to check privacy
		f, err := h.folder.GetFolder(ctx, *p.FolderId)
		if err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		}

		// if folder is private and user not owner, return StatusMethodNotAllowed
		if f.IsPrivate {
			ad := c.Get("address")
			if ad == nil || !ad.(domain.Address).Equals(f.Owner) {
				return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
			}
		}

		opts = append(opts, token.WithFolderId(*p.FolderId))
	}

	res, err := h.token.SearchV2(ctx, opts...)

	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if !user.IsEmpty() {
		for _, item := range res.Items {
			if isLiked, err := h.like.IsLiked(ctx, item.ChainId, item.ContractAddress, item.TokenId, user); err != nil {
				ctx.WithField("err", err).WithField("nftitemId", item.NftItem.ToId()).Error("like.IsLiked")
			} else {
				item.IsLiked = &isLiked
			}
		}
	}

	return delivery.MakeJsonResp(c, http.StatusOK, res)
}

// GetNFT godoc
//
//	@Description	This api returns NFT information based on `chain id`, `contract address` and `token id`
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Param			chainId		path		int		true	"chain id"			example(1)
//	@Param			contract	path		string	true	"contract address"	example(0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d)
//	@Param			tokenId		path		string	true	"token id"			example(6969)
//	@Success		200			{object}	token.TokenWithDetail
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/token/{chainId}/{contract}/{tokenId} [get]
func (h *handler) get(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address")

	user := domain.Address("")

	if address != nil {
		user = address.(domain.Address)
	}

	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		TokenId  domain.TokenId `param:"tokenId"`
	}

	p := params{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	id := nftitem.Id{ChainId: p.ChainId, ContractAddress: p.Contract, TokenId: p.TokenId}

	res, err := h.token.FindOne(ctx, id)

	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if _, err := h.token.GetUnlockableContent(ctx, id); err == domain.ErrNotFound {
		res.HasUnlockable = ptr.Bool(false)
	} else {
		res.HasUnlockable = ptr.Bool(true)
	}

	if !user.IsEmpty() {
		if isLiked, err := h.like.IsLiked(ctx, p.ChainId, p.Contract, p.TokenId, user); err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		} else {
			res.IsLiked = &isLiked
		}
	}

	return delivery.MakeJsonResp(c, http.StatusOK, res)
}

// getActivities
//
//	@Summary		List token activities
//	@Description	Retrieve a list of activities for a token
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Param			chainId		path		int		true	"chain id"			example(1)
//	@Param			contract	path		string	true	"contract address"	example(0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d)
//	@Param			tokenId		path		string	true	"token id"			example(6969)
//	@Param			offset		query		int		false	"paging offset"
//	@Param			limit		query		int		false	"paging size"	example(100)
//	@Success		200			{object}	token.ActivityResult
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/token/{chainId}/{contract}/{tokenId}/activities [get]
func (h *handler) getActivities(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		TokenId  domain.TokenId `param:"tokenId"`
		Offset   int            `query:"offset"`
		Limit    int            `query:"limit"`
	}

	p := params{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	id := nftitem.Id{ChainId: p.ChainId, ContractAddress: p.Contract, TokenId: p.TokenId}

	if p.Limit == 0 {
		p.Limit = 5
	}

	if res, err := h.token.GetActivities(ctx, id, p.Offset, p.Limit); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) getPriceHistories(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		ChainId  domain.ChainId    `param:"chainId"`
		Contract domain.Address    `param:"contract"`
		TokenId  domain.TokenId    `param:"tokenId"`
		Period   domain.TimePeriod `param:"period"`
	}

	p := params{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if !p.Period.IsValid() {
		return delivery.MakeJsonResp(c, http.StatusNotFound, nil)
	}

	id := nftitem.Id{ChainId: p.ChainId, ContractAddress: p.Contract, TokenId: p.TokenId}

	if res, err := h.token.GetPriceHistories(ctx, id, p.Period); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}

}

func (h *handler) getLikers(c echo.Context) error {
	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		TokenId  domain.TokenId `param:"tokenId"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	if res, err := h.like.GetLikers(ctx, p.ChainId, p.Contract, p.TokenId); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) getViewCount(c echo.Context) error {
	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		TokenId  domain.TokenId `param:"tokenId"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	id := nftitem.Id{ChainId: p.ChainId, ContractAddress: p.Contract, TokenId: p.TokenId}

	if res, err := h.token.GetViewCount(ctx, id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) getLores(c echo.Context) error {
	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		TokenId  domain.TokenId `param:"tokenId"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)
	if p.ChainId != 1 {
		return c.JSONBlob(http.StatusOK, []byte("[]"))
	}
	if res, err := h.hyypeClient.GetLoresOfNft(ctx, p.Contract.ToLower().ToLowerStr(), p.TokenId.String(), 0, 10); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return c.JSONBlob(http.StatusOK, res)
	}

}

func (h *handler) likeToken(c echo.Context) error {
	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		TokenId  domain.TokenId `param:"tokenId"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	liker := c.Get("address").(domain.Address)

	if res, err := h.like.Like(ctx, p.ChainId, p.Contract, p.TokenId, liker); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, res)
	}
}

func (h *handler) unlikeToken(c echo.Context) error {
	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		TokenId  domain.TokenId `param:"tokenId"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	liker := c.Get("address").(domain.Address)

	if res, err := h.like.Unlike(ctx, p.ChainId, p.Contract, p.TokenId, liker); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, res)
	}
}

func (h *handler) ban(c echo.Context) error {

	type params struct {
		ChainId   domain.ChainId `param:"chainId"`
		Contract  domain.Address `param:"contract"`
		TokenId   domain.TokenId `param:"tokenId"`
		Signature string         `json:"signature"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	if err := h.account.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	id := nftitem.Id{ChainId: p.ChainId, ContractAddress: p.Contract, TokenId: p.TokenId}

	if err := h.token.BanNftItem(ctx, id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusAccepted, nil)
	}
}

func (h *handler) unban(c echo.Context) error {
	type params struct {
		ChainId   domain.ChainId `param:"chainId"`
		Contract  domain.Address `param:"contract"`
		TokenId   domain.TokenId `param:"tokenId"`
		Signature string         `json:"signature"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	if err := h.account.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	id := nftitem.Id{ChainId: p.ChainId, ContractAddress: p.Contract, TokenId: p.TokenId}

	if err := h.token.UnbanNftItem(ctx, id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusAccepted, nil)
	}
}

func (h *handler) upload(c echo.Context) error {
	type params struct {
		token.UploadPayload
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	if res, err := h.token.Upload(ctx, address, p.UploadPayload); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, res)
	}
}

func (h *handler) getUnlockableContent(c echo.Context) error {
	type params struct {
		ChainId   domain.ChainId `param:"chainId"`
		Contract  domain.Address `param:"contract"`
		TokenId   domain.TokenId `param:"tokenId"`
		Signature string         `json:"signature"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	if err := h.account.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	id := nftitem.Id{ChainId: p.ChainId, ContractAddress: p.Contract, TokenId: p.TokenId}

	if res, err := h.token.GetUnlockableContent(ctx, id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) addUnlockableContent(c echo.Context) error {
	type params struct {
		ChainId   domain.ChainId `param:"chainId"`
		Contract  domain.Address `param:"contract"`
		TokenId   domain.TokenId `param:"tokenId"`
		Signature string         `json:"signature"`
		Content   string         `json:"content"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	if err := h.account.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	id := nftitem.Id{ChainId: p.ChainId, ContractAddress: p.Contract, TokenId: p.TokenId}

	if err := h.token.AddUnlockableContent(ctx, id, p.Content); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusCreated, nil)
	}
}

func (h *handler) refreshMetadata(c echo.Context) error {
	type params struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		TokenId  domain.TokenId `param:"tokenId"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	ctx := c.Get("ctx").(ctx.Ctx)
	id := nftitem.Id{ChainId: p.ChainId, ContractAddress: p.Contract, TokenId: p.TokenId}

	if err := h.token.RefreshIndexerState(ctx, id); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, 1)
	}
}

func (h *handler) markTokensPrivate(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type payload struct {
		Marks     []nftitem.Id `json:"marks"`
		Unmarks   []nftitem.Id `json:"unmarks"`
		Signature string       `json:"signature"`
	}

	p := payload{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	address := c.Get("address").(domain.Address)

	if err := h.account.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	if err := h.folder.MarkNftPrivate(ctx, address, p.Marks, p.Unmarks); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, 1)
}

//	@Summary		Create new order
//	@Description	Send a signed order to create new maker orders for listings/offers
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Param			order	body	order.Order	true	"signed order"
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/tokens/make-order [post]
func (h *handler) makeOrder(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	payload := struct {
		Order order.Order `json:"order"`
	}{}

	if err := c.Bind(&payload); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	payload.Order.FeeDistType = order.ToFeeDistType(payload.Order.FeeDistType)

	if err := h.order.MakeOrder(ctx, payload.Order); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, 1)
}

// getOrder godoc
//
//	@Description	Get order information by order hash
//	@Description	Order can then be send to contract to fulfill purchase
//	@Description	prod contract: https://etherscan.io/address/0xb4a2E49818dd8a5CdD818f22aB99263b62DDEB6c
//	@Description	dev contract: https://goerli.etherscan.io/address/0x33962e44cd0a6fa8dbca34c62ffcac418cb079f3
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Param			chainId		path		int		true	"chain id. e.g: `1` for ethereum"					example(1)
//	@Param			orderHash	path		string	true	"order hash value, can be aquired by `/tokens/v2`"	example(0x633edcf47aa679dda02cd5a4494fd5844c6ec6264449896e0264984b367ed25b)
//	@Success		200			{object}	order.Order
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/tokens/order/{chainId}/{orderHash} [get]
func (h *handler) getOrder(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	payload := struct {
		ChainId   domain.ChainId `param:"chainId"`
		OrderHash string         `param:"orderHash"`
	}{}

	if err := c.Bind(&payload); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	id := order.OrderId{ChainId: payload.ChainId, OrderHash: domain.OrderHash(payload.OrderHash)}
	order, err := h.order.GetOrder(ctx, id)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, order)
}

func (h *handler) getOpenrarityScore(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	payload := struct {
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
		TokenId  domain.TokenId `param:"tokenId"`
	}{}

	if err := c.Bind(&payload); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	score, err := h.token.GetOpenRararityScore(ctx, nftitem.Id{
		ChainId:         payload.ChainId,
		ContractAddress: payload.Contract.ToLower(),
		TokenId:         payload.TokenId,
	})

	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, score)
}
