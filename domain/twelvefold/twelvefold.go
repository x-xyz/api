package twelvefold

import (
	"github.com/x-xyz/goapi/base/ctx"
)

type Twelvefold struct {
	Edition  string `bson:"edition"`
	Series   string `bson:"series"`
	Season   string `bson:"season"`
	Satpoint string `bson:"satpoint"`
}

type TwelvefoldSelectOptions struct {
	Edition *string `bson:"edition"`
	Series  *string `bson:"series"`
	Season  *string `bson:"season"`
	Dummy   *bool   `bson:"dummy"`
	Offset  *int32  `bson:"-"`
	Limit   *int32  `bson:"-"`
}

type TwelvefoldFindAllOptionsFunc func(*TwelvefoldSelectOptions) error

func GetTwelvefoldFindAllOptions(opts ...TwelvefoldFindAllOptionsFunc) (TwelvefoldSelectOptions, error) {
	res := TwelvefoldSelectOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func TwelvefoldWithPagination(offset int32, limit int32) TwelvefoldFindAllOptionsFunc {
	return func(options *TwelvefoldSelectOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func TwelvefoldWithEdition(edition string) TwelvefoldFindAllOptionsFunc {
	return func(options *TwelvefoldSelectOptions) error {
		options.Edition = &edition
		return nil
	}
}

func TwelvefoldWithSeries(series string) TwelvefoldFindAllOptionsFunc {
	return func(options *TwelvefoldSelectOptions) error {
		options.Series = &series
		return nil
	}
}

func TwelvefoldWithSeason(season string) TwelvefoldFindAllOptionsFunc {
	return func(options *TwelvefoldSelectOptions) error {
		options.Season = &season
		return nil
	}
}

func TwelvefoldWithDummy(dummy bool) TwelvefoldFindAllOptionsFunc {
	return func(options *TwelvefoldSelectOptions) error {
		options.Dummy = &dummy
		return nil
	}
}

type TwelvefoldRepo interface {
	FindAll(ctx.Ctx, ...TwelvefoldFindAllOptionsFunc) ([]Twelvefold, error)
}

type TwelvefoldUsecase interface {
	FindAll(ctx.Ctx, ...TwelvefoldFindAllOptionsFunc) ([]Twelvefold, error)
}
