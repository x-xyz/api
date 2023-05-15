package usecase

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/shopspring/decimal"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/collection_promotion"
	"github.com/x-xyz/goapi/domain/erc1155"
	erc721contract "github.com/x-xyz/goapi/domain/erc721/contract"
	"github.com/x-xyz/goapi/domain/file"
	"github.com/x-xyz/goapi/domain/like"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/domain/token"
	"github.com/x-xyz/goapi/service/chain/contract"
	"github.com/x-xyz/goapi/service/pinata"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CollectionUseCaseCfg struct {
	CollectionRepo        collection.Repo
	RegistrationRepo      collection.RegistrationRepo
	Erc721contractRepo    erc721contract.Repo
	Erc1155contractRepo   erc1155.Repo
	Erc1155holdingRepo    erc1155.HoldingRepo
	FileUC                file.Usecase
	Erc721ChainService    contract.Erc721Contract
	Erc1155ChainService   contract.Erc1155Contract
	NftitemRepo           nftitem.Repo
	OpenseaDataRepo       domain.OpenseaDataRepo
	ChainlinkUC           domain.ChainlinkUsacase
	FloorPriceHistoryRepo collection.FloorPriceHistoryRepo
	OrderItemRepo         order.OrderItemRepo
	ActivityHistoryRepo   account.ActivityHistoryRepo
	LikeRepo              like.Repo
	PromotedCollectionsUC collection_promotion.CollPromotionUsecase
	TokenUC               token.Usecase
}

type impl struct {
	collection            collection.Repo
	registration          collection.RegistrationRepo
	erc721contract        erc721contract.Repo
	erc1155contract       erc1155.Repo
	erc1155holding        erc1155.HoldingRepo
	file                  file.Usecase
	erc721ChainService    contract.Erc721Contract
	erc1155ChainService   contract.Erc1155Contract
	nftitem               nftitem.Repo
	openseaData           domain.OpenseaDataRepo
	chainlink             domain.ChainlinkUsacase
	floorPriceHistory     collection.FloorPriceHistoryRepo
	orderItemRepo         order.OrderItemRepo
	activityHistoryRepo   account.ActivityHistoryRepo
	likeRepo              like.Repo
	promotedCollectionsUC collection_promotion.CollPromotionUsecase
	tokenUC               token.Usecase
}

func NewCollection(cfg *CollectionUseCaseCfg) collection.Usecase {
	return &impl{
		collection:            cfg.CollectionRepo,
		registration:          cfg.RegistrationRepo,
		erc721contract:        cfg.Erc721contractRepo,
		erc1155contract:       cfg.Erc1155contractRepo,
		erc1155holding:        cfg.Erc1155holdingRepo,
		file:                  cfg.FileUC,
		erc721ChainService:    cfg.Erc721ChainService,
		erc1155ChainService:   cfg.Erc1155ChainService,
		nftitem:               cfg.NftitemRepo,
		openseaData:           cfg.OpenseaDataRepo,
		chainlink:             cfg.ChainlinkUC,
		floorPriceHistory:     cfg.FloorPriceHistoryRepo,
		orderItemRepo:         cfg.OrderItemRepo,
		activityHistoryRepo:   cfg.ActivityHistoryRepo,
		likeRepo:              cfg.LikeRepo,
		promotedCollectionsUC: cfg.PromotedCollectionsUC,
		tokenUC:               cfg.TokenUC,
	}
}

func (im *impl) FindAll(c ctx.Ctx, opts ...collection.FindAllOptions) (*collection.SearchResult, error) {
	options, err := collection.GetFindAllOptions(opts...)
	if err != nil {
		c.WithField("err", err).Error("failed to collection.GetFindAllOptions")
		return nil, err
	}

	if options.LikedBy != nil {
		likes, err := im.likeRepo.FindAll(c, like.WithLiker(*options.LikedBy))
		if err != nil {
			c.WithFields(log.Fields{
				"err": err,
			}).Error("failed to likeRepo.FindAll")
			return nil, err
		}

		addresses := []domain.Address{}

		for _, l := range likes {
			addresses = append(addresses, l.ContractAddress.ToLower())
		}

		opts = append(opts, collection.WithAddresses(addresses))
	}

	if options.ListedBy != nil {
		listedOrders, err := im.orderItemRepo.FindAll(
			c,
			order.WithSigner(*options.ListedBy),
			order.WithIsAsk(true),
			order.WithIsValid(true),
			order.WithIsUsed(false),
		)
		if err != nil {
			c.WithFields(log.Fields{
				"err": err,
			}).Error("failed to orderItemRepo.FindAll")
			return nil, err
		}

		addresses := []domain.Address{}

		for _, l := range listedOrders {
			addresses = append(addresses, l.Collection.ToLower())
		}

		opts = append(opts, collection.WithAddresses(addresses))
	}

	if options.OfferedBy != nil {
		offeredOrders, err := im.orderItemRepo.FindAll(
			c,
			order.WithSigner(*options.OfferedBy),
			order.WithIsAsk(false),
			order.WithIsValid(true),
			order.WithIsUsed(false),
		)
		if err != nil {
			c.WithFields(log.Fields{
				"err": err,
			}).Error("failed to orderItemRepo.FindAll")
			return nil, err
		}

		addresses := []domain.Address{}

		for _, l := range offeredOrders {
			addresses = append(addresses, l.Collection.ToLower())
		}

		opts = append(opts, collection.WithAddresses(addresses))
	}

	opts = append(opts, collection.WithStatus(true))

	items, err := im.collection.FindAll(c, opts...)
	if err != nil {
		c.WithField("err", err).Error("FindAll failed")
		return nil, err
	}

	for _, col := range items {
		col.IsRegistered = true
	}

	cnt, err := im.collection.Count(c, opts...)
	if err != nil {
		c.WithField("err", err).Error("FindAll failed")
		return nil, err
	}

	res := &collection.SearchResult{
		Items: []*collection.CollectionWithHoldingCount{},
		Count: cnt,
	}

	for _, i := range items {
		res.Items = append(res.Items, &collection.CollectionWithHoldingCount{Collection: *i})
	}

	return res, nil
}

func (im *impl) FindAllIncludingUnregistered(c ctx.Ctx, optFns ...collection.FindAllOptions) ([]*collection.CollectionWithHoldingCount, error) {
	optFns = append(optFns, collection.WithPagination(0, 5000))

	m := map[string]*collection.CollectionWithHoldingCount{}

	cols, err := im.FindAll(c, optFns...)
	if err != nil {
		c.WithField("err", err).Error("FindAll failed")
		return nil, err
	}

	for _, col := range cols.Items {
		col.IsRegistered = true
		m[collection.ToCollectionKey(col.ChainId, col.Erc721Address)] = col
	}

	opts, err := collection.GetFindAllOptions(optFns...)
	if err != nil {
		c.WithField("err", err).Error("collection.GetFindAllOptions failed")
		return nil, err
	}

	erc721Opts := []erc721contract.FindOptions{
		erc721contract.WithPagination(0, 5000),
	}

	erc1155Opts := []erc1155.FindOptions{
		erc1155.WithPagination(0, 5000),
	}

	if opts.ChainId != nil {
		erc721Opts = append(erc721Opts, erc721contract.WithChainId(*opts.ChainId))
		erc1155Opts = append(erc1155Opts, erc1155.WithChainId(*opts.ChainId))
	}

	if opts.IsAppropriate != nil {
		erc721Opts = append(erc721Opts, erc721contract.WithIsAppropriate(*opts.IsAppropriate))
		erc1155Opts = append(erc1155Opts, erc1155.WithIsAppropriate(*opts.IsAppropriate))
	}

	erc721s, err := im.erc721contract.FindAll(c, erc721Opts...)
	if err != nil {
		c.WithField("err", err).Error("erc721contract.FindAll failed")
		return nil, err
	}

	for _, erc721 := range erc721s {
		k := collection.ToCollectionKey(erc721.ChainId, erc721.Address)
		// dedup
		if _, ok := m[k]; ok {
			continue
		}
		m[k] = &collection.CollectionWithHoldingCount{Collection: *erc721.ToCollection()}
	}

	erc1155s, err := im.erc1155contract.FindAll(c, erc1155Opts...)
	if err != nil {
		c.WithField("err", err).Error("erc1155.FindAll failed")
		return nil, err
	}

	for _, e := range erc1155s {
		k := collection.ToCollectionKey(e.ChainId, e.Address)
		if _, ok := m[k]; ok {
			continue
		}
		m[k] = &collection.CollectionWithHoldingCount{Collection: *e.ToCollection()}
	}

	res := []*collection.CollectionWithHoldingCount{}

	for _, col := range m {
		res = append(res, col)
	}

	return res, nil
}

func (im *impl) FindAllMintable(c ctx.Ctx, eoa domain.Address, opts ...collection.FindAllOptions) ([]*collection.Collection, error) {
	internalOpts := append(
		opts,
		collection.WithIsInternal(true),
		collection.WithIsOwnerble(false),
		collection.WithIsAppropriate(true),
	)

	ownedOpts := append(
		opts,
		collection.WithOwner(eoa),
		collection.WithIsInternal(true),
		collection.WithIsOwnerble(true),
		collection.WithIsAppropriate(true),
	)

	if internals, err := im.collection.FindAll(c, internalOpts...); err != nil {
		c.WithField("err", err).Error("collection.FindAll for internals failed")
		return nil, err
	} else if owneds, err := im.collection.FindAll(c, ownedOpts...); err != nil {
		c.WithField("err", err).Error("collection.FindAll for owneds failed")
		return nil, err
	} else {
		return append(internals, owneds...), nil
	}
}

func (im *impl) FindAllUnreviewd(c ctx.Ctx, eoa domain.Address) ([]*collection.Registration, error) {
	if res, err := im.registration.FindAll(c); err != nil {
		c.WithField("err", err).Error("collection.FindAll failed")
		return nil, err
	} else {
		return res, nil
	}
}

func (im *impl) FindOne(c ctx.Ctx, id collection.CollectionId) (*collection.Collection, error) {
	col, err := im.collection.FindOne(c, id)

	if err != nil && err != domain.ErrNotFound {
		c.WithField("err", err).Error("collection.FindOne failed")
		return nil, err
	} else if col != nil {
		col.IsRegistered = true
		return col, nil
	}

	erc721, err := im.erc721contract.FindOne(c, erc721contract.WithChainId(id.ChainId), erc721contract.WithAddress(id.Address))

	if err != nil && err != domain.ErrNotFound {
		c.WithField("err", err).Error("collection.FindOne failed")
		return nil, err
	} else if erc721 != nil {
		return erc721.ToCollection(), nil
	}

	erc1155c, err := im.erc1155contract.FindOne(c, erc1155.WithChainId(id.ChainId), erc1155.WithAddress(id.Address))
	if err != nil && err != domain.ErrNotFound {
		c.WithField("err", err).Error("collection.FindOne failed")
		return nil, err
	} else if erc1155c != nil {
		return erc1155c.ToCollection(), nil
	}

	return nil, domain.ErrNotFound
}

func (im *impl) FindOneWithStat(c ctx.Ctx, id collection.CollectionId) (*collection.CollectionWithStat, error) {
	col, err := im.FindOne(c, id)
	if err != nil {
		return nil, err
	}

	ethChainId := domain.ChainId(1)
	apePrice, err := im.chainlink.GetLatestAnswer(c, ethChainId, apeAddr)
	if err != nil {
		c.WithField("err", err).Error("chainlink.GetLatestAnswer failed")
		return nil, err
	}

	res := &collection.CollectionWithStat{
		Collection: *col,
	}

	osDataId := domain.OpenseaDataId{
		ChainId: id.ChainId,
		Address: id.Address,
	}
	osData, err := im.openseaData.FindOne(c, osDataId)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
			"id":  osDataId,
		}).Error("failed to openseaData.FindOne")
	}

	if osData != nil {
		res.OpenseaSalesVolume = osData.OneDayVolume
		res.OpenseaSalesVolumeChange = osData.OneDayChange
		res.OpenseaFloorPriceInApe = decimal.NewFromFloat(col.OpenseaFloorPriceInUsd).Div(apePrice).InexactFloat64()
	}

	return res, nil
}

func (im *impl) CreateErc1155(c ctx.Ctx, value collection.CreatePayload) error {
	if err := im.collection.Create(c, value); err != nil {
		c.WithFields(log.Fields{
			"err":     err,
			"chainId": value.ChainId,
			"address": value.Erc721Address,
		}).Error("collection.Create failed")
		return err
	}

	contract := erc1155.Contract{
		ChainId:       value.ChainId,
		Address:       value.Erc721Address.ToLower(),
		Name:          value.CollectionName,
		IsVerified:    false,
		IsAppropriate: true,
	}

	if err := im.erc1155contract.Create(c, contract); err != nil {
		c.WithField("err", err).WithField("erc1155contract", contract).Error("erc1155contract.Create failed")
		return err
	}
	return nil
}

func (im *impl) CreateErc721(c ctx.Ctx, value collection.CreatePayload) error {
	if err := im.collection.Create(c, value); err != nil {
		c.WithFields(log.Fields{
			"err":     err,
			"chainId": value.ChainId,
			"address": value.Erc721Address,
		}).Error("collection.Create failed")
		return err
	}
	erc721contract := erc721contract.Contract{
		ChainId:       value.ChainId,
		Address:       value.Erc721Address.ToLower(),
		Name:          value.CollectionName,
		IsVerified:    false,
		IsAppropriate: true,
	}
	if err := im.erc721contract.Create(c, erc721contract); err != nil {
		c.WithField("err", err).WithField("erc721contract", erc721contract).Error("erc721contract.Create failed")
		return err
	}
	return nil
}

func (im *impl) Register(c ctx.Ctx, value collection.Registration) (*collection.Registration, error) {
	if value.TokenType == domain.TokenType1155 {
		ok, err := im.erc1155ChainService.Supports1155Interface(c, int32(value.ChainId), value.Erc721Address.ToLowerStr())
		if err != nil {
			c.WithFields(log.Fields{
				"err":     err,
				"chainId": value.ChainId,
				"address": value.Erc721Address,
			}).Error("erc1155ChainService.Supports1155Interface failed")
			return nil, err
		}
		if !ok {
			return nil, domain.ErrErc1155InterfaceUnsupported
		}
	} else {
		value.TokenType = domain.TokenType721
		ok, err := im.erc721ChainService.Supports721Interface(c, int32(value.ChainId), value.Erc721Address.ToLowerStr())
		if err != nil {
			c.WithFields(log.Fields{
				"err":     err,
				"chainId": value.ChainId,
				"address": value.Erc721Address,
			}).Error("erc721ChainService.Supports721Interface failed")
			return nil, err
		}
		if !ok {
			return nil, domain.ErrErc721InterfaceUnsupported
		}
	}

	if len(value.LogoImage) > 0 {
		opts := pinata.PinOptions{
			Metadata: &pinata.PinataMetadata{
				Name: fmt.Sprintf("%d-%s", value.ChainId, value.Erc721Address),
			},
			Options: &pinata.PinataOptions{
				CidVersion: pinata.CidVersion_0,
			},
		}

		hash, err := im.file.Upload(c, value.LogoImage, opts)
		if err != nil {
			return nil, err
		}
		value.LogoImageHash = hash
	}

	if len(value.CoverImage) > 0 {
		opts := pinata.PinOptions{
			Metadata: &pinata.PinataMetadata{
				Name: fmt.Sprintf("%d-%s-cover", value.ChainId, value.Erc721Address),
			},
			Options: &pinata.PinataOptions{
				CidVersion: pinata.CidVersion_0,
			},
		}

		hash, err := im.file.Upload(c, value.CoverImage, opts)
		if err != nil {
			return nil, err
		}
		value.CoverImageHash = hash
	}
	if err := im.registration.Create(c, value); err != nil {
		c.WithField("err", err).Error("registration.Create failed")
		return nil, err
	}

	col, err := im.registration.FindOne(c, collection.CollectionId{ChainId: value.ChainId, Address: value.Erc721Address})
	if err != nil {
		c.WithField("err", err).Error("collection.FindOne failed")
		return nil, err
	}
	return col, nil
}

func (im *impl) Accept(c ctx.Ctx, id collection.CollectionId) (*collection.Collection, error) {
	reg, err := im.registration.FindOne(c, id)
	if err != nil {
		c.WithField("err", err).Error("registration.FindOne failed")
		return nil, err
	}

	if !common.IsHexAddress(reg.FeeRecipient) {
		//	@todo	send email
		//	@todo	remove collection
		return nil, domain.ErrInvalidAddress
	}

	royalty := reg.Royalty * 100

	if royalty > 10000 || royalty < 0 {
		//	@todo	send email
		//	@todo	remove collection
		return nil, errors.New("invalid royalty")
	}

	//	@todo	call marketplace.registerCollectionRoyalty

	col := collection.CreatePayload{
		ChainId:         id.ChainId,
		Erc721Address:   id.Address,
		TokenType:       reg.TokenType,
		Owner:           reg.Owner,
		Email:           reg.Email,
		CollectionName:  reg.CollectionName,
		Description:     reg.Description,
		Categories:      reg.Categories,
		LogoImageHash:   reg.LogoImageHash,
		LogoImageUrl:    reg.LogoImageUrl,
		CoverImageHash:  reg.CoverImageHash,
		CoverImageURL:   reg.CoverImageURL,
		SiteUrl:         reg.SiteUrl,
		Discord:         reg.Discord,
		TwitterHandle:   reg.TwitterHandle,
		InstagramHandle: reg.InstagramHandle,
		MediumHandle:    reg.MediumHandle,
		Telegram:        reg.Telegram,
		Royalty:         reg.Royalty,
		FeeRecipient:    reg.FeeRecipient,
		Status:          true,
		IsAppropriate:   true,
		IsInternal:      false,
		IsOwnerble:      false,
		IsVerified:      false,
	}

	if err := im.collection.Upsert(c, col); err != nil {
		c.WithField("err", err).Error("collection.Upsert failed")
		return nil, err
	}

	if reg.TokenType == domain.TokenType1155 {
		if _, err := im.erc1155contract.FindOne(c, erc1155.WithChainId(reg.ChainId), erc1155.WithAddress(reg.Erc721Address)); err == domain.ErrNotFound {
			erc1155contract := erc1155.Contract{
				ChainId:       reg.ChainId,
				Address:       reg.Erc721Address.ToLower(),
				Name:          reg.CollectionName,
				IsVerified:    false,
				IsAppropriate: true,
			}

			if err := im.erc1155contract.Create(c, erc1155contract); err != nil {
				c.WithField("err", err).WithField("erc1155contract", erc1155contract).Error("erc1155contract.Create failed")
				return nil, err
			}
		} else if err != nil {
			c.WithField("err", err).Error("erc1155contract.FindOne failed")
			return nil, err
		} else {
			if err := im.erc1155contract.Update(c,
				erc1155.UpdatePayload{IsAppropriate: ptr.Bool(true)},
				erc1155.WithChainId(id.ChainId), erc1155.WithAddress(id.Address)); err != nil {
				c.WithField("err", err).Error("erc1155contract.Update failed")
				return nil, err
			}
		}

	} else {
		if _, err := im.erc721contract.FindOne(c, erc721contract.WithChainId(reg.ChainId), erc721contract.WithAddress(reg.Erc721Address)); err == domain.ErrNotFound {
			erc721contract := erc721contract.Contract{
				ChainId:       reg.ChainId,
				Address:       reg.Erc721Address.ToLower(),
				Name:          reg.CollectionName,
				IsVerified:    false,
				IsAppropriate: true,
			}

			if err := im.erc721contract.Create(c, erc721contract); err != nil {
				c.WithField("err", err).WithField("erc721contract", erc721contract).Error("erc721contract.Create failed")
				return nil, err
			}
		} else if err != nil {
			c.WithField("err", err).Error("erc721contract.FindOne failed")
			return nil, err
		} else {
			if err := im.erc721contract.Update(c,
				erc721contract.UpdatePayload{IsAppropriate: ptr.Bool(true)},
				erc721contract.WithChainId(id.ChainId), erc721contract.WithAddress(id.Address)); err != nil {
				c.WithField("err", err).Error("erc721contract.Update failed")
				return nil, err
			}
		}
	}

	if err := im.registration.Patch(c, id, collection.UpdateRegistration{State: collection.RegistrationStateAccept}); err != nil {
		return nil, err
	}

	//	@todo	send email

	return im.FindOne(c, id)
}

func (im *impl) Reject(c ctx.Ctx, id collection.CollectionId, reason string) (*collection.Registration, error) {
	//	@todo	send email

	if err := im.registration.Patch(c, id, collection.UpdateRegistration{State: collection.RegistrationStateReject}); err != nil {
		return nil, err
	}

	return im.registration.FindOne(c, id)
}

func (im *impl) Ban(c ctx.Ctx, id collection.CollectionId, ban bool) (*collection.Collection, error) {
	if err := im.collection.Update(c, id, collection.UpdatePayload{IsAppropriate: ptr.Bool(!ban)}); err != nil && err != domain.ErrNotFound {
		c.WithField("err", err).Error("collection.Update failed")
		return nil, err
	}

	col, err := im.FindOne(c, id)
	if err != nil {
		c.WithField("err", err).Error("FindOne failed")
		return nil, err
	}

	if col.TokenType == domain.TokenType1155 {
		if err := im.erc1155contract.Update(c, erc1155.UpdatePayload{IsAppropriate: ptr.Bool(false)}, erc1155.WithChainId(id.ChainId), erc1155.WithAddress(id.Address)); err != nil && err != domain.ErrNotFound {
			c.WithField("err", err).Error("erc1155contract.Update failed")
			return nil, err
		}
	} else {
		if err := im.erc721contract.Update(c, erc721contract.UpdatePayload{IsAppropriate: ptr.Bool(false)}, erc721contract.WithChainId(id.ChainId), erc721contract.WithAddress(id.Address)); err != nil && err != domain.ErrNotFound {
			c.WithField("err", err).Error("erc721contract.Update failed")
			return nil, err
		}
	}
	return col, nil
}

func (im *impl) getNftitemScore(ctx ctx.Ctx, item *nftitem.NftItem, supply int64, attrs map[string]map[string]int64, collectionEntropy float64) (float64, error) {
	allTraitTypes := map[string]bool{}
	for traitType := range attrs {
		allTraitTypes[traitType] = false
	}

	probablities := []float64{}
	itemAttrs := item.Attributes
	for _, attr := range itemAttrs {
		allTraitTypes[attr.TraitType] = true
		probablities = append(probablities, float64(attrs[attr.TraitType][attr.Value])/float64(supply))
	}

	for traitType, exists := range allTraitTypes {
		if !exists {
			probablities = append(probablities, float64(attrs[traitType]["Null"])/float64(supply))
		}
	}

	score := float64(0)
	for _, p := range probablities {
		score += -math.Log2(p)
	}

	return score / collectionEntropy, nil
}

func (im *impl) calculateOpenrarityScoreAndRank(ctx ctx.Ctx, id collection.CollectionId) error {
	col, err := im.FindOne(ctx, id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to FindOne")
		return err
	}

	attrs := col.Attributes

	hash, err := hashstructure.Hash(attrs, hashstructure.FormatV2, nil)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to hashstructure.Hash")
		return err
	}
	attrsHash := fmt.Sprint(hash)

	if attrsHash == col.AttributesHash {
		return nil
	}

	supply := col.Supply

	// calculate Null attribute
	for traitType, traitValues := range attrs {
		sum := int64(0)
		for _, count := range traitValues {
			sum += count
		}
		if sum < supply {
			attrs[traitType]["Null"] = supply - sum
		}
	}

	collectionEntropy := float64(0)
	for _, traitValues := range attrs {
		for _, count := range traitValues {
			p := float64(count) / float64(supply)
			collectionEntropy += -p * math.Log2(p)
		}
	}

	type itemWithScore struct {
		id    nftitem.Id
		score float64
	}
	itemSlice := []itemWithScore{}
	limit := int32(500)
	lastObjectId := primitive.NewObjectID()
	for {
		items, err := im.nftitem.FindAll(
			ctx,
			nftitem.WithChainId(id.ChainId),
			nftitem.WithContractAddresses([]domain.Address{id.Address}),
			nftitem.WithPagination(0, limit),
			nftitem.WithObjectIdLT(lastObjectId),
			nftitem.WithIndexerStates(nftitem.ReadyToServeIndexerStates),
			nftitem.WithSorts([]string{"chainId", "contractAddress", "indexerState", "-_id"}),
		)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
				"id":  id,
			}).Error("nftitem.FindAll failed")
			return err
		}

		for _, item := range items {
			score, err := im.getNftitemScore(ctx, item, supply, attrs, collectionEntropy)
			if err != nil {
				ctx.WithFields(log.Fields{
					"err": err,
				}).Error("failed to getNftitemScore")
				return err
			}

			itemSlice = append(itemSlice, itemWithScore{
				id:    *item.ToId(),
				score: score,
			})
		}

		if len(items) < int(limit) {
			break
		}
		lastObjectId = items[len(items)-1].ObjectId
	}

	sort.Slice(itemSlice, func(i, j int) bool {
		return itemSlice[i].score > itemSlice[j].score
	})

	for rank, item := range itemSlice {
		err := im.nftitem.Patch(ctx, item.id, nftitem.PatchableNftItem{
			OpenrarityRank:  &rank,
			OpenrarityScore: &item.score,
		})
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
			}).Error("failed to nftitem.Patch")
			return err
		}
	}

	err = im.collection.Update(ctx, id, collection.UpdatePayload{
		AttributesHash: attrsHash,
	})
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to collection.Update")
		return err
	}

	return nil
}

func (im *impl) RefreshStat(c ctx.Ctx, id collection.CollectionId) error {
	supply := int64(0)
	attrs := map[string]map[string]int64{}
	var fpItems []*order.OrderItem
	traitsColl := map[string]map[string][]domain.TokenId{}
	traitsFp := map[string]map[string]float64{}
	floorPriceColl := map[domain.TokenId]float64{}
	// objectID with current timestamp
	lastObjectId := primitive.NewObjectID()
	limit := int32(500)
	owners := map[domain.Address]bool{}
	hasFloorPrice := false
	floorPriceInNative := float64(0)
	floorPriceInUsd := float64(0)

	col, err := im.FindOne(c, id)
	if err != nil {
		c.WithField("err", err).Error("FindOne failed")
		return err
	}

	// break once found all items
	for {
		items, err := im.nftitem.FindAll(
			c,
			nftitem.WithChainId(id.ChainId),
			nftitem.WithContractAddresses([]domain.Address{id.Address}),
			nftitem.WithPagination(0, limit),
			nftitem.WithObjectIdLT(lastObjectId),
			nftitem.WithIndexerStates(nftitem.ReadyToServeIndexerStates),
			nftitem.WithSorts([]string{"chainId", "contractAddress", "indexerState", "-_id"}),
		)
		if err != nil {
			c.WithFields(log.Fields{
				"err": err,
				"id":  id,
			}).Error("nftitem.FindAll failed")
			return err
		}

		supply += int64(len(items))

		for _, item := range items {
			for _, attr := range item.Attributes {
				// If an attribute has display type, then it it more like stats.
				// No need to use such attribue as filter
				if len(attr.DisplayType) > 0 {
					continue
				}
				if attrs[attr.TraitType] == nil {
					attrs[attr.TraitType] = map[string]int64{}
				}

				attrs[attr.TraitType][attr.Value] += 1
				if traitsColl[attr.TraitType] == nil {
					traitsColl[attr.TraitType] = map[string][]domain.TokenId{}
				}
				traitsColl[attr.TraitType][attr.Value] = append(traitsColl[attr.TraitType][attr.Value], item.TokenId)
			}
			owners[item.Owner] = true
		}

		if len(items) < int(limit) {
			break
		}
		lastObjectId = items[len(items)-1].ObjectId
	}

	now := time.Now()
	offset := int32(0)

	// break once found all orderItems
	for {
		orderItems, err := im.orderItemRepo.FindAll(c,
			order.WithChainId(id.ChainId),
			order.WithContractAddress(id.Address),
			order.WithStartTimeLT(now),
			order.WithEndTimeGT(now),
			order.WithIsAsk(true),
			order.WithStrategy(order.StrategyFixedPrice),
			order.WithIsValid(true),
			order.WithIsUsed(false),
			order.WithSort("priceInUsd"),
			order.WithPagination(offset, limit),
		)
		if err != nil {
			c.WithFields(log.Fields{
				"err": err,
			}).Error("failed to orderItemRepo.FindAll")
			return err
		}
		offset += limit
		if len(orderItems) > 0 {
			fpItems = append(fpItems, orderItems...)
		}
		if len(orderItems) < int(limit) {
			break
		}
	}

	if len(fpItems) > 0 {
		hasFloorPrice = true
		floorPriceInUsd = fpItems[0].PriceInUsd
		floorPriceInNative = fpItems[0].PriceInNative

		for _, fpItem := range fpItems {
			// initial value
			if floorPriceColl[fpItem.TokenId] == 0 {
				floorPriceColl[fpItem.TokenId] = fpItem.PriceInNative
			}
			if fpItem.PriceInNative < floorPriceColl[fpItem.TokenId] {
				floorPriceColl[fpItem.TokenId] = fpItem.PriceInNative
			}
		}

		for traitType, valueM := range traitsColl {
			for traitValue, tokenIDs := range valueM {
				var floorPrice float64
				for _, tokenID := range tokenIDs {
					// initial value
					if floorPriceColl[tokenID] > 0 && floorPrice == 0 {
						floorPrice = floorPriceColl[tokenID]
					}
					if floorPriceColl[tokenID] > 0 && floorPriceColl[tokenID] < floorPrice {
						floorPrice = floorPriceColl[tokenID]
					}
				}
				if floorPrice > 0 {
					if traitsFp[traitType] == nil {
						traitsFp[traitType] = map[string]float64{}
					}
					traitsFp[traitType][traitValue] = floorPrice
				}
			}
		}
	}

	numOwners := int64(len(owners))
	if col.TokenType == domain.TokenType1155 {
		n, err := im.erc1155holding.CountUniqueOwner(c, col.ChainId, col.Erc721Address)
		if err != nil {
			return err
		}
		numOwners = n
	}

	yesterday := time.Now().Add(-day).Truncate(day)
	fpId := collection.FloorPriceId{ChainId: id.ChainId, Address: id.Address.ToLower(), Date: yesterday}
	fp, err := im.floorPriceHistory.FindOne(c, fpId)
	var previousCollectionFloorInUsd float64 = 0
	var previousCollectionNumOwners int64 = 0
	var previousCollectionOpenseFloorInUsd float64 = 0
	if err == nil {
		previousCollectionFloorInUsd = fp.PriceInUsd
		previousCollectionNumOwners = fp.NumOwners
		previousCollectionOpenseFloorInUsd = fp.OpenseaPriceInUsd
	} else if err != nil && err != domain.ErrNotFound {
		c.WithFields(log.Fields{
			"err": err,
			"id":  fpId,
		}).Error("failed to FindOne")
		return err
	}

	floorPriceMovement := float64(0)
	if previousCollectionFloorInUsd != 0 {
		floorPriceMovement = (floorPriceInUsd - previousCollectionFloorInUsd) / previousCollectionFloorInUsd
	}

	openseaFloorPriceMovement := float64(0)
	if previousCollectionOpenseFloorInUsd != 0 {
		openseaFloorPriceMovement = (col.OpenseaFloorPriceInUsd - previousCollectionOpenseFloorInUsd) / previousCollectionOpenseFloorInUsd
	}

	numOwnersMovement := float64(0)
	if previousCollectionNumOwners != 0 {
		numOwnersMovement = (float64(numOwners) - float64(previousCollectionNumOwners)) / float64(previousCollectionNumOwners)
	}

	updater := collection.UpdatePayload{
		Supply:                    supply,
		Attributes:                attrs,
		NumOwners:                 numOwners,
		HasFloorPrice:             &hasFloorPrice,
		FloorPriceInNative:        &floorPriceInNative,
		FloorPriceInUsd:           &floorPriceInUsd,
		FloorPriceMovement:        &floorPriceMovement,
		OpenseaFloorPriceMovement: &openseaFloorPriceMovement,
		NumOwnersMovement:         &numOwnersMovement,
		TraitFloorPrice:           traitsFp,
	}

	if err := im.collection.Update(c, id, updater); err != nil {
		c.WithFields(log.Fields{
			"id":                 id,
			"supply":             supply,
			"attrs":              attrs,
			"numOwners":          len(owners),
			"floorPriceInNative": floorPriceInNative,
			"err":                err,
		}).Error("collection.Update failed")
		return err
	}

	floorPriceHistory := collection.FloorPriceHistory{
		ChainId:              id.ChainId,
		Address:              id.Address,
		Date:                 time.Now().Truncate(day),
		PriceInNative:        floorPriceInNative,
		PriceInUsd:           floorPriceInUsd,
		NumOwners:            numOwners,
		OpenseaPriceInNative: col.OpenseaFloorPriceInNative,
		OpenseaPriceInUsd:    col.OpenseaFloorPriceInUsd,
	}

	if err := im.floorPriceHistory.Upsert(c, floorPriceHistory); err != nil {
		c.WithFields(log.Fields{
			"floorPriceHistory": floorPriceHistory,
			"err":               err,
		}).Error("floorPriceHistory.Upsert failed")
		return err
	}

	// update collection state first, then calcuate the rarity score
	if col.ShouldCalculateOpenrarity {
		err = im.calculateOpenrarityScoreAndRank(c, id)
		if err != nil {
			c.WithFields(log.Fields{
				"err": err,
			}).Error("failed to calculateOpenrarityScoreAndRank")
		}
	}

	return nil
}

func (im *impl) GetTopCollections(c ctx.Ctx, periodType collection.PeriodType, opts ...domain.OpenseaDataFindAllOptions) ([]collection.CollectionWithTradingVolume, error) {
	if periodType == collection.PeriodTypeUnknown {
		return nil, domain.ErrBadParamInput
	}

	now := time.Now()
	promotedColsSet := map[collection.CollectionId]interface{}{}
	_, promotedCols, err := im.promotedCollectionsUC.GetPromotedCollections(c, &now)
	if err == nil {
		for _, col := range promotedCols {
			promotedColsSet[collection.CollectionId{ChainId: col.ChainId, Address: col.Address}] = struct{}{}
		}
	} else if err == domain.ErrNotFound {
		// fallthrough
	} else {
		return nil, err
	}
	ethChainId := domain.ChainId(1)
	ethPrice, err := im.chainlink.GetLatestAnswer(c, ethChainId, zeroAddr)
	if err != nil {
		c.WithField("err", err).Error("chainlink.GetLatestAnswer failed")
		return nil, err
	}
	apePrice, err := im.chainlink.GetLatestAnswer(c, ethChainId, apeAddr)
	if err != nil {
		c.WithField("err", err).Error("chainlink.GetLatestAnswer failed")
		return nil, err
	}

	var sort string
	switch periodType {
	case collection.PeriodTypeOneHour:
		sort = "oneHourVolume"
	case collection.PeriodTypeSixHour:
		sort = "sixHourVolume"
	case collection.PeriodTypeDay:
		sort = "oneDayVolume"
	case collection.PeriodTypeWeek:
		sort = "sevenDayVolume"
	case collection.PeriodTypeMonth:
		sort = "thirtyDayVolume"
	case collection.PeriodTypeAll:
		sort = "totalVolume"
	}

	findAllOpts := append([]domain.OpenseaDataFindAllOptions{
		domain.OpenseaDataWithChainId(ethChainId),
		domain.OpenseaDataWithSort(sort, domain.SortDirDesc),
	}, opts...)

	tvs, err := im.openseaData.FindAll(c,
		findAllOpts...,
	)
	if err != nil {
		return nil, err
	}
	topCollections := make([]collection.CollectionWithTradingVolume, len(tvs))
	for i, tv := range tvs {
		id := collection.CollectionId{ChainId: tv.ChainId, Address: tv.Address}
		col, err := im.collection.FindOne(c, id)
		if err != nil {
			c.WithFields(log.Fields{
				"id":  id,
				"err": err,
			}).Error("collection.FindOne failed")
			return nil, err
		}

		_, promo := promotedColsSet[id]
		colWithTv := collection.CollectionWithTradingVolume{
			ChainId:                   tv.ChainId,
			Erc721Address:             tv.Address,
			CollectionName:            col.CollectionName,
			LogoImageHash:             col.LogoImageHash,
			LogoImageUrl:              col.LogoImageUrl,
			OpenseaFloorPriceInNative: col.OpenseaFloorPriceInNative,
			OpenseaFloorPriceInUsd:    col.OpenseaFloorPriceInUsd,
			OpenseaFloorPriceInApe:    decimal.NewFromFloat(col.OpenseaFloorPriceInUsd).Div(apePrice).InexactFloat64(),
			OpenseaFloorPriceMovement: col.OpenseaFloorPriceMovement,
			NumOwners:                 col.NumOwners,
			Supply:                    col.Supply,
			EligibleForPromo:          promo,
		}
		switch periodType {
		case collection.PeriodTypeOneHour:
			colWithTv.Sales = tv.OneHourSales
			colWithTv.Volume = tv.OneHourVolume
			colWithTv.ChangeRatio = tv.OneHourChange
		case collection.PeriodTypeSixHour:
			colWithTv.Sales = tv.SixHourSales
			colWithTv.Volume = tv.SixHourVolume
			colWithTv.ChangeRatio = tv.SixHourChange
		case collection.PeriodTypeDay:
			colWithTv.Sales = tv.OneDaySales
			colWithTv.Volume = tv.OneDayVolume
			colWithTv.ChangeRatio = tv.OneDayChange
		case collection.PeriodTypeWeek:
			colWithTv.Sales = tv.SevenDaySales
			colWithTv.Volume = tv.SevenDayVolume
			colWithTv.ChangeRatio = tv.SevenDayChange
		case collection.PeriodTypeMonth:
			colWithTv.Sales = tv.ThirtyDaySales
			colWithTv.Volume = tv.ThirtyDayVolume
			colWithTv.ChangeRatio = tv.ThirtyDayChange
		case collection.PeriodTypeAll:
			colWithTv.Sales = tv.TotalSales
			colWithTv.Volume = tv.TotalVolume
			colWithTv.ChangeRatio = 0
		}
		colWithTv.VolumeInUsd = ethPrice.Mul(decimal.NewFromFloat(colWithTv.Volume)).InexactFloat64()
		colWithTv.VolumeInApe = decimal.NewFromFloat(colWithTv.VolumeInUsd).Div(apePrice).InexactFloat64()
		topCollections[i] = colWithTv
	}
	return topCollections, nil
}

func (im *impl) GetViewCount(c ctx.Ctx, id collection.CollectionId) (int32, error) {
	if _, err := im.FindOne(c, id); err != nil {
		if err != domain.ErrNotFound {
			c.WithField("id", id).WithField("err", err).Error("FindOne failed")
		}
		return 0, err
	}

	return im.collection.IncreaseViewCount(c, id, 1)
}

func (im *impl) UpdateSaleStat(c ctx.Ctx, id collection.CollectionId, priceInNative, priceInUsd float64, blkTime time.Time) error {
	col, err := im.collection.FindOne(c, id)
	if err == domain.ErrNotFound {
		c.WithField("id", id).Warn("collection not found")
		return nil
	} else if err != nil {
		c.WithField("err", err).Error("collection.FindOne failed")
		return err
	}
	payload := collection.UpdatePayload{
		HasBeenSold: true,
		LastSoldAt:  blkTime,
	}
	if priceInNative >= col.HighestSale {
		payload.HighestSale = priceInNative
		payload.HighestSaleInUsd = priceInUsd
	}
	if err := im.collection.Update(c, id, payload); err != nil {
		c.WithField("err", err).Error("collection.Update failed")
		return err
	}
	return nil
}

func (im *impl) UpdateLastListedAt(c ctx.Ctx, id collection.CollectionId, blkTime time.Time) error {
	if _, err := im.collection.FindOne(c, id); err == nil {
		patchable := collection.UpdatePayload{LastListedAt: blkTime, HasBeenListed: true}
		if err := im.collection.Update(c, id, patchable); err != nil {
			c.WithFields(log.Fields{
				"id":            id,
				"updatePayload": patchable,
				"err":           err,
			}).Error("collection.Update failed")
			return err
		}
	}
	return nil
}

func (im *impl) UpdateInfo(c ctx.Ctx, id collection.CollectionId, info collection.UpdateInfoPayload) error {
	patchable := collection.UpdatePayload{}

	if info.Email != nil {
		patchable.Email = *info.Email
	}

	if info.ColectionName != nil {
		patchable.CollectionName = *info.ColectionName
	}

	if info.Description != nil {
		patchable.Description = *info.Description
	}

	if info.Categories != nil {
		patchable.Categories = info.Categories
	}

	if info.LogoImageHash != nil {
		patchable.LogoImageHash = *info.LogoImageHash
	}

	if info.LogoImageUrl != nil {
		patchable.LogoImageUrl = *info.LogoImageUrl
	}

	if info.SiteUrl != nil {
		patchable.SiteUrl = *info.SiteUrl
	}

	if info.Discord != nil {
		patchable.Discord = *info.Discord
	}

	if info.TwitterHandle != nil {
		patchable.TwitterHandle = *info.TwitterHandle
	}

	if info.InstagramHandle != nil {
		patchable.InstagramHandle = *info.InstagramHandle
	}

	if info.MediumHandle != nil {
		patchable.MediumHandle = *info.MediumHandle
	}

	if info.Telegram != nil {
		patchable.Telegram = *info.Telegram
	}

	if err := im.collection.Update(c, id, patchable); err != nil {
		c.WithFields(log.Fields{
			"id":            id,
			"updatePayload": patchable,
			"err":           err,
		}).Error("collection.Update failed")
		return err
	}
	return nil
}

func (im *impl) UpdateLastOpenseaEventIndexAt(c ctx.Ctx, id collection.CollectionId, t time.Time) error {
	if _, err := im.collection.FindOne(c, id); err == nil {
		patchable := collection.UpdatePayload{LastOpenseaEventIndexAt: t}
		if err := im.collection.Update(c, id, patchable); err != nil {
			c.WithFields(log.Fields{
				"id":            id,
				"updatePayload": patchable,
				"err":           err,
			}).Error("collection.Update failed")
			return err
		}
	}
	return nil
}

func (im *impl) UpdateTraitFloorPrice(c ctx.Ctx, id collection.CollectionId, traitName, traitValue string, price float64) error {
	col, err := im.collection.FindOne(c, id)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to FindOne")
		return err
	}

	traitFloorPrice := col.TraitFloorPrice
	if traitFloorPrice == nil {
		traitFloorPrice = map[string]map[string]float64{}
	}
	if _, ok := traitFloorPrice[traitName]; !ok {
		traitFloorPrice[traitName] = map[string]float64{
			traitValue: price,
		}
	} else {
		traitFloorPrice[traitName][traitValue] = price
	}

	value := collection.UpdatePayload{
		TraitFloorPrice: traitFloorPrice,
	}

	if err := im.collection.Update(c, id, value); err != nil {
		c.WithFields(log.Fields{
			"err":   err,
			"id":    id,
			"value": value,
		}).Error("failed to Update")
		return err
	}

	return nil
}

func (im *impl) UpdateOpenseaFloorPrice(c ctx.Ctx, id collection.CollectionId, price float64) error {
	ethChainId := domain.ChainId(1)
	ethPrice, err := im.chainlink.GetLatestAnswer(c, ethChainId, zeroAddr)
	if err != nil {
		c.WithField("err", err).Error("chainlink.GetLatestAnswer failed")
		return err
	}

	value := collection.UpdatePayload{
		OpenseaFloorPriceInNative: &price,
		OpenseaFloorPriceInUsd:    ptr.Float64(price * ethPrice.InexactFloat64()),
	}

	if err := im.collection.Update(c, id, value); err != nil {
		c.WithFields(log.Fields{
			"err":   err,
			"id":    id,
			"value": value,
		}).Error("failed to Update")
		return err
	}

	return nil
}

func (im *impl) GetCollectionStatByAccount(c ctx.Ctx, id collection.CollectionId, account domain.Address) (*collection.CollectionWithStatByAccount, error) {
	ethChainId := domain.ChainId(1)
	ethPrice, err := im.chainlink.GetLatestAnswer(c, ethChainId, zeroAddr)
	if err != nil {
		c.WithField("err", err).Error("chainlink.GetLatestAnswer failed")
		return nil, err
	}

	col, err := im.collection.FindOne(c, id)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to collection.FindOne")
		return nil, err
	}

	holdings, err := im.erc1155holding.FindAll(
		c,
		erc1155.WithHoldingAddress(col.Erc721Address),
		erc1155.WithOwner(account.ToLower()),
	)
	if err != nil {
		c.WithFields(log.Fields{
			"err":     err,
			"address": col.Erc721Address,
			"account": account,
		}).Error("failed to erc1155holding.FindAll")
		return nil, err
	}

	holdingBalanceMap := map[string]int{}
	nftitemIds := []nftitem.Id{}
	for _, holding := range holdings {
		nftid := nftitem.Id{
			ChainId:         holding.ChainId,
			ContractAddress: holding.Address.ToLower(),
			TokenId:         holding.TokenId,
		}
		key := nftid.ToString()
		holdingBalanceMap[key] = int(holding.Balance)
		nftitemIds = append(nftitemIds, nftid)
	}

	salesVolume := float64(0)
	salesVolumeChange := float64(0)
	osData, err := im.openseaData.FindOne(
		c,
		domain.OpenseaDataId{
			ChainId: col.ChainId,
			Address: col.Erc721Address.ToLower(),
		},
	)
	if err != nil && err != domain.ErrNotFound {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to openseaData.FindOne")
		return nil, err
	}

	// opensea data may not found if collection is not on eth
	if osData != nil {
		salesVolume = osData.OneDayVolume * ethPrice.InexactFloat64()
		salesVolumeChange = osData.OneDayChange
	}

	items, err := im.nftitem.FindAll(
		c,
		nftitem.WithChainId(id.ChainId),
		nftitem.WithContractAddresses([]domain.Address{id.Address.ToLower()}),
		nftitem.WithOwner(account.ToLower()),
		nftitem.WithHoldingIds(nftitemIds),
	)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to nftitem.FindAll")
		return nil, err
	}

	totalCount := 0
	for _, item := range items {
		if item.TokenType == 1155 {
			key := item.ToId().ToString()
			totalCount += holdingBalanceMap[key]
		} else {
			totalCount += 1
		}
	}

	instantLiquidityInUsd := float64(0)
	for _, item := range items {
		instantLiquidityInUsd += item.InstantLiquidityInUsd
	}

	totalValue := col.OpenseaFloorPriceInNative * ethPrice.InexactFloat64() * float64(totalCount)
	instantLiquidityRatio := float64(0)
	if totalValue != 0 {
		instantLiquidityRatio = instantLiquidityInUsd / totalValue
	}

	return &collection.CollectionWithStatByAccount{
		Collection:               *col,
		OwnedNftCount:            totalCount,
		TotalValue:               totalValue,
		InstantLiquidityInUsd:    instantLiquidityInUsd,
		InstantLiquidityRatio:    instantLiquidityRatio,
		OpenseaSalesVolume:       salesVolume,
		OpenseaSalesVolumeChange: salesVolumeChange,
	}, nil
}

func (im *impl) GetActivities(c ctx.Ctx, id collection.CollectionId, optFns ...account.FindActivityHistoryOptions) (*collection.ActivityResult, error) {
	activityOpts := append(
		[]account.FindActivityHistoryOptions{
			account.ActivityHistoryWithCollection(id.ChainId, id.Address),
		},
		optFns...,
	)
	items, err := im.activityHistoryRepo.FindActivities(c, activityOpts...)
	if err != nil {
		c.WithFields(log.Fields{
			"err":  err,
			"opts": activityOpts,
		}).Error("FindActivities failed")
		return nil, err
	}

	count, err := im.activityHistoryRepo.CountActivities(c, activityOpts...)
	if err != nil {
		c.WithFields(log.Fields{
			"err":  err,
			"opts": activityOpts,
		}).Error("CountActivities failed")
		return nil, err
	}
	return &collection.ActivityResult{Items: items, Count: count}, nil
}

func (im *impl) GetGlobalOfferStats(c ctx.Ctx, id collection.CollectionId) (*collection.GlobalOfferStatResult, error) {
	now := time.Now()
	collectionOffers, err := im.orderItemRepo.FindAll(c,
		// assuming we support only weth
		order.WithChainId(id.ChainId),
		order.WithContractAddress(id.Address),
		order.WithIsValid(true),
		order.WithIsUsed(false),
		order.WithStartTimeLT(now),
		order.WithEndTimeGT(now),
		order.WithStrategy(order.StrategyCollectionOffer),
		order.WithSort("-priceInNative"),
	)
	if err != nil {
		c.WithFields(log.Fields{
			"id":  id,
			"err": err,
		}).Error("orderItemRepo.FindAll failed")
		return nil, err
	}

	var (
		rows         []collection.GlobalOfferStatRow
		currentPrice string
		priceInUsd   float64
		bidders      map[string]interface{}
		size         int
	)
	rows = make([]collection.GlobalOfferStatRow, 0)
	bidders = make(map[string]interface{})
	for _, o := range collectionOffers {
		if o.DisplayPrice != currentPrice {
			if len(currentPrice) > 0 {
				row := collection.GlobalOfferStatRow{
					DisplayPrice: currentPrice,
					PriceInUsd:   priceInUsd,
					Size:         size,
					Sum:          decimal.RequireFromString(currentPrice).Mul(decimal.NewFromInt(int64(size))).String(),
					Bidders:      len(bidders),
				}
				rows = append(rows, row)
			}
			currentPrice = o.DisplayPrice
			priceInUsd = o.PriceInUsd
			size = 0
			bidders = make(map[string]interface{})
		}
		size++
		bidders[o.Signer.ToLowerStr()] = struct{}{}
	}
	if len(currentPrice) > 0 {
		row := collection.GlobalOfferStatRow{
			DisplayPrice: currentPrice,
			PriceInUsd:   priceInUsd,
			Size:         size,
			Sum:          decimal.RequireFromString(currentPrice).Mul(decimal.NewFromInt(int64(size))).String(),
			Bidders:      len(bidders),
		}
		rows = append(rows, row)
	}

	return &collection.GlobalOfferStatResult{
		Rows: rows,
	}, nil
}
