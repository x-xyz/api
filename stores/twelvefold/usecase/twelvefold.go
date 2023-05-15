package usecase

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/twelvefold"
)

type twelvefoldImpl struct {
	repo twelvefold.TwelvefoldRepo
}

func NewTwelvefoldUseCase(repo twelvefold.TwelvefoldRepo) twelvefold.TwelvefoldUsecase {
	return &twelvefoldImpl{repo: repo}
}

func (im *twelvefoldImpl) FindAll(c ctx.Ctx, optFns ...twelvefold.TwelvefoldFindAllOptionsFunc) ([]twelvefold.Twelvefold, error) {
	items, err := im.repo.FindAll(c, optFns...)
	if err != nil {
		c.WithField("err", err).Error("repo.FindAll failed")
		return nil, err
	}
	return items, nil
}
