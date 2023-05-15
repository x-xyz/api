package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/unlockable"
	"github.com/x-xyz/goapi/service/query"
)

type unlockableImpl struct {
	q query.Mongo
}

func NewUnlockable(q query.Mongo) unlockable.Repo {
	return &unlockableImpl{q}
}

func (im *unlockableImpl) FindOne(c ctx.Ctx, id unlockable.UnlockableId) (*unlockable.Unlockable, error) {
	res := &unlockable.Unlockable{}

	if err := im.q.FindOne(c, domain.TableUnlockableContents, id, res); err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		c.WithField("err", err).Error("q.FindOne failed")
		return nil, err
	}

	return res, nil
}

func (im *unlockableImpl) Create(c ctx.Ctx, val unlockable.Unlockable) error {
	if err := im.q.Insert(c, domain.TableUnlockableContents, val); err != nil {
		c.WithField("err", err).Error("q.Insert failed")
		return err
	}

	return nil
}
