package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/twelvefold"
	"github.com/x-xyz/goapi/service/query"
)

type twelvefoldImpl struct {
	q query.Mongo
}

func NewTwelvefoldRepo(q query.Mongo) twelvefold.TwelvefoldRepo {
	return &twelvefoldImpl{q: q}
}

func (im *twelvefoldImpl) FindAll(c ctx.Ctx, optFns ...twelvefold.TwelvefoldFindAllOptionsFunc) ([]twelvefold.Twelvefold, error) {
	opts, err := twelvefold.GetTwelvefoldFindAllOptions(optFns...)
	if err != nil {
		c.WithField("err", err).Error("twelvefold.GetTwelvefoldFindAllOptions failed")
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

	qry, err := mongoclient.MakeBsonM(opts)

	if err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return nil, err
	}

	res := []twelvefold.Twelvefold{}

	if err := im.q.Search(c, domain.TableTwelvefold, offset, limit, "_id", qry, &res); err != nil {
		c.WithField("err", err).Error("q.Search failed")
		return nil, err
	} else {
		return res, nil
	}
}
