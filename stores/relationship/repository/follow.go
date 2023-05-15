package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/follow"
	"github.com/x-xyz/goapi/service/query"
)

type followImpl struct {
	q query.Mongo
}

func NewFollow(q query.Mongo) follow.Repo {
	return &followImpl{q}
}

func (im *followImpl) Upsert(c ctx.Ctx, from, to domain.Address) error {
	rel := follow.Follow{
		From: from.ToLower(),
		To:   to.ToLower(),
	}
	if err := im.q.Upsert(c, domain.TableFollows, rel, rel); err != nil {
		c.WithField("err", err).Error("upsert follow relation falied")
		return err
	}
	return nil
}

func (im *followImpl) Remove(c ctx.Ctx, from, to domain.Address) error {
	rel := follow.Follow{
		From: from.ToLower(),
		To:   to.ToLower(),
	}
	// ignore ErrNotFound since the relation doesn't exist
	if err := im.q.Remove(c, domain.TableFollows, rel); err != nil && err != query.ErrNotFound {
		c.WithField("err", err).Error("remove follow relation falied")
		return err
	}
	return nil
}

func (im *followImpl) FindAll(c ctx.Ctx, optFns ...follow.FindAllOptions) ([]*follow.Follow, error) {
	opts, err := follow.GetFindAllOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("follow.GetFindAllOptions failed")
		return nil, err
	}

	offset := int(0)

	limit := int(0)

	sort := "_id"

	if opts.Offset != nil {
		offset = int(*opts.Offset)
	}

	if opts.Limit != nil {
		limit = int(*opts.Limit)
	}

	if opts.SortBy != nil && opts.SortDir != nil {
		sort = *opts.SortBy
		if *opts.SortDir == domain.SortDirDesc {
			sort = "-" + sort
		}
	}

	res := []*follow.Follow{}

	if qry, err := mongoclient.MakeBsonM(opts); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return nil, err
	} else if err := im.q.Search(c, domain.TableFollows, offset, limit, sort, qry, &res); err != nil {
		c.WithField("err", err).Error("q.Search failed")
		return nil, err
	} else {
		return res, nil
	}
}

func (im *followImpl) Count(c ctx.Ctx, optFns ...follow.FindAllOptions) (int, error) {
	opts, err := follow.GetFindAllOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("follow.GetFindAllOptions failed")
		return 0, err
	}

	if qry, err := mongoclient.MakeBsonM(opts); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return 0, err
	} else if count, err := im.q.Count(c, domain.TableFollows, qry); err != nil {
		c.WithField("err", err).Error("q.Count failed")
		return 0, err
	} else {
		return count, nil
	}
}

func (im *followImpl) FindOne(c ctx.Ctx, from, to domain.Address) (*follow.Follow, error) {
	res := &follow.Follow{}

	if qry, err := mongoclient.MakeBsonM(follow.Follow{From: from.ToLower(), To: to.ToLower()}); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return nil, err
	} else if err := im.q.FindOne(c, domain.TableFollows, qry, res); err != nil && err != query.ErrNotFound {
		c.WithField("err", err).Error("q.FindOne failed")
		return nil, err
	} else if err == query.ErrNotFound {
		return nil, nil
	} else {
		return res, nil
	}
}
