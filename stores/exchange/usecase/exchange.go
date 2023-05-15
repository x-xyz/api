package usecase

import (
	"math/big"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	pricefomatter "github.com/x-xyz/goapi/base/price_fomatter"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/exchange"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/domain/token"
)

type ExchangeUseCaseCfg struct {
	OrderUseCase      order.UseCase
	OrderNonceUseCase account.OrderNonceUseCase

	Nftitem         nftitem.Repo
	Token           token.Usecase
	ActivityHistory account.ActivityHistoryRepo
	Collection      collection.Usecase
	TradingVolume   collection.TradingVolumeUseCase
	PriceFormatter  pricefomatter.PriceFormatter
}

type ExchangeUseCase struct {
	OrderUseCase      order.UseCase
	OrderNonceUseCase account.OrderNonceUseCase

	Nftitem         nftitem.Repo
	Token           token.Usecase
	ActivityHistory account.ActivityHistoryRepo
	Collection      collection.Usecase
	TradingVolume   collection.TradingVolumeUseCase
	PriceFormatter  pricefomatter.PriceFormatter
}

func NewExchangeUseCase(cfg *ExchangeUseCaseCfg) exchange.UseCase {
	return &ExchangeUseCase{
		OrderUseCase:      cfg.OrderUseCase,
		OrderNonceUseCase: cfg.OrderNonceUseCase,
		Nftitem:           cfg.Nftitem,
		Token:             cfg.Token,
		ActivityHistory:   cfg.ActivityHistory,
		Collection:        cfg.Collection,
		TradingVolume:     cfg.TradingVolume,
		PriceFormatter:    cfg.PriceFormatter,
	}
}

func (u *ExchangeUseCase) CancelAllOrders(ctx bCtx.Ctx, chainId domain.ChainId, event *exchange.CancelAllOrdersEvent, lMeta *domain.LogMeta) error {
	id := account.OrderNonceId{
		Address: event.User,
		ChainId: chainId,
	}
	if err := u.OrderNonceUseCase.UpdateMinValidOrderNonce(ctx, id, event.NewNonce.String()); err != nil {
		return err
	}

	usedNonce := new(big.Int).Sub(event.NewNonce, big.NewInt(1))
	if err := u.OrderNonceUseCase.UpdateAvailableNonceIfNeeded(ctx, id, usedNonce.String()); err != nil {
		return err
	}

	if err := u.OrderUseCase.CancelOrderItemByNonce(ctx, chainId, event.User, event.NewNonce, lMeta); err != nil {
		return err
	}
	return nil
}

func (u *ExchangeUseCase) CancelMultipleOrders(ctx bCtx.Ctx, chainId domain.ChainId, event *exchange.CancelMultipleOrdersEvent, lMeta *domain.LogMeta) error {
	for _, orderItemHash := range event.OrderItemHashes {
		if err := u.OrderUseCase.CancelOrderItemByOrderItemHash(ctx, chainId, orderItemHash, true, lMeta); err != nil {
			return err
		}
	}
	return nil
}

type saleType string

type SaleInfo struct {
	From          domain.Address
	To            domain.Address
	Type          saleType
	Strategy      domain.Address
	OrderHash     domain.OrderHash
	ItemIdx       *big.Int
	OrderItemHash domain.OrderHash
	Fulfillment   exchange.Fulfillment
}

func (u *ExchangeUseCase) TakerAsk(ctx bCtx.Ctx, chainId domain.ChainId, event *exchange.TakerAskEvent, lMeta *domain.LogMeta) error {
	sale := SaleInfo{
		From:          event.Taker,
		To:            event.Maker,
		Type:          "takerAsk",
		Strategy:      event.Strategy,
		OrderHash:     event.OrderHash,
		ItemIdx:       event.ItemIdx,
		OrderItemHash: event.OrderItemHash,
		Fulfillment:   event.Fulfillment,
	}
	return u.sale(ctx, chainId, &sale, lMeta)
}

func (u *ExchangeUseCase) TakerBid(ctx bCtx.Ctx, chainId domain.ChainId, event *exchange.TakerBidEvent, lMeta *domain.LogMeta) error {
	sale := SaleInfo{
		From:          event.Maker,
		To:            event.Taker,
		Type:          "takerBid",
		Strategy:      event.Strategy,
		OrderHash:     event.OrderHash,
		ItemIdx:       event.ItemIdx,
		OrderItemHash: event.OrderItemHash,
		Fulfillment:   event.Fulfillment,
	}
	return u.sale(ctx, chainId, &sale, lMeta)
}

func (u *ExchangeUseCase) sale(ctx bCtx.Ctx, chainId domain.ChainId, sale *SaleInfo, lMeta *domain.LogMeta) error {
	if err := u.OrderUseCase.CancelOrderItemByOrderItemHash(ctx, chainId, sale.OrderItemHash, false, lMeta); err != nil {
		return err
	}

	ctx = bCtx.WithValues(ctx, map[string]interface{}{"chainId": chainId, "sale": sale, "lMeta": lMeta})
	id := nftitem.Id{
		ChainId:         chainId,
		ContractAddress: sale.Fulfillment.Collection,
		TokenId:         domain.TokenId(sale.Fulfillment.TokenId.String()),
	}
	if _, err := u.Token.EnsureNftExists(ctx, id); err != nil {
		ctx.WithFields(log.Fields{
			"id":  id,
			"err": err,
		}).Error("ensureNftExists failed")
		return err
	}

	displayPrice, priceInUsd, priceInNative, err := u.PriceFormatter.GetPrices(ctx, chainId, sale.Fulfillment.Currency, sale.Fulfillment.Price)
	if err != nil {
		ctx.WithField("err", err).Error("failed to parse price")
		return err
	}

	displayPricePerItem, pricePerItemInUsd, pricePerItemInNative, err := u.PriceFormatter.GetPrices(ctx, chainId, sale.Fulfillment.Currency, new(big.Int).Div(sale.Fulfillment.Price, sale.Fulfillment.Amount))
	if err != nil {
		ctx.WithField("err", err).Error("failed to parse price")
		return err
	}

	// refresh listing and offer state?
	if err := u.Token.RefreshListingAndOfferState(ctx, id); err != nil {
		ctx.Error("token.RefreshListingAndOfferState failed")
		return err
	}
	/*
		if err := u.Token.RemoveListing(ctx, id, &sale.Seller); err != nil {
			ctx.Error("token.RemoveListing failed")
			return err
		}
	*/

	history := account.ActivityHistory{
		ChainId:         chainId,
		ContractAddress: sale.Fulfillment.Collection,
		TokenId:         domain.TokenId(sale.Fulfillment.TokenId.String()),
		Type:            account.ActivityHistoryTypeSale,
		Account:         sale.From,
		To:              sale.To,
		Quantity:        sale.Fulfillment.Amount.String(),
		Price:           displayPrice.String(),
		PaymentToken:    sale.Fulfillment.Currency,
		PriceInUsd:      priceInUsd,
		PriceInNative:   priceInNative,
		BlockNumber:     lMeta.BlockNumber,
		TxHash:          lMeta.TxHash,
		Time:            lMeta.BlockTime,
		Source:          account.SourceX,
	}

	if err := u.ActivityHistory.Insert(ctx, &history); err != nil {
		ctx.WithFields(log.Fields{
			"activityHistory": history,
			"err":             err,
		}).Error("activityHistory.Insert failed")
		return err
	}

	cId := collection.CollectionId{ChainId: chainId, Address: sale.Fulfillment.Collection}
	if err := u.Collection.UpdateSaleStat(ctx, cId, pricePerItemInNative, pricePerItemInUsd, lMeta.BlockTime); err != nil {
		ctx.WithFields(log.Fields{
			"id":                   cId,
			"chainId":              chainId,
			"nft":                  sale.Fulfillment.Collection,
			"pricePerItemInNative": pricePerItemInNative,
			"err":                  err,
		}).Error("collection.UpdateSaleStat failed")
		return err
	}

	if _, err := u.TradingVolume.IncDailyVolume(ctx, chainId, sale.Fulfillment.Collection, lMeta.BlockTime, priceInNative); err != nil {
		ctx.WithFields(log.Fields{
			"chainId": chainId,
			"nft":     sale.Fulfillment.Collection,
			"time":    lMeta.BlockTime,
			"volume":  priceInNative,
			"err":     err,
		}).Error("tradingVolume.IncDailyVolume failed")
		return err
	}

	if _, err := u.TradingVolume.IncTotalVolume(ctx, chainId, sale.Fulfillment.Collection, priceInNative); err != nil {
		ctx.WithFields(log.Fields{
			"chainId": chainId,
			"nft":     sale.Fulfillment.Collection,
			"volume":  priceInNative,
			"err":     err,
		}).Error("tradingVolume.IncTotalVolume failed")
		return err
	}

	item := &nftitem.PatchableNftItem{
		LastSalePrice:             ptr.Float64(displayPricePerItem.InexactFloat64()),
		LastSalePricePaymentToken: ptr.String(sale.Fulfillment.Currency.ToLowerStr()),
		LastSalePriceInUsd:        ptr.Float64(pricePerItemInUsd),
		LastSalePriceInNative:     ptr.Float64(pricePerItemInNative),
		SoldAt:                    &lMeta.BlockTime,
	}
	if err := u.Token.PatchNft(ctx, &id, item); err != nil {
		ctx.WithFields(log.Fields{
			"item": item,
			"err":  err,
		}).Error("token.PatchNft failed")
		return err
	}
	return nil
}
