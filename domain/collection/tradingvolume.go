package collection

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type PeriodType uint8

const (
	PeriodTypeUnknown PeriodType = iota
	PeriodTypeDay                // 24 hours
	PeriodTypeWeek               // 7 days
	PeriodTypeMonth              // 30 days
	PeriodTypeAll                // all time
	PeriodTypeOneHour
	PeriodTypeSixHour
)

type TradingVolume struct {
	ChainId     domain.ChainId `json:"chainId" bson:"chainId"`
	Address     domain.Address `json:"address" bson:"address"`
	Period      PeriodType     `json:"periodType" bson:"periodType"`
	Date        time.Time      `json:"date" bson:"date"`     // utc 00:00
	Volume      float64        `json:"volume" bson:"volume"` // native
	VolumeInUsd float64        `json:"volumeInUsd" bson:"volumeInUsd"`
}

func (v *TradingVolume) ToId() TradingVolumeId {
	return TradingVolumeId{
		ChainId: v.ChainId,
		Address: v.Address,
		Period:  v.Period,
		Date:    v.Date,
	}
}

type TradingVolumeId struct {
	ChainId domain.ChainId `json:"chainId" bson:"chainId"`
	Address domain.Address `json:"address" bson:"address"`
	Period  PeriodType     `json:"periodType" bson:"periodType"`
	Date    time.Time      `json:"date" bson:"date"` // utc 00:00
}

type TradingVolumeUpdatePayload struct {
	VolumeInUsd float64 `bson:"volumeInUsd,omitempty"`
}

type tradingVolumeFindAllOptions struct {
	SortBy     *string         `bson:"-"`
	SortDir    *domain.SortDir `bson:"-"`
	Offset     *int32          `bson:"-"`
	Limit      *int32          `bson:"-"`
	ChainId    *domain.ChainId `bson:"chainId"`
	PeriodType *PeriodType     `bson:"periodType"`
	Date       *time.Time      `bson:"date"`
}
type TradingVolumeFindAllOptions func(*tradingVolumeFindAllOptions) error

func GetTradingVolumeFindAllOptions(opts ...TradingVolumeFindAllOptions) (tradingVolumeFindAllOptions, error) {
	res := tradingVolumeFindAllOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func TradingVolumeWithSort(sortby string, sortdir domain.SortDir) TradingVolumeFindAllOptions {
	return func(options *tradingVolumeFindAllOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func TradingVolumeWithPagination(offset int32, limit int32) TradingVolumeFindAllOptions {
	return func(options *tradingVolumeFindAllOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func TradingVolumeWithChainId(chainId domain.ChainId) TradingVolumeFindAllOptions {
	return func(options *tradingVolumeFindAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

func TradingVolumeWithPeriodType(periodType PeriodType) TradingVolumeFindAllOptions {
	return func(options *tradingVolumeFindAllOptions) error {
		options.PeriodType = &periodType
		return nil
	}
}

func TradingVolumeWithDate(date time.Time) TradingVolumeFindAllOptions {
	return func(options *tradingVolumeFindAllOptions) error {
		options.Date = &date
		return nil
	}
}

type TradingVolumeUseCase interface {
	FindOne(ctx.Ctx, TradingVolumeId) (*TradingVolume, error)
	Upsert(ctx.Ctx, TradingVolume) error
	IncDailyVolume(ctx.Ctx, domain.ChainId, domain.Address, time.Time, float64) (float64, error)
	IncTotalVolume(ctx.Ctx, domain.ChainId, domain.Address, float64) (float64, error)
}

type TradingVolumeRepo interface {
	FindAll(c ctx.Ctx, opts ...TradingVolumeFindAllOptions) ([]TradingVolume, error)
	FindOne(ctx.Ctx, TradingVolumeId) (*TradingVolume, error)
	Upsert(ctx.Ctx, TradingVolume) error
	Patch(ctx.Ctx, TradingVolumeId, TradingVolumeUpdatePayload) error
	IncVolume(ctx.Ctx, TradingVolumeId, float64) (float64, error)
}
