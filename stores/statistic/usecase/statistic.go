package usecase

import (
	"github.com/x-xyz/goapi/domain/statistic"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
)

type uc struct {
	statisticRepo statistic.Repo
}

func New(repo statistic.Repo) statistic.UseCase {
	return &uc{repo}
}

func (u *uc) Get(ctx bCtx.Ctx, key string) (string, error) {
	s, err := u.statisticRepo.FindOne(ctx, key)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"key": key,
		}).Error("repo.FindOne failed")
		return "", err
	}
	return s.Value, nil
}

func (u *uc) Set(ctx bCtx.Ctx, key string, value string) error {
	s := &statistic.Statistic{Key: key, Value: value}
	err := u.statisticRepo.Upsert(ctx, s)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"s":   s,
		}).Error("repo.Upsert failed")
		return err
	}
	return nil
}
