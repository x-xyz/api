package usecase

import (
	"time"

	"github.com/shopspring/decimal"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
)

const zeroAddr = domain.Address("0x0000000000000000000000000000000000000000")
const apeAddr = domain.Address("0x4d224452801aced8b2f0aebe155379bb5d594381")
const day = 24 * time.Hour

type tradingVolumeUseCase struct {
	repo      collection.TradingVolumeRepo
	chainlink domain.ChainlinkUsacase
}

func NewTradingVolumeUseCase(repo collection.TradingVolumeRepo, chainlink domain.ChainlinkUsacase) collection.TradingVolumeUseCase {
	return &tradingVolumeUseCase{repo: repo, chainlink: chainlink}
}

func (u *tradingVolumeUseCase) FindOne(ctx bCtx.Ctx, id collection.TradingVolumeId) (*collection.TradingVolume, error) {
	id.Date = id.Date.Truncate(day)
	return u.repo.FindOne(ctx, id)
}

func (u *tradingVolumeUseCase) Upsert(ctx bCtx.Ctx, v collection.TradingVolume) error {
	v.Date = v.Date.Truncate(day)
	return u.repo.Upsert(ctx, v)
}

func (u *tradingVolumeUseCase) IncDailyVolume(ctx bCtx.Ctx, chainId domain.ChainId, address domain.Address, date time.Time, volume float64) (float64, error) {
	id := collection.TradingVolumeId{
		ChainId: chainId,
		Address: address,
		Date:    date.Truncate(day),
		Period:  collection.PeriodTypeDay,
	}
	v, err := u.repo.IncVolume(ctx, id, volume)
	if err != nil {
		ctx.WithFields(log.Fields{
			"id":     id,
			"volume": volume,
			"err":    err,
		}).Error("repo.IncVolume failed")
		return 0, err
	}
	if err := u.updatePriceInUsd(ctx, id, v); err != nil {
		ctx.WithFields(log.Fields{
			"id":     id,
			"volume": v,
			"err":    err,
		}).Warn("updatePriceInUsd failed")
		// still count as success
	}
	return v, nil
}

func (u *tradingVolumeUseCase) IncTotalVolume(
	ctx bCtx.Ctx,
	chainId domain.ChainId,
	address domain.Address,
	volume float64,
) (float64, error) {
	id := collection.TradingVolumeId{
		ChainId: chainId,
		Address: address,
		Period:  collection.PeriodTypeAll,
	}
	v, err := u.repo.IncVolume(ctx, id, volume)
	if err != nil {
		ctx.WithFields(log.Fields{
			"id":     id,
			"volume": volume,
			"err":    err,
		}).Error("repo.IncVolume failed")
		return 0, err
	}
	if err := u.updatePriceInUsd(ctx, id, v); err != nil {
		ctx.WithFields(log.Fields{
			"id":     id,
			"volume": v,
			"err":    err,
		}).Warn("updatePriceInUsd failed")
		// still count as success
	}
	return v, nil
}

func (u *tradingVolumeUseCase) updatePriceInUsd(ctx bCtx.Ctx, id collection.TradingVolumeId, volume float64) error {
	price, err := u.chainlink.GetLatestAnswer(ctx, id.ChainId, zeroAddr)
	if err != nil {
		ctx.WithFields(log.Fields{
			"chainId": id.ChainId,
			"token":   zeroAddr,
			"err":     err,
		}).Error("chainlink.GetLatestAnswer failed")
		return err
	}
	volumeInUsd := price.Mul(decimal.NewFromFloat(volume))
	payload := collection.TradingVolumeUpdatePayload{VolumeInUsd: volumeInUsd.InexactFloat64()}
	if err := u.repo.Patch(ctx, id, payload); err != nil {
		ctx.WithFields(log.Fields{
			"id":      id,
			"payload": payload,
			"err":     err,
		}).Error("repo.Patch failed")
		return err
	}
	return nil
}
