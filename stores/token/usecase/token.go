package usecase

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/erc1155"
	"github.com/x-xyz/goapi/domain/file"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/domain/like"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/domain/token"
	"github.com/x-xyz/goapi/domain/unlockable"
	"github.com/x-xyz/goapi/service/paging"
	"github.com/x-xyz/goapi/service/pinata"
	"github.com/x-xyz/goapi/service/redis"
)

type TokenUseCaseCfg struct {
	LikeRepo           like.Repo
	NftitemRepo        nftitem.Repo
	CollectionRepo     collection.Repo
	UnlockableRepo     unlockable.Repo
	FileUC             file.Usecase
	IpfsUri            string
	ActivityRepo       account.ActivityHistoryRepo
	FolderRelationRepo account.FolderNftRelationshipRepo
	Erc1155HoldingRepo erc1155.HoldingRepo
	OrderItemRepo      order.OrderItemRepo
	Redis              redis.Service
}

type impl struct {
	like               like.Repo
	nftitem            nftitem.Repo
	collection         collection.Repo
	unlockable         unlockable.Repo
	file               file.Usecase
	ipfsUri            string
	activity           account.ActivityHistoryRepo
	folderRelationRepo account.FolderNftRelationshipRepo
	erc1155Holding     erc1155.HoldingRepo
	orderItemRepo      order.OrderItemRepo

	searchV2Paging paging.Service
}

func New(cfg *TokenUseCaseCfg) token.Usecase {
	im := &impl{
		like:               cfg.LikeRepo,
		nftitem:            cfg.NftitemRepo,
		collection:         cfg.CollectionRepo,
		unlockable:         cfg.UnlockableRepo,
		file:               cfg.FileUC,
		ipfsUri:            cfg.IpfsUri,
		activity:           cfg.ActivityRepo,
		folderRelationRepo: cfg.FolderRelationRepo,
		erc1155Holding:     cfg.Erc1155HoldingRepo,
		orderItemRepo:      cfg.OrderItemRepo,
	}

	if cfg.Redis != nil {
		im.searchV2Paging = paging.New(&paging.PagingConfig{
			RedisCache:    cfg.Redis,
			KeyPfx:        keys.PfxSearchV2Paging,
			RenewDuration: 10 * time.Second,
			CacheDuration: 10 * time.Minute,
			Getter:        im.searchV2Getter,
			ShardSize:     100,
		})
	}

	return im
}

func (im *impl) Search(c ctx.Ctx, optFns ...token.SearchOptionsFunc) (*token.SearchResult, error) {
	return nil, domain.ErrDeprecated
}

func getNftitemIdsIntersection(l1, l2 []nftitem.Id) []nftitem.Id {
	m := map[string]struct{}{}
	for _, it := range l1 {
		key := fmt.Sprintf("%v_%s_%s", it.ChainId, it.ContractAddress, it.TokenId)
		m[key] = struct{}{}
	}

	res := []nftitem.Id{}
	for _, it := range l2 {
		key := fmt.Sprintf("%v_%s_%s", it.ChainId, it.ContractAddress, it.TokenId)
		if _, ok := m[key]; ok {
			res = append(res, it)
		}
	}

	return res
}

func (im *impl) searchV2(c ctx.Ctx, opts *token.SearchOptions, withCount bool) (*token.SearchResult, error) {
	findOpts := []nftitem.FindAllOptionsFunc{
		nftitem.WithIndexerStates(nftitem.ReadyToServeIndexerStates),
	}

	if opts.SortBy != nil && opts.SortDir != nil {
		findOpts = append(findOpts, nftitem.WithSort(*opts.SortBy, *opts.SortDir))
	}

	if opts.Sorts != nil {
		findOpts = append(findOpts, nftitem.WithSorts(*opts.Sorts))
	}

	if opts.Limit != nil && opts.Offset != nil {
		findOpts = append(findOpts, nftitem.WithPagination(*opts.Offset, *opts.Limit))
	}

	if opts.ChainId != nil {
		findOpts = append(findOpts, nftitem.WithChainId(*opts.ChainId))
	}

	holdingBalanceMap := map[string]int{}
	if opts.BelongsTo != nil && !opts.BelongsTo.IsEmpty() {
		holdings, err := im.erc1155Holding.FindAll(c, erc1155.WithOwner(*opts.BelongsTo))
		if err != nil {
			c.WithFields(log.Fields{
				"err":   err,
				"owner": *opts.BelongsTo,
			}).Error("failed to FindAll")
			return nil, err
		}
		holdingNftitemIds := []nftitem.Id{}
		for _, holding := range holdings {
			if holding.Balance > 0 {
				nftid := nftitem.Id{
					ChainId:         holding.ChainId,
					ContractAddress: holding.Address.ToLower(),
					TokenId:         holding.TokenId,
				}
				holdingNftitemIds = append(holdingNftitemIds, nftid)
				holdingBalanceMap[nftid.ToString()] = int(holding.Balance)
			}
		}
		if len(holdingNftitemIds) > 0 {
			findOpts = append(findOpts, nftitem.WithHoldingIds(holdingNftitemIds))
		}

		findOpts = append(findOpts, nftitem.WithOwner(*opts.BelongsTo))
	}

	if opts.NotBelongsTo != nil && !opts.NotBelongsTo.IsEmpty() {
		findOpts = append(findOpts, nftitem.WithNotOwner(*opts.NotBelongsTo))
	}

	if opts.ListingFrom != nil && !opts.ListingFrom.IsEmpty() {
		findOpts = append(findOpts, nftitem.WithListingFrom(*opts.ListingFrom))
	}

	if opts.InactiveListingFrom != nil && !opts.ListingFrom.IsEmpty() {
		findOpts = append(findOpts, nftitem.WithInactiveListingFrom(*opts.InactiveListingFrom))
	}

	if len(opts.Attributes) > 0 {
		findOpts = append(findOpts, nftitem.WithAttributeFilters(opts.Attributes))
	}

	contractAddresses, err := im.getContractWhitelist(c, *opts)
	if err != nil {
		c.WithField("err", err).Error("getContractWhitelist failed")
		return nil, err
	}

	if len(contractAddresses) > 0 {
		findOpts = append(findOpts, nftitem.WithContractAddresses(contractAddresses))
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusBuyNow) {
		findOpts = append(findOpts, nftitem.WithBuyNow())
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusHasOffer) {
		findOpts = append(findOpts, nftitem.WithHasOffer())
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusOnAuction) {
		findOpts = append(findOpts, nftitem.WithOnAuction())
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusHasBid) {
		findOpts = append(findOpts, nftitem.WithHasBid())
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusHasTraded) {
		findOpts = append(findOpts, nftitem.WithHasTraded())
	}

	if opts.PriceGTE != nil {
		findOpts = append(findOpts, nftitem.WithPriceGTE(*opts.PriceGTE))
	}

	if opts.PriceLTE != nil {
		findOpts = append(findOpts, nftitem.WithPriceLTE(*opts.PriceLTE))
	}

	if opts.PriceInUsdGTE != nil {
		findOpts = append(findOpts, nftitem.WithPriceInUsdGTE(*opts.PriceInUsdGTE))
	}

	if opts.PriceInUsdLTE != nil {
		findOpts = append(findOpts, nftitem.WithPriceInUsdLTE(*opts.PriceInUsdLTE))
	}

	if opts.OfferPriceInUsdGTE != nil {
		findOpts = append(findOpts, nftitem.WithOfferPriceInUsdGTE(*opts.OfferPriceInUsdGTE))
	}

	if opts.OfferPriceInUsdLTE != nil {
		findOpts = append(findOpts, nftitem.WithOfferPriceInUsdLTE(*opts.OfferPriceInUsdLTE))
	}

	if opts.Name != nil {
		findOpts = append(findOpts, nftitem.WithName(*opts.Name))
	}

	if opts.Search != nil {
		findOpts = append(findOpts, nftitem.WithSearch(*opts.Search))
	}

	if opts.OfferOwners != nil {
		findOpts = append(findOpts, nftitem.WithOfferOwners(opts.OfferOwners))
	}

	if opts.TokenType != nil {
		findOpts = append(findOpts, nftitem.WithTokenType(*opts.TokenType))
	}

	if opts.BidOwner != nil {
		findOpts = append(findOpts, nftitem.WithBidOwner(*opts.BidOwner))
	}

	if opts.ObjectIdLT != nil {
		findOpts = append(findOpts, nftitem.WithObjectIdLT(*opts.ObjectIdLT))
	}

	if opts.HasOrder != nil {
		findOpts = append(findOpts, nftitem.WithHasOrder(*opts.HasOrder))
	}

	folderIds := []nftitem.Id{}
	if opts.FolderId != nil {
		relations, err := im.folderRelationRepo.GetAllRelations(c, account.WithFolderId(*opts.FolderId))
		if err != nil {
			c.WithFields(log.Fields{
				"folderId": *opts.FolderId,
				"err":      err,
			}).Error("failed to folderRelationRepo.GetAllRelationsByFolderID")
			return nil, err
		}
		// if relations equals 0, should not return any nftitem
		if len(relations) == 0 {
			return &token.SearchResult{
				Count: 0,
				Items: []*token.TokenWithDetail{},
			}, nil
		}

		for _, r := range relations {
			folderIds = append(folderIds, *r.ToNftItemId())
		}
	}

	likeIds := []nftitem.Id{}
	if opts.LikedBy != nil {
		likes, err := im.like.FindAll(c, like.WithLiker(*opts.LikedBy))
		if err != nil {
			c.WithFields(log.Fields{
				"likedBy": *opts.LikedBy,
				"err":     err,
			}).Error("failed to like.FindAll")
			return nil, err
		}

		// if likes equals 0, should not return any nftitem
		if len(likes) == 0 {
			return &token.SearchResult{
				Count: 0,
				Items: []*token.TokenWithDetail{},
			}, nil
		}

		for _, l := range likes {
			likeIds = append(likeIds, *l.ToNftItemId())
		}
	}

	if opts.FolderId != nil && opts.LikedBy != nil {
		ids := getNftitemIdsIntersection(folderIds, likeIds)
		findOpts = append(findOpts, nftitem.WithNftitemIds(ids))
	} else if opts.FolderId != nil {
		findOpts = append(findOpts, nftitem.WithNftitemIds(folderIds))
	} else if opts.LikedBy != nil {
		findOpts = append(findOpts, nftitem.WithNftitemIds(likeIds))
	}

	items, err := im.nftitem.FindAll(c, findOpts...)
	if err != nil {
		c.WithField("err", err).Error("nftitem.FindAll failed")
		return nil, err
	}

	details := []*token.TokenWithDetail{}
	for _, item := range items {
		detail := token.TokenWithDetail{NftItem: *item}
		if opts.BelongsTo != nil && item.TokenType == 1155 {
			detail.Balance = ptr.Int(holdingBalanceMap[item.ToId().ToString()])
		}

		if opts.IncludeInactiveOrders != nil && *opts.IncludeInactiveOrders {
			inactiveListings, err := im.getListingsByNftitemId(c, *item.ToId(), false)
			if err != nil {
				c.WithFields(log.Fields{
					"err": err,
				}).Error("failed to im.getListingsByNftitemId")
				return nil, err
			}
			detail.InactiveListings = inactiveListings
		}

		if opts.IncludeOrders != nil && *opts.IncludeOrders {
			listings, err := im.getListingsByNftitemId(c, *item.ToId(), true)
			if err != nil {
				c.WithFields(log.Fields{
					"err": err,
				}).Error("failed to im.getListingsByNftitemId")
				return nil, err
			}
			detail.Listings = listings
			if len(listings) > 0 {
				detail.ActiveListing = listings[0]
			}

			offers, err := im.getOffersByNftitemId(c, *item.ToId(), true)
			if err != nil {
				c.WithFields(log.Fields{
					"err": err,
				}).Error("failed to im.getOffersByNftitemId")
			}
			detail.Offers = offers
		}

		details = append(details, &detail)
	}

	totalCnt := 0
	if withCount {
		cnt, err := im.nftitem.Count(c, findOpts...)
		if err != nil {
			c.WithFields(log.Fields{
				"err": err,
			}).Error("failed to nftitem.Count")
			return nil, err
		}
		totalCnt = cnt
	}

	return &token.SearchResult{
		Items: details,
		Count: totalCnt,
	}, nil
}

func (im *impl) searchV2Getter(c ctx.Ctx, key string) (interface{}, error) {
	opts, err := token.ParseKeyToOptions(key)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
			"key": key,
		}).Error("failed to token.ParseKeyToOptions")
		return nil, err
	}

	res, err := im.searchV2(c, opts, false)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to im.searchV2")
		return nil, err
	}

	return res.Items, nil
}

func (im *impl) SearchV2(c ctx.Ctx, optFns ...token.SearchOptionsFunc) (*token.SearchResult, error) {
	opts, err := token.GetSearchOptions(optFns...)
	if err != nil {
		c.WithField("err", err).Error("token.GetSearchOptions failed")
		return nil, err
	}

	key := token.OptionsToKey(opts)

	if im.searchV2Paging != nil && opts.Cursor != nil {
		// if cursor exists, use paging service
		cursor := *opts.Cursor
		// size default set to 10
		size := 10
		if opts.Size != nil {
			size = *opts.Size
		}
		details := []*token.TokenWithDetail{}

		nextCursor, cnt, err := im.searchV2Paging.Get(c, key, cursor, size, &details)
		if err != nil {
			c.WithFields(log.Fields{
				"err":  err,
				"opts": opts,
			}).Error("failed to searchV2Paging.Get")
			return nil, err
		}

		res := &token.SearchResult{Count: cnt, NextCursor: nextCursor, Items: details}

		return res, nil
	} else {
		// if searchV2Paging == nil, fall back to query db
		// if cursor not exists, bypass caching
		return im.searchV2(c, &opts, true)
	}
}

// getOffersByNftitemId returns offers, error
func (im *impl) getOffersByNftitemId(c ctx.Ctx, id nftitem.Id, isValid bool) ([]*order.OrderItem, error) {
	now := time.Now()

	offers, err := im.orderItemRepo.FindAll(c,
		order.WithIsAsk(false),
		order.WithNftItemId(id),
		order.WithIsValid(isValid),
		order.WithIsUsed(false),
		order.WithStartTimeLT(now),
		order.WithEndTimeGT(now),
	)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to orderRepo.FindAll")
		return nil, err
	}

	collectionOffers, err := im.orderItemRepo.FindAll(c,
		order.WithChainId(id.ChainId),
		order.WithContractAddress(id.ContractAddress),
		order.WithIsValid(isValid),
		order.WithIsUsed(false),
		order.WithStartTimeLT(now),
		order.WithEndTimeGT(now),
		order.WithStrategy(order.StrategyCollectionOffer),
	)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to orderRepo.FindAll")
		return nil, err
	}
	offers = append(offers, collectionOffers...)

	sort.Slice(offers, func(i, j int) bool {
		return offers[i].PriceInUsd > offers[j].PriceInUsd
	})

	return offers, nil
}

// getListingsByNftitemId returns listings, error
func (im *impl) getListingsByNftitemId(c ctx.Ctx, id nftitem.Id, isValid bool) ([]*order.OrderItem, error) {
	now := time.Now()

	listings, err := im.orderItemRepo.FindAll(c,
		order.WithIsAsk(true),
		order.WithNftItemId(id),
		order.WithIsValid(isValid),
		order.WithIsUsed(false),
		order.WithStartTimeLT(now),
		order.WithEndTimeGT(now),
	)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to orderRepo.FindAll")
		return nil, err
	}

	// sort listings, offers
	sort.Slice(listings, func(i, j int) bool {
		return listings[i].PriceInUsd < listings[j].PriceInUsd
	})

	return listings, nil
}

func (im *impl) FindOne(c ctx.Ctx, id nftitem.Id) (*token.TokenWithDetail, error) {
	item, err := im.nftitem.FindOne(c, id.ChainId, id.ContractAddress, id.TokenId)
	if err != nil {
		return nil, err
	}

	listings, err := im.getListingsByNftitemId(c, id, true)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to im.getOrdersByNftitemId")
		return nil, err
	}

	offers, err := im.getOffersByNftitemId(c, id, true)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to im.getOffersByNftitemId")
		return nil, err
	}

	var activeListing *order.OrderItem
	if len(listings) > 0 {
		activeListing = listings[0]
	}

	return &token.TokenWithDetail{NftItem: *item, Listings: listings, Offers: offers, ActiveListing: activeListing}, nil
}

func (im *impl) getContractWhitelist(c ctx.Ctx, opts token.SearchOptions) ([]domain.Address, error) {
	contractMap := map[string]struct{}{}

	// find out collections belong to category
	if opts.Category != nil {
		findOpts := []collection.FindAllOptions{
			collection.WithCategory(*opts.Category),
			collection.WithStatus(true),
			collection.WithPagination(0, 5000),
		}

		if opts.ChainId != nil {
			findOpts = append(findOpts, collection.WithChainId(*opts.ChainId))
		}

		collections, err := im.collection.FindAll(c, findOpts...)

		if err != nil {
			c.WithField("err", err).Error("collection.FindAll failed")
			return nil, err
		}

		for _, collection := range collections {
			//	@todo	fix type
			//	@todo	should compose chain id and contract address as key?
			contractMap[string(collection.Erc721Address)] = struct{}{}
		}
	}

	if len(opts.Collections) > 0 {
		for _, collection := range opts.Collections {
			contractMap[collection.ToLowerStr()] = struct{}{}
		}
	}

	contractWhitelist := []domain.Address{}

	for contract := range contractMap {
		contractWhitelist = append(contractWhitelist, domain.Address(contract))
	}

	return contractWhitelist, nil
}

func (im *impl) AddUnlockableContent(c ctx.Ctx, id nftitem.Id, content string) error {
	v := unlockable.Unlockable{
		ChainId:         id.ChainId,
		ContractAddress: id.ContractAddress,
		TokenId:         id.TokenId,
		Content:         content,
	}

	if err := im.unlockable.Create(c, v); err != nil {
		c.WithField("err", err).WithField("id", id).Error("unlockable.Create failed")
		return err
	}

	return nil
}

func (im *impl) GetUnlockableContent(c ctx.Ctx, id nftitem.Id) (string, error) {
	uid := unlockable.UnlockableId{ChainId: id.ChainId, ContractAddress: id.ContractAddress.ToLower(), TokenId: id.TokenId}

	if res, err := im.unlockable.FindOne(c, uid); err != nil {
		c.WithField("err", err).Error("unlockable.FindOne failed")
		return "", err
	} else {
		return res.Content, nil
	}
}

func (im *impl) BanNftItem(c ctx.Ctx, id nftitem.Id) error {
	if err := im.nftitem.Patch(c, id, nftitem.PatchableNftItem{IsAppropriate: ptr.Bool(false)}); err != nil {
		c.WithField("err", err).WithField("id", id).Error("nftitem.Patch failed")
		return err
	}
	return nil
}

func (im *impl) UnbanNftItem(c ctx.Ctx, id nftitem.Id) error {
	if err := im.nftitem.Patch(c, id, nftitem.PatchableNftItem{IsAppropriate: ptr.Bool(true)}); err != nil {
		c.WithField("err", err).WithField("id", id).Error("nftitem.Patch failed")
		return err
	}
	return nil
}

func (im *impl) Upload(c ctx.Ctx, account domain.Address, payload token.UploadPayload) (*token.UploadResult, error) {
	fileHash, err := im.file.Upload(c, payload.Image, pinata.PinOptions{
		Metadata: &pinata.PinataMetadata{
			Name: payload.Name,
			KeyValues: map[string]interface{}{
				"royalty":   payload.Royalty,
				"IP_Rights": payload.XtraUrl,
				"recipient": account,
			},
		},
	})
	if err != nil {
		c.WithField("err", err).Error("file.Upload failed")
		return nil, err
	}

	type properties struct {
		Address    domain.Address `json:"address"`
		Royalty    string         `json:"royalty"`
		Recipient  domain.Address `json:"recipient"`
		IP_Rights  string         `json:"IP_Rights"`
		CreatedAt  string         `json:"createdAt"`
		Collection string         `json:"collection"`
	}

	type metadata struct {
		Name        string              `json:"name"`
		Image       string              `json:"image"`
		Description string              `json:"description"`
		Properties  properties          `json:"properties"`
		Attributes  []*token.TokenTrait `json:"attributes"`
	}

	m := metadata{
		Name:        payload.Name,
		Image:       fmt.Sprintf("%s%s", "ipfs://", fileHash),
		Description: payload.Description,
		Properties: properties{
			Address:    account,
			Royalty:    payload.Royalty,
			Recipient:  account,
			IP_Rights:  payload.XtraUrl,
			CreatedAt:  time.Now().UTC().Format("2006-01-02T15:04:05-0700"),
			Collection: payload.CollectionName,
		},
		Attributes: payload.Traits,
	}

	jsonHash, err := im.file.UploadJson(c, m, pinata.PinOptions{
		Metadata: &pinata.PinataMetadata{
			Name: payload.Name,
			KeyValues: map[string]interface{}{
				"address": account,
			},
		},
	})
	if err != nil {
		c.WithField("err", err).Error("file.UploadJson failed")
		return nil, err
	}

	return &token.UploadResult{
		FileHash: fmt.Sprintf("%s%s", im.ipfsUri, fileHash),
		JsonHash: fmt.Sprintf("%s%s", im.ipfsUri, jsonHash),
	}, nil
}

func (im *impl) PatchNft(c ctx.Ctx, id *nftitem.Id, patchable *nftitem.PatchableNftItem) error {
	err := im.nftitem.Patch(c, *id, *patchable)
	if err != nil {
		c.WithField("err", err).Error("nftitem.Patch failed")
		return err
	}
	return nil
}

func (im *impl) SearchForIndexerState(c ctx.Ctx, indexerStates []nftitem.IndexerState, retryCountLimit int, optFns ...token.SearchOptionsFunc) ([]*nftitem.NftItem, int, error) {
	opts, err := token.GetSearchOptions(optFns...)
	if err != nil {
		c.WithField("err", err).Error("token.GetSearchOptions failed")
		return nil, 0, err
	}

	findOpts := []nftitem.FindAllOptionsFunc{
		nftitem.WithIndexerStates(indexerStates),
		nftitem.WithIndexerRetryCountLT(retryCountLimit),
	}

	if opts.ChainId != nil {
		findOpts = append(findOpts, nftitem.WithChainId(*opts.ChainId))
	}

	if opts.Collections != nil {
		findOpts = append(findOpts, nftitem.WithContractAddresses(*&opts.Collections))
	}

	if opts.SortBy != nil && opts.SortDir != nil {
		findOpts = append(findOpts, nftitem.WithSort(string(*opts.SortBy), *opts.SortDir))
	}

	if opts.Offset != nil && opts.Limit != nil {
		findOpts = append(findOpts, nftitem.WithPagination(*opts.Offset, *opts.Limit))
	}

	items, err := im.nftitem.FindAll(c, findOpts...)
	if err != nil {
		c.WithField("err", err).Error("nftitem.FindAll failed")
		return nil, 0, err
	}

	count, err := im.nftitem.Count(c, findOpts...)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("nftitem.Count failed")
		return nil, 0, err
	}
	return items, count, nil
}

func (im *impl) GetViewCount(c ctx.Ctx, id nftitem.Id) (int32, error) {
	if _, err := im.nftitem.FindOne(c, id.ChainId, id.ContractAddress, id.TokenId); err != nil {
		c.WithField("err", err).Error("nftitem.FindOne failed")
		return 0, err
	}
	return im.nftitem.IncreaseViewCount(c, id, 1)
}

func (im *impl) UpsertListing(c ctx.Ctx, id nftitem.Id, listing *nftitem.Listing, overrideActive bool) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) RemoveListing(c ctx.Ctx, id nftitem.Id, owner *domain.Address) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) SetActiveListingTo(c ctx.Ctx, id nftitem.Id, owner *domain.Address) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) GetListing(c ctx.Ctx, id nftitem.Id, owner *domain.Address) (*nftitem.Listing, error) {
	panic(domain.ErrDeprecated)
}

func (im *impl) UpsertOffer(c ctx.Ctx, id nftitem.Id, offer *nftitem.Offer) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) RemoveOffer(c ctx.Ctx, id nftitem.Id, offerer *domain.Address) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) GetOffer(c ctx.Ctx, id nftitem.Id, offerer *domain.Address) (*nftitem.Offer, error) {
	panic(domain.ErrDeprecated)
}

func (im *impl) SetAuction(c ctx.Ctx, id nftitem.Id, auction *nftitem.Auction) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) UpdateAuction(c ctx.Ctx, id nftitem.Id, auction *nftitem.Auction) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) ClearAuction(c ctx.Ctx, id nftitem.Id) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) SetHighestBid(c ctx.Ctx, id nftitem.Id, bid *nftitem.Bid) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) ClearHighestBid(c ctx.Ctx, id nftitem.Id) error {
	panic(domain.ErrDeprecated)
}

func (im *impl) GetActivities(c ctx.Ctx, id nftitem.Id, offset, limit int) (*token.ActivityResult, error) {
	res := token.ActivityResult{}

	items, err := im.activity.FindActivities(
		c,
		account.ActivityHistoryWithToken(id.ChainId, id.ContractAddress, id.TokenId),
		account.ActivityHistoryWithPagination(offset, limit),
		account.ActivityHistoryWithSource(account.SourceX),
	)
	if err != nil {
		c.WithField("err", err).WithField("id", id).Error("activity.FindTokenActivities failed")
		return nil, err
	}

	res.Items = items

	cnt, err := im.activity.CountActivities(
		c,
		account.ActivityHistoryWithToken(id.ChainId, id.ContractAddress, id.TokenId),
		account.ActivityHistoryWithPagination(offset, limit),
		account.ActivityHistoryWithSource(account.SourceX),
	)
	if err != nil {
		c.WithField("err", err).WithField("id", id).Error("activity.CountTokenActivities failed")
		return nil, err
	}

	res.Count = cnt

	return &res, nil
}

func (im *impl) GetPriceHistories(c ctx.Ctx, id nftitem.Id, period domain.TimePeriod) ([]token.PriceHistory, error) {
	opts := []account.FindActivityHistoryOptions{
		account.ActivityHistoryWithToken(id.ChainId, id.ContractAddress, id.TokenId),
		account.ActivityHistoryWithTypes(
			account.ActivityHistoryTypeSold,
			account.ActivityHistoryTypeOfferTaken,
			account.ActivityHistoryTypeResultAuction,
		),
	}

	if !period.IsAll() {
		opts = append(opts, account.ActivityHistoryWithTimeGTE(time.Now().Add(-period.ToDuration())))
	}

	ahs, err := im.activity.FindActivities(c, opts...)

	if err != nil {
		return nil, err
	}

	m := map[time.Time][]account.ActivityHistory{}

	for _, ah := range ahs {
		time := ah.Time.UTC().Truncate(domain.TimePeriodDay.ToDuration())
		m[time] = append(m[time], ah)
	}

	res := []token.PriceHistory{}

	for time, ahs := range m {
		sumOfUsd := float64(0)
		sumOfNative := float64(0)
		priceInUsd := float64(0)
		priceInNative := float64(0)
		for _, ah := range ahs {
			sumOfUsd += ah.PriceInUsd
			sumOfNative += ah.PriceInNative
		}
		if len(ahs) > 0 {
			priceInUsd = sumOfUsd / float64(len(ahs))
			priceInNative = sumOfNative / float64(len(ahs))
		}
		res = append(res, token.PriceHistory{
			Time:          time,
			PriceInUsd:    priceInUsd,
			PriceInNative: priceInNative,
		})
	}

	sort.Slice(res, func(a, b int) bool {
		return res[a].Time.After(res[b].Time)
	})

	return res, nil
}

func (im *impl) EnsureNftExists(c ctx.Ctx, id nftitem.Id) (*nftitem.NftItem, error) {
	if nft, err := im.FindOne(c, id); err == nil {
		return &nft.NftItem, nil
	} else if err != domain.ErrNotFound {
		return nil, err
	}

	nft := &nftitem.NftItem{
		ChainId:         id.ChainId,
		ContractAddress: id.ContractAddress,
		TokenId:         id.TokenId,
		NumOwners:       1,
		IsAppropriate:   ptr.Bool(true),
		ThumbnailPath:   "-",
		ImagePath:       "-",
		ContentType:     "image",
		TokenType:       domain.TokenType721,
		IndexerState:    nftitem.IndexerStateNew,
	}

	collection, err := im.collection.FindOne(c, collection.CollectionId{
		ChainId: id.ChainId,
		Address: id.ContractAddress.ToLower(),
	})
	if err == nil {
		nft.Creator = collection.Owner
	}

	if err := im.nftitem.Create(c, nft); err != nil {
		c.WithField("err", err).Error("nftitem.Create failed")
		return nil, err
	}

	return nft, nil
}

func (im *impl) RefreshIndexerState(c ctx.Ctx, id nftitem.Id) error {
	_, err := im.nftitem.FindOne(c, id.ChainId, id.ContractAddress, id.TokenId)
	if err != nil {
		c.WithFields(log.Fields{
			"id":  id,
			"err": err,
		}).Error("nftitem.FindOne failed")
		return err
	}
	s := nftitem.IndexerState(nftitem.IndexerStateNewRefreshing)
	patchable := nftitem.PatchableNftItem{IndexerState: &s, IndexerRetryCount: ptr.Int32(0)}
	if err := im.nftitem.Patch(c, id, patchable); err != nil {
		c.WithFields(log.Fields{
			"id":        id,
			"patchable": patchable,
			"err":       err,
		}).Error("nftitem.Patch failed")
		return err
	}
	return nil
}

func (im *impl) RefreshListingAndOfferState(ctx ctx.Ctx, id nftitem.Id) error {
	now := time.Now()
	orderItems, err := im.orderItemRepo.FindAll(
		ctx,
		order.WithStrategy(order.StrategyFixedPrice),
		order.WithNftItemId(id),
		order.WithIsUsed(false),
		order.WithEndTimeGT(now),
		order.WithStartTimeLT(now),
	)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to orderRepo.FindAll")
		return err
	}

	collectionOffers, err := im.orderItemRepo.FindAll(
		ctx,
		order.WithChainId(id.ChainId),
		order.WithContractAddress(id.ContractAddress),
		order.WithIsValid(true),
		order.WithIsUsed(false),
		order.WithStartTimeLT(now),
		order.WithEndTimeGT(now),
		order.WithStrategy(order.StrategyCollectionOffer),
	)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to orderItemRepo.FindAll")
		return err
	}
	orderItems = append(orderItems, collectionOffers...)

	listingEndsAt := time.Time{}
	listingOwnerMap := map[domain.Address]struct{}{}
	inactiveListingOwnerMap := map[domain.Address]struct{}{}
	offerStartsAt := time.Time{}
	offerEndsAt := time.Time{}
	offerOwnerMap := map[domain.Address]struct{}{}
	instantLiquidityInUsd := float64(0)
	for _, od := range orderItems {
		if od.IsAsk {
			if od.IsValid {
				if listingEndsAt.IsZero() || od.EndTime.After(listingEndsAt) {
					listingEndsAt = od.EndTime
				}
				listingOwnerMap[od.Signer] = struct{}{}
			} else {
				inactiveListingOwnerMap[od.Signer] = struct{}{}
			}
		} else {
			if offerStartsAt.IsZero() || od.StartTime.Before(offerStartsAt) {
				offerStartsAt = od.StartTime
			}
			if offerEndsAt.IsZero() || od.EndTime.After(offerEndsAt) {
				offerEndsAt = od.EndTime
			}
			offerOwnerMap[od.Signer] = struct{}{}
			if od.PriceInUsd > instantLiquidityInUsd {
				instantLiquidityInUsd = od.PriceInUsd
			}
		}
	}

	listingOwners := []domain.Address{}
	for owner := range listingOwnerMap {
		listingOwners = append(listingOwners, owner)
	}

	inactiveListingOwners := []domain.Address{}
	for owner := range inactiveListingOwnerMap {
		inactiveListingOwners = append(inactiveListingOwners, owner)
	}

	offerOwners := []domain.Address{}
	for owner := range offerOwnerMap {
		offerOwners = append(offerOwners, owner)
	}

	latestPriceSource := order.ResolveLatestPrice(orderItems)

	hasActiveListings := !listingEndsAt.IsZero()
	hasOrder := !listingEndsAt.IsZero() || !offerEndsAt.IsZero()

	patchable := nftitem.PatchableNftItem{
		HasActiveListings:     &hasActiveListings,
		ListingEndsAt:         &listingEndsAt,
		ListingOwners:         listingOwners,
		InactiveListingOwners: inactiveListingOwners,
		OfferStartsAt:         &offerStartsAt,
		OfferEndsAt:           &offerEndsAt,
		OfferOwners:           offerOwners,
		Price:                 &latestPriceSource.Price,
		PaymentToken:          &latestPriceSource.PaymentToken,
		PriceInUsd:            &latestPriceSource.PriceInUsd,
		PriceSource:           &latestPriceSource.Source,
		InstantLiquidityInUsd: &instantLiquidityInUsd,
		HasOrder:              &hasOrder,
	}

	err = im.nftitem.Patch(ctx, id, patchable)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"id":        id,
			"patchable": patchable,
		}).Error("failed to nftitem.Patch")
		return err
	}

	return nil
}

func (im *impl) GetOpenRararityScore(ctx ctx.Ctx, id nftitem.Id) (float64, error) {
	col, err := im.collection.FindOne(ctx, collection.CollectionId{
		ChainId: id.ChainId,
		Address: id.ContractAddress.ToLower(),
	})
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to collection.FindOne")
		return 0, err
	}

	totalSupply := col.Supply

	colAttrs := col.Attributes
	allTraitTypes := map[string]bool{}
	// add null attr
	for traitType, traitValues := range colAttrs {
		sum := int64(0)
		for _, count := range traitValues {
			sum += count
		}
		if sum < totalSupply {
			colAttrs[traitType]["Null"] = totalSupply - sum
		}
		allTraitTypes[traitType] = false
	}

	collectionEntropy := float64(0)
	for _, traitValues := range colAttrs {
		for _, count := range traitValues {
			p := float64(count) / float64(totalSupply)
			collectionEntropy += -p * math.Log2(p)
		}
	}

	item, err := im.nftitem.FindOne(ctx, id.ChainId, id.ContractAddress, id.TokenId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to nftitem.FindOne")
		return 0, err
	}

	probablities := []float64{}
	itemAttrs := item.Attributes
	for _, attr := range itemAttrs {
		allTraitTypes[attr.TraitType] = true
		probablities = append(probablities, float64(colAttrs[attr.TraitType][attr.Value])/float64(totalSupply))
	}

	for traitType, exists := range allTraitTypes {
		if !exists {
			probablities = append(probablities, float64(colAttrs[traitType]["Null"])/float64(totalSupply))
		}
	}

	score := float64(0)
	for _, p := range probablities {
		score += -math.Log2(p)
	}

	return score / collectionEntropy, nil
}
