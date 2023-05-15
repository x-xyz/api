package usecase

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/follow"
)

type followImpl struct {
	follow follow.Repo
}

func NewFollow(follow follow.Repo) follow.Usecase {
	return &followImpl{follow}
}

func (im *followImpl) Follow(c ctx.Ctx, from, to domain.Address) error {
	if err := im.follow.Upsert(c, from, to); err != nil {
		c.WithField("err", err).Error("follow.Follow failed")
		return err
	}
	return nil
}

func (im *followImpl) Unfollow(c ctx.Ctx, from, to domain.Address) error {
	if err := im.follow.Remove(c, from, to); err != nil {
		c.WithField("err", err).Error("follow.Remove failed")
		return err
	}
	return nil
}

func (im *followImpl) GetFollowers(c ctx.Ctx, address domain.Address) ([]domain.Address, error) {
	if res, err := im.follow.FindAll(c, follow.WithTo(address), follow.WithPagination(0, 5000)); err != nil {
		c.WithField("err", err).Error("follow.FindAll failed")
		return nil, err
	} else {
		addresses := []domain.Address{}
		for _, follow := range res {
			addresses = append(addresses, follow.From)
		}
		return addresses, nil
	}
}

func (im *followImpl) GetFollowerCount(c ctx.Ctx, address domain.Address) (int, error) {
	if res, err := im.follow.Count(c, follow.WithTo(address)); err != nil {
		c.WithField("err", err).Error("follow.Count failed")
		return 0, err
	} else {
		return res, nil
	}
}

func (im *followImpl) GetFollowings(c ctx.Ctx, address domain.Address) ([]domain.Address, error) {
	if res, err := im.follow.FindAll(c, follow.WithFrom(address), follow.WithPagination(0, 5000)); err != nil {
		c.WithField("err", err).Error("follow.FindAll failed")
		return nil, err
	} else {
		addresses := []domain.Address{}
		for _, follow := range res {
			addresses = append(addresses, follow.To)
		}
		return addresses, nil
	}
}

func (im *followImpl) GetFollowingCount(c ctx.Ctx, address domain.Address) (int, error) {
	if res, err := im.follow.Count(c, follow.WithFrom(address)); err != nil {
		c.WithField("err", err).Error("follow.Count failed")
		return 0, err
	} else {
		return res, nil
	}
}

func (im *followImpl) IsFollowing(c ctx.Ctx, from, to domain.Address) (bool, error) {
	if res, err := im.follow.FindOne(c, from, to); err != nil {
		return false, err
	} else {
		return res != nil, nil
	}
}
