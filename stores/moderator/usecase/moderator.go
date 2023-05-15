package usecase

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/moderator"
)

type impl struct {
	moderator moderator.Repo
}

func New(moderator moderator.Repo) moderator.Usecase {
	return &impl{moderator}
}

func (im *impl) FindAll(c ctx.Ctx) ([]*moderator.Moderator, error) {
	return im.moderator.FindAll(c)
}

func (im *impl) IsModerator(c ctx.Ctx, address domain.Address) (bool, error) {
	if res, err := im.moderator.FindOne(c, address); err != nil {
		c.WithField("err", err).Error("moderator.FindOne failed")
		return false, err
	} else {
		return res != nil, nil
	}
}

func (im *impl) Add(c ctx.Ctx, address domain.Address, name string) error {
	if err := im.moderator.Create(c, moderator.Moderator{Address: address.ToLower(), Name: name}); err != nil {
		c.WithField("err", err).Error("moderator.Create failed")
		return err
	}
	return nil
}

func (im *impl) Remove(c ctx.Ctx, address domain.Address) error {
	if err := im.moderator.Delete(c, address); err != nil {
		c.WithField("err", err).Error("moderator.Delete failed")
		return err
	}
	return nil
}
