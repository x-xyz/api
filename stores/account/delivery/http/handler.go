package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/like"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/middleware"
	authMiddleware "github.com/x-xyz/goapi/stores/auth/delivery/http/middleware"
)

type handler struct {
	au         account.Usecase
	like       like.Usecase
	fu         account.FolderUseCase
	collection collection.Usecase
	orderNonce account.OrderNonceUseCase
}

// New will initialize the healthcheck/
func New(e *echo.Echo, au account.Usecase, like like.Usecase, fu account.FolderUseCase, collection collection.Usecase, authMiddleware *authMiddleware.AuthMiddleware, orderNonce account.OrderNonceUseCase) {
	h := &handler{
		au:         au,
		like:       like,
		fu:         fu,
		collection: collection,
		orderNonce: orderNonce,
	}
	g := e.Group("/account")
	g.GET("/:account", h.getAccount, middleware.IsValidAddress("account"))
	g.GET("/:account/liked", h.getLikedTokens, middleware.IsValidAddress("account"))
	// TODO: add pagination for followers & followings
	g.GET("/:account/followers", h.getFollowers, middleware.IsValidAddress("account"))
	g.GET("/:account/followings", h.getFollowings, middleware.IsValidAddress("account"))
	g.GET("/:account/follow", h.isFollowing, middleware.IsValidAddress("account"), authMiddleware.Auth())
	g.POST("/:account/follow", h.follow, middleware.IsValidAddress("account"), authMiddleware.Auth())
	g.DELETE("/:account/follow", h.unfollow, middleware.IsValidAddress("account"), authMiddleware.Auth())
	g.GET("/:account/activities", h.getActivities, middleware.IsValidAddress("account"))
	g.GET("/:account/stat", h.getStat, middleware.IsValidAddress("account"))
	g.GET("/:account/folders", h.getFolders, middleware.IsValidAddress("account"), authMiddleware.OptionalAuth())
	g.POST("/:account/folders", h.createFolder, middleware.IsValidAddress("account"), authMiddleware.Auth())
	g.GET("/:account/folder/:folderId", h.getFolder, middleware.IsValidAddress("account"), authMiddleware.OptionalAuth())
	g.GET("/:account/folder/:folderId/nfts", h.getNFTsInFolder, middleware.IsValidAddress("account"), authMiddleware.OptionalAuth())
	g.PUT("/:account/folder/:folderId", h.updateFolder, middleware.IsValidAddress("account"), authMiddleware.Auth())
	g.DELETE("/:account/folder/:folderId", h.deleteFolder, middleware.IsValidAddress("account"), authMiddleware.Auth())
	g.GET("/:account/collection/:chainId/:contract", h.getCollectionStatByAccount)
	g.GET("/:account/collection-summary", h.getCollectionSummary)
	g.GET("/:account/orderNonce/:chainId", h.useOrderNonce, authMiddleware.Auth())

	// self
	g.PATCH("", h.updateAccount, authMiddleware.Auth())
	g.PATCH("/banner", h.updateBanner, authMiddleware.Auth())
	g.PATCH("/avatar", h.updateAvatar, authMiddleware.Auth())
	g.POST("/nonce", h.generateNonce, authMiddleware.Auth())
	g.GET("/settings/notification", h.getNotifSettings, authMiddleware.Auth())
	g.PUT("/settings/notification", h.updateNotifSettings, authMiddleware.Auth())

	// admin
	g.POST("/ban", h.ban, authMiddleware.Auth(), authMiddleware.IsModerator())
	g.POST("/unban", h.unban, authMiddleware.Auth(), authMiddleware.IsModerator())

}

func (h *handler) getAccount(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	pAaccount := domain.Address(c.Param("account"))
	info, err := h.au.Get(ctx, pAaccount)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}
	return delivery.MakeJsonResp(c, http.StatusOK, info)
}

func (h *handler) updateAccount(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	type payload struct {
		account.Updater
		ImgData       string `json:"imgData"`
		AvatarImgData string `json:"avatarImgData"`
		BannerImgData string `json:"bannerImgData"`
		Signature     string `json:"signature"`
	}

	p := &payload{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.au.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	// for backward compatibility
	if len(p.AvatarImgData) > 0 {
		if _, err := h.au.UpdateAvatar(ctx, address, p.AvatarImgData); err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		}
	} else if len(p.ImgData) > 0 {
		if _, err := h.au.UpdateAvatar(ctx, address, p.ImgData); err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		}
	}

	if len(p.BannerImgData) > 0 {
		if _, err := h.au.UpdateBanner(ctx, address, p.BannerImgData); err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		}
	}

	if info, err := h.au.Update(ctx, address, &p.Updater); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, info)
	}
}

func (h *handler) updateBanner(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	address := c.Get("address").(domain.Address)

	type formParams struct {
		ImgData   string `form:"imgData"`
		Signature string `form:"signature"`
	}

	p := formParams{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.au.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	if cid, err := h.au.UpdateBanner(ctx, address, p.ImgData); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, cid)
	}
}

func (h *handler) updateAvatar(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	address := c.Get("address").(domain.Address)

	type formParams struct {
		ImgData   string `form:"imgData"`
		Signature string `form:"signature"`
	}

	p := formParams{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.au.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	if cid, err := h.au.UpdateAvatar(ctx, address, p.ImgData); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, cid)
	}
}

// generateNonce
//
//	@Summary		Generate nonce for signing
//	@Description	Generate nonce for signing
//	@Tags			account
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{integer}	integer	"nonce"
//	@Failure		500
//	@Router			/account/nonce [post]
func (h *handler) generateNonce(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	address := c.Get("address").(domain.Address)
	nonce, err := h.au.GenerateNonce(ctx, address)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, nonce)
}

func (h *handler) getNotifSettings(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	address := c.Get("address").(domain.Address)
	settings, err := h.au.GetNotificationSettings(ctx, address)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}
	return delivery.MakeJsonResp(c, http.StatusOK, settings)
}

func (h *handler) updateNotifSettings(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	address := c.Get("address").(domain.Address)

	type payload struct {
		account.NotificationSettings
		Signature string `json:"signature"`
	}

	p := &payload{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	p.NotificationSettings.Address = address

	settings, err := h.au.UpsertNotificationSettings(ctx, &p.NotificationSettings)

	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, settings)
}

func (h *handler) getLikedTokens(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	liker := domain.Address(c.Param("address"))

	if res, err := h.like.GetLikeds(ctx, liker); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) follow(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)
	toAddress := domain.Address(c.Param("account"))

	if err := h.au.Follow(ctx, address, toAddress); err != nil {
		ctx.WithField("err", err).Error("follow.Follow failed")
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusNoContent, nil)
	}
}

func (h *handler) unfollow(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)
	toAddress := domain.Address(c.Param("account"))

	if err := h.au.Unfollow(ctx, address, toAddress); err != nil {
		ctx.WithField("err", err).Error("follow.Unfollow failed")
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusNoContent, nil)
	}
}

func (h *handler) isFollowing(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	address := c.Get("address").(domain.Address)
	toAddress := domain.Address(c.Param("account"))

	if isFollowing, err := h.au.IsFollowing(ctx, address, toAddress); err != nil {
		ctx.WithField("err", err).Error("follow.GetFollowers failed")
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, isFollowing)
	}
}

func (h *handler) getFollowers(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	address := domain.Address(c.Param("account"))

	if res, err := h.au.GetFollowers(ctx, address); err != nil {
		ctx.WithField("err", err).Error("follow.GetFollowers failed")
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) getFollowings(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	address := domain.Address(c.Param("account"))

	if res, err := h.au.GetFollowings(ctx, address); err != nil {
		ctx.WithField("err", err).Error("follow.GetFollowers failed")
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) ban(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	type payload struct {
		Addresses []domain.Address `json:"addresses"`
		Signature string           `json:"signature"`
	}

	p := &payload{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.au.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	type response struct {
		Successes []domain.Address `json:"successes"`
		Fails     []domain.Address `json:"fails"`
	}

	resp := &response{}

	for _, address := range p.Addresses {
		if err := h.au.Ban(ctx, address); err != nil {
			resp.Fails = append(resp.Fails, address)
		} else {
			resp.Successes = append(resp.Successes, address)
		}
	}

	return delivery.MakeJsonResp(c, http.StatusAccepted, resp)
}

func (h *handler) unban(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	address := c.Get("address").(domain.Address)

	type payload struct {
		Addresses []domain.Address `json:"addresses"`
		Signature string           `json:"signature"`
	}

	p := &payload{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.au.ValidateSignature(ctx, address, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	type response struct {
		Successes []domain.Address `json:"successes"`
		Fails     []domain.Address `json:"fails"`
	}

	resp := &response{}

	for _, address := range p.Addresses {
		if err := h.au.Unban(ctx, address); err != nil {
			resp.Fails = append(resp.Fails, address)
		} else {
			resp.Successes = append(resp.Successes, address)
		}
	}

	return delivery.MakeJsonResp(c, http.StatusAccepted, resp)
}

// GetActivities
//
//	@Description	This api returns a list of activities based on the specified query parameters.
//	@Tags			activities
//	@Accept			json
//	@Produce		json
//	@Param			account	path		string	true	"account address"	example(0x020ca66c30bec2c4fe3861a94e4db4a498a35872)
//	@Param			limit	query		int		false	"paging size"
//	@Param			offset	query		int		false	"paging offset"
//	@Success		200		{object}	account.ActivityResult
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/account/{account}/activities [get]
func (h *handler) getActivities(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		Address  domain.Address                `param:"account"`
		Offset   int                           `query:"offset"`
		Limit    int                           `query:"limit"`
		ChainId  *domain.ChainId               `query:"chainId"`
		Contract *domain.Address               `query:"contract"`
		TokenId  *domain.TokenId               `query:"tokenId"`
		Types    []account.ActivityHistoryType `query:"types"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	opts := []account.FindActivityHistoryOptions{
		account.ActivityHistoryWithPagination(p.Offset, p.Limit),
		account.ActivityHistoryWithSource(account.SourceX),
	}

	if p.ChainId != nil && p.Contract != nil && p.TokenId != nil {
		opts = append(opts, account.ActivityHistoryWithToken(*p.ChainId, *p.Contract, *p.TokenId))
	} else if p.ChainId != nil && p.Contract != nil {
		opts = append(opts, account.ActivityHistoryWithCollection(*p.ChainId, *p.Contract))
	} else if p.ChainId != nil {
		opts = append(opts, account.ActivityHistoryWithChainId(*p.ChainId))
	}

	if len(p.Types) > 0 {
		opts = append(opts, account.ActivityHistoryWithTypes(p.Types...))
	}

	if res, err := h.au.GetActivities(ctx, p.Address, opts...); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) getStat(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		Address domain.Address `param:"account"`
	}

	p := &params{}

	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if res, err := h.au.GetAccountStat(ctx, p.Address); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	} else {
		return delivery.MakeJsonResp(c, http.StatusOK, res)
	}
}

func (h *handler) getFolders(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		Account domain.Address `param:"account"`
	}

	p := &params{}
	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	folders, err := h.fu.GetFolders(ctx, account.WithOwner(p.Account.ToLower()))
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if len(folders) == 0 {
		// init folders if not getting any folder
		err := h.fu.InitBuiltInFolders(ctx, p.Account.ToLower())
		if err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		}

		folders, err = h.fu.GetFolders(ctx, account.WithOwner(p.Account.ToLower()))
		if err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		}
	}

	isAuthed := false
	ad := c.Get("address")
	if ad != nil && ad.(domain.Address).Equals(p.Account) {
		isAuthed = true
	}

	res := []*account.Folder{}
	if isAuthed {
		res = folders
	} else {
		for _, f := range folders {
			if !f.IsPrivate {
				res = append(res, f)
			}
		}
	}

	return delivery.MakeJsonResp(c, http.StatusOK, res)
}

func (h *handler) getFolder(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		FolderId string         `param:"folderId"`
		Account  domain.Address `param:"account"`
	}

	p := params{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	folder, err := h.fu.GetFolder(ctx, p.FolderId)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if !folder.Owner.Equals(p.Account) {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "folder not belongs to this account")
	}

	if folder.IsPrivate {
		ad := c.Get("address")
		if ad == nil || !ad.(domain.Address).Equals(p.Account) {
			return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
		}
	}

	return delivery.MakeJsonResp(c, http.StatusOK, folder)
}

func (h *handler) getNFTsInFolder(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		FolderId string         `param:"folderId"`
		Account  domain.Address `param:"account"`
	}

	p := params{}
	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	folder, err := h.fu.GetFolder(ctx, p.FolderId)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if !folder.Owner.Equals(p.Account) {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "folder not belongs to this account")
	}

	res, err := h.fu.GetNFTsInFolder(ctx, p.FolderId)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if folder.IsPrivate {
		ad := c.Get("address")
		if ad == nil || !ad.(domain.Address).Equals(p.Account) {
			return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
		}
	}

	return delivery.MakeJsonResp(c, http.StatusOK, res)
}

func (h *handler) createFolder(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type payload struct {
		Account   domain.Address `param:"account"`
		Signature string         `json:"signature"`
		Name      string         `json:"name" bson:"name"`
		IsPrivate bool           `json:"isPrivate" bson:"isPrivate"`
		Nfts      []nftitem.Id   `json:"nfts"`
	}

	p := payload{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	if err := h.au.ValidateSignature(ctx, p.Account, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	folder := account.Folder{
		Name:      p.Name,
		IsPrivate: p.IsPrivate,
		Owner:     p.Account,
	}

	id, err := h.fu.Create(ctx, &folder)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if len(p.Nfts) > 0 {
		patchable := &account.FolderUpdater{}
		if err := h.fu.Update(ctx, id, patchable, p.Nfts); err != nil {
			return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
		}
	}

	type response struct {
		FolderId string `json:"folderId"`
	}

	res := response{
		FolderId: id,
	}

	return delivery.MakeJsonResp(c, http.StatusOK, res)
}

func (h *handler) updateFolder(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type payload struct {
		FolderId  string         `param:"folderId"`
		Account   domain.Address `param:"account"`
		Signature string         `json:"signature"`
		Name      string         `json:"name" bson:"name"`
		IsPrivate bool           `json:"isPrivate" bson:"isPrivate"`
		Nfts      []nftitem.Id   `json:"nfts"`
	}

	p := payload{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	folder, err := h.fu.GetFolder(ctx, p.FolderId)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if folder.IsBuiltIn {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, "cannot edit built-in folder")
	}

	if !folder.Owner.Equals(p.Account) {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "folder not belongs to this account")
	}

	if err := h.au.ValidateSignature(ctx, folder.Owner, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	patchable := account.FolderUpdater{
		Name:      &p.Name,
		IsPrivate: &p.IsPrivate,
	}

	if err := h.fu.Update(ctx, p.FolderId, &patchable, p.Nfts); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, "")
}

func (h *handler) deleteFolder(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type payload struct {
		FolderId  string         `param:"folderId"`
		Account   domain.Address `param:"account"`
		Signature string         `json:"signature"`
	}

	p := payload{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	folder, err := h.fu.GetFolder(ctx, p.FolderId)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	if !folder.Owner.Equals(p.Account) {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, "folder not belongs to this account")
	}

	if err := h.au.ValidateSignature(ctx, folder.Owner, p.Signature); err != nil {
		return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, err)
	}

	if err := h.fu.Delete(ctx, p.FolderId); err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, "")
}

func (h *handler) getCollectionStatByAccount(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)
	p := struct {
		Account  domain.Address `param:"account"`
		ChainId  domain.ChainId `param:"chainId"`
		Contract domain.Address `param:"contract"`
	}{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	collectionId := collection.CollectionId{
		ChainId: p.ChainId,
		Address: p.Contract.ToLower(),
	}

	stat, err := h.collection.GetCollectionStatByAccount(ctx, collectionId, p.Account)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	return delivery.MakeJsonResp(c, http.StatusOK, stat)
}

func (h *handler) getCollectionSummary(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	p := struct {
		Account domain.Address `param:"account"`
	}{}

	if err := c.Bind(&p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	folders, err := h.fu.GetFolders(
		ctx,
		account.WithOwner(p.Account.ToLower()),
		account.WithBuiltIn(true),
	)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	userCollections, err := h.au.GetAccountCollectionHoldings(ctx, p.Account.ToLower())
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	var addresses []domain.Address
	for cId := range userCollections.Collections {
		addresses = append(addresses, cId.Address)
	}

	collectionsRes, err := h.collection.FindAll(ctx, collection.WithAddresses(addresses), collection.WithPagination(0, 1))
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}

	totalValue := float64(0)
	nftCount := 0
	previousValue := float64(0)
	instantLiquidityValue := float64(0)
	for _, f := range folders {
		totalValue += f.TotalValueInUsd
		nftCount += f.NftCount
		previousValue += f.TotalValueInUsd / (1 + f.TotalValueMovement)
		instantLiquidityValue += f.InstantLiquidityInUsd
	}

	totalValueChange := float64(0)
	if previousValue > 0 {
		totalValueChange = (totalValue - previousValue) / previousValue
	}

	instantLiquidityRatio := float64(0)
	if totalValue > 0 {
		instantLiquidityRatio = instantLiquidityValue / totalValue
	}

	res := struct {
		TotalCollectionValue       float64 `json:"totalCollectionValue"`
		NftCount                   int     `json:"nftCount"`
		CollectionCount            int     `json:"collectionCount"`
		TotalCollectionValueChange float64 `json:"totalCollectionValueChange"`
		InstantLiquidityValue      float64 `json:"instantLiquidityValue"`
		InstantLiquidityRatio      float64 `json:"instantLiquidityRatio"`
	}{
		TotalCollectionValue:       totalValue,
		NftCount:                   nftCount,
		CollectionCount:            collectionsRes.Count,
		TotalCollectionValueChange: totalValueChange,
		InstantLiquidityValue:      instantLiquidityValue,
		InstantLiquidityRatio:      instantLiquidityRatio,
	}

	return delivery.MakeJsonResp(c, http.StatusOK, res)
}

// userOrderNonce
//
//	@Summary		Get next valid nonce for account
//	@Description	Order nonce is used in #/tokens/post_tokens_make_order
//	@Tags			account
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			account	path		string					true	"account address"					example(0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d)
//	@Param			chainId	path		int						true	"chain id. e.g: `1` for ethereum"	example(1)
//	@Success		200		{object}	object{nonce=string}	"nonce"
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/account/{account}/orderNonce/{chainId} [get]
func (h *handler) useOrderNonce(c echo.Context) error {
	ctx := c.Get("ctx").(ctx.Ctx)

	type params struct {
		Account domain.Address `param:"account"`
		ChainId domain.ChainId `param:"chainId"`
	}
	p := &params{}
	if err := c.Bind(p); err != nil {
		return delivery.MakeJsonResp(c, http.StatusBadRequest, err)
	}

	id := account.OrderNonceId{Address: p.Account.ToLower(), ChainId: p.ChainId}
	nonce, err := h.orderNonce.UseAvailableNonce(ctx, id)
	if err != nil {
		return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
	}
	res := struct {
		Nonce string `json:"nonce"`
	}{
		Nonce: nonce,
	}
	return delivery.MakeJsonResp(c, http.StatusOK, res)
}
