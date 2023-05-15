package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/like"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type likeImpl struct {
	q query.Mongo
}

func NewLike(q query.Mongo) like.Repo {
	return &likeImpl{q}
}

func (im *likeImpl) FindAll(c ctx.Ctx, optFns ...like.SelectOptions) ([]*like.Like, error) {
	opts, err := like.GetSelectOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("like.GetSelectOptions failed")
		return nil, err
	}

	offset := int(0)

	limit := int(0)

	if opts.Offset != nil {
		offset = int(*opts.Offset)
	}

	if opts.Limit != nil {
		limit = int(*opts.Limit)
	}

	res := []*like.Like{}

	qry, err := mongoclient.MakeBsonM(opts)

	if err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return nil, err
	}

	if len(opts.ContractAddresses) > 0 {
		qry["contractAddress"] = bson.M{"$in": opts.ContractAddresses}
	}

	if err := im.q.Search(c, domain.TableLiks, offset, limit, "_id", qry, &res); err != nil {
		c.WithField("err", err).Error("q.Search failed")
		return nil, err
	} else {
		return res, nil
	}
}

func (im *likeImpl) Count(c ctx.Ctx, optFns ...like.SelectOptions) (int, error) {
	opts, err := like.GetSelectOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("like.GetSelectOptions failed")
		return 0, err
	}

	if qry, err := mongoclient.MakeBsonM(opts); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return 0, err
	} else if count, err := im.q.Count(c, domain.TableLiks, qry); err != nil {
		c.WithField("err", err).Error("q.Count failed")
		return 0, err
	} else {
		return count, nil
	}
}

func (im *likeImpl) Create(c ctx.Ctx, value like.Like) error {
	value.ContractAddress = value.ContractAddress.ToLower()
	value.Liker = value.Liker.ToLower()
	if err := im.q.Insert(c, domain.TableLiks, value); err != nil {
		return err
	}
	return nil
}

func (im *likeImpl) Delete(c ctx.Ctx, optFns ...like.SelectOptions) error {
	opts, err := like.GetSelectOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("like.GetSelectOptions failed")
		return err
	}

	if slt, err := mongoclient.MakeBsonM(opts); err != nil {
		return err
	} else if err := im.q.Remove(c, domain.TableLiks, slt); err != nil && err != query.ErrNotFound {
		return err
	} else if err == query.ErrNotFound {
		return domain.ErrNotFound
	}
	return nil
}
