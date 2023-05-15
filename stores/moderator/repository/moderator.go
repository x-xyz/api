package repository

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/moderator"
	"github.com/x-xyz/goapi/service/query"
)

type impl struct {
	q query.Mongo
}

func New(q query.Mongo) moderator.Repo {
	return &impl{q}
}

func (im *impl) FindAll(c ctx.Ctx) ([]*moderator.Moderator, error) {
	res := []*moderator.Moderator{}

	// to prevent scancol error
	qry := bson.M{"address": bson.M{"$exists": true}}

	if err := im.q.Search(c, domain.TableModerators, 0, 0, "_id", qry, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (im *impl) FindOne(c ctx.Ctx, address domain.Address) (*moderator.Moderator, error) {
	res := &moderator.Moderator{}

	if qry, err := mongoclient.MakeBsonM(&moderator.Moderator{Address: address.ToLower()}); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return nil, err
	} else if err := im.q.FindOne(c, domain.TableModerators, qry, res); err != nil && err != query.ErrNotFound {
		c.WithField("err", err).Error("q.FindOne failed")
		return nil, err
	} else if err == query.ErrNotFound {
		return nil, nil
	}
	return res, nil
}

func (im *impl) Create(c ctx.Ctx, value moderator.Moderator) error {
	if err := im.q.Insert(c, domain.TableModerators, value); err != nil {
		c.WithField("err", err).Error("q.Insert failed")
		return err
	}
	return nil
}

func (im *impl) Delete(c ctx.Ctx, address domain.Address) error {
	if slr, err := mongoclient.MakeBsonM(moderator.Moderator{Address: address}); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return err
	} else if err := im.q.Remove(c, domain.TableModerators, slr); err != nil {
		c.WithField("err", err).Error("q.Remove failed")
		return err
	}
	return nil
}
