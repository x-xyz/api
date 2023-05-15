package usecase

import (
	"strconv"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/service/opensea"
)

type activityUsecase struct {
	activityHistoryRepo account.ActivityHistoryRepo
}

func NewActivityHistoryUsecase(activityHistoryRepo account.ActivityHistoryRepo) account.ActivityHistoryUseCase {
	return &activityUsecase{
		activityHistoryRepo: activityHistoryRepo,
	}
}

func (im *activityUsecase) Insert(ctx ctx.Ctx, ac *account.ActivityHistory) error {
	return im.activityHistoryRepo.Insert(ctx, ac)
}

func (im *activityUsecase) ParseAndInsertOpenseaEventToActivityHistory(ctx ctx.Ctx, ev opensea.AssetEvent) error {
	switch ev.EventType {
	case opensea.EventTypeSuccessful:
		return im.parseSuccessfulEvent(ctx, ev)
	}
	return nil
}

func (im *activityUsecase) parseSuccessfulEvent(ctx ctx.Ctx, ev opensea.AssetEvent) error {
	// TODO: handle asset bundle
	if ev.Asset == nil {
		return nil
	}

	if ev.Transaction.BlockHash == nil || ev.Transaction.BlockNumber == nil {
		// ignore failed transaction event
		return nil
	}

	id := strconv.FormatInt(ev.Id, 10)
	blockNumber, err := strconv.ParseUint(*ev.Transaction.BlockNumber, 10, 64)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":         err,
			"blockNumber": ev.Transaction.BlockNumber,
		}).Error("failed to ParseUint")
		return err
	}

	// only track eth on opensea
	ethereumChainId := domain.ChainId(1)

	displayPrice, priceInUsd, priceInNative, err := ev.GetPrices()
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"event": ev,
		}).Error("failed to priceFormatter.GetPrices")
	}

	t, err := time.Parse("2006-01-02T15:04:05", ev.CreatedDate)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":  err,
			"date": ev.CreatedDate,
		}).Error("failed to time.Parse")
		return err
	}

	buyAc := account.ActivityHistory{
		ChainId:         ethereumChainId,
		ContractAddress: ev.Asset.AssetContract.Address.ToLower(),
		TokenId:         ev.Asset.TokenId,
		Type:            account.ActivityHistoryTypeBuy,
		Account:         ev.WinnerAccount.Address.ToLower(),
		Quantity:        ev.Quantity,
		Price:           displayPrice.String(),
		PaymentToken:    ev.PaymentToken.Address.ToLower(),
		PriceInUsd:      priceInUsd,
		PriceInNative:   priceInNative,
		BlockNumber:     domain.BlockNumber(blockNumber),
		TxHash:          ev.Transaction.TransactionHash,
		Time:            t,
		Source:          account.SourceOpensea,
		SourceEventId:   id,
	}

	err = im.activityHistoryRepo.UpsertBySourceEventId(ctx, buyAc.Source, buyAc.SourceEventId, buyAc.Type, &buyAc)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"activity": buyAc,
		}).Error("failed to UpsertBySourceEventId")
		return err
	}

	soldAc := account.ActivityHistory{
		ChainId:         ethereumChainId,
		ContractAddress: ev.Asset.AssetContract.Address.ToLower(),
		TokenId:         ev.Asset.TokenId,
		Type:            account.ActivityHistoryTypeSold,
		Account:         ev.Seller.Address.ToLower(),
		Quantity:        ev.Quantity,
		Price:           displayPrice.String(),
		PaymentToken:    ev.PaymentToken.Address.ToLower(),
		PriceInUsd:      priceInUsd,
		PriceInNative:   priceInNative,
		BlockNumber:     domain.BlockNumber(blockNumber),
		TxHash:          ev.Transaction.TransactionHash,
		Time:            t,
		Source:          account.SourceOpensea,
		SourceEventId:   id,
	}

	err = im.activityHistoryRepo.UpsertBySourceEventId(ctx, soldAc.Source, soldAc.SourceEventId, soldAc.Type, &soldAc)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"activity": soldAc,
		}).Error("failed to UpsertBySourceEventId")
		return err
	}
	return nil
}
