package usecase

import (
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/airdrop"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/domain/token"
)

type CountType int

const (
	CountTypeMain CountType = iota
	CountTypeSub
)

type ListingRecordUseCaseCfg struct {
	TokenUC           token.Usecase
	CollectionUC      collection.Usecase
	ListingRecordRepo airdrop.ListingRecordRepo
}

type listingRecordUseCaseImpl struct {
	tokenUC           token.Usecase
	collectionUC      collection.Usecase
	listingRecordRepo airdrop.ListingRecordRepo
}

func NewListingRecordUseCaseImpl(cfg *ListingRecordUseCaseCfg) airdrop.ListingRecordUseCase {
	return &listingRecordUseCaseImpl{
		tokenUC:           cfg.TokenUC,
		collectionUC:      cfg.CollectionUC,
		listingRecordRepo: cfg.ListingRecordRepo,
	}
}

func (u *listingRecordUseCaseImpl) SnapshotCollectionListings(ctx bCtx.Ctx, chainId domain.ChainId, address domain.Address, snapshotTime time.Time) error {
	colId := collection.CollectionId{ChainId: chainId, Address: address}
	col, err := u.collectionUC.FindOne(ctx, colId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"colId": colId,
			"err":   err,
		}).Warn("collection.FindOne failed")
		return err
	}

	tokens, err := u.tokenUC.SearchV2(ctx, token.WithChainId(chainId),
		token.WithCollections(address),
		token.WithSaleStatus(nftitem.SaleStatusBuyNow),
		token.WithIncludeOrders(true),
	)
	if err != nil {
		ctx.WithFields(log.Fields{
			"id":  colId,
			"err": err,
		}).Error("token.SearchV2 failed")
		return err
	}

	accountListings := make(map[domain.Address]map[int]int)
	for _, token := range tokens.Items {
		if !u.isEligible(ctx, col, token) {
			continue
		}
		if accountListings[token.Owner] == nil {
			accountListings[token.Owner] = make(map[int]int)
		}
		accountListings[token.Owner][int(CountTypeMain)]++
	}
	now := time.Now()
	for account, count := range accountListings {
		record := airdrop.ListingRecord{
			Owner:           account,
			ChainId:         chainId,
			ContractAddress: address,
			Count:           count[int(CountTypeMain)],
			SubCount:        count[int(CountTypeSub)],
			SnapshotTime:    &snapshotTime,
			UpdatedAt:       &now,
		}
		if err := u.listingRecordRepo.Upsert(ctx, &record); err != nil {
			ctx.WithFields(log.Fields{
				"record": record,
				"err":    err,
			}).Error("listingRecordRepo.Upsert failed")
			return err
		}
	}
	return nil
}

func (u *listingRecordUseCaseImpl) isEligible(ctx bCtx.Ctx, collection *collection.Collection, token *token.TokenWithDetail) bool {
	// no or not public listing
	if token.ActiveListing == nil || token.ActiveListing.Strategy != order.StrategyFixedPrice {
		return false
	}
	if token.ActiveListing.Marketplace != order.MarketplaceApecoin.String() {
		return false
	}

	return token.ActiveListing.PriceInNative >= collection.OpenseaFloorPriceInNative*0.75 && token.ActiveListing.PriceInNative <= collection.OpenseaFloorPriceInNative*1.25
}
