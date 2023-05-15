package statistic

import bCtx "github.com/x-xyz/goapi/base/ctx"

var (
	Apeburned = "apeburned"
)

type Statistic struct {
	Key   string `bson:"key"`
	Value string `bson:"Value"`
}

type StatisticId struct {
	Key string `bson:"key"`
}

func (s *Statistic) ToId() StatisticId {
	return StatisticId{Key: s.Key}
}

type Repo interface {
	FindOne(ctx bCtx.Ctx, key string) (*Statistic, error)
	Upsert(ctx bCtx.Ctx, s *Statistic) error
}

type UseCase interface {
	Get(ctx bCtx.Ctx, key string) (string, error)
	Set(ctx bCtx.Ctx, key string, value string) error
}
