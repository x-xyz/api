package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/erc1155"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type contractImpl struct {
	q query.Mongo
}

func NewContractRepo(q query.Mongo) erc1155.Repo {
	return &contractImpl{q}
}

func (im *contractImpl) FindAll(c ctx.Ctx, optFns ...erc1155.FindOptions) ([]*erc1155.Contract, error) {
	opts, err := erc1155.GetFindOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("erc1155.GetFindOptions failed")
		return nil, err
	}

	offset := int(0)

	limit := int(0)

	sort := "_id"

	query := bson.M{}

	if opts.Offset != nil {
		offset = int(*opts.Offset)
	}

	if opts.Limit != nil {
		limit = int(*opts.Limit)
	}

	if opts.ChainId != nil {
		query["chainId"] = *opts.ChainId
	}

	if opts.Address != nil {
		query["address"] = *opts.Address
	}

	if opts.IsAppropriate != nil {
		query["isAppropriate"] = *opts.IsAppropriate
	}

	if opts.SortBy != nil && opts.SortDir != nil {
		sort = *opts.SortBy
		if *opts.SortDir == domain.SortDirDesc {
			sort = "-" + sort
		}
		if len(query) == 0 {
			query[*opts.SortBy] = bson.M{"$exists": true}
		}
	}

	res := []*erc1155.Contract{}

	if err := im.q.Search(c, domain.TableERC1155Contracts, offset, limit, sort, query, &res); err != nil {
		c.WithField("err", err).Error("q.Search failed")
		return res, err
	}

	return res, nil
}

func (im *contractImpl) FindOne(c ctx.Ctx, optFns ...erc1155.FindOptions) (*erc1155.Contract, error) {
	opts, err := erc1155.GetFindOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("erc1155.GetFindOptions failed")
		return nil, err
	}

	qry := bson.M{}

	if opts.ChainId != nil {
		qry["chainId"] = *opts.ChainId
	}

	if opts.Address != nil {
		qry["address"] = *opts.Address
	}

	if opts.IsAppropriate != nil {
		qry["isAppropriate"] = *opts.IsAppropriate
	}

	res := &erc1155.Contract{}

	if err := im.q.FindOne(c, domain.TableERC1155Contracts, qry, res); err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		c.WithField("err", err).Error("q.FindOne failed")
		return nil, err
	}

	return res, nil
}

func (im *contractImpl) Update(c ctx.Ctx, value erc1155.UpdatePayload, optFns ...erc1155.FindOptions) error {
	opts, err := erc1155.GetFindOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("erc1155.GetFindOptions failed")
		return err
	}

	qry := bson.M{}

	if opts.ChainId != nil {
		qry["chainId"] = *opts.ChainId
	}

	if opts.Address != nil {
		qry["address"] = *opts.Address
	}

	if opts.IsAppropriate != nil {
		qry["isAppropriate"] = *opts.IsAppropriate
	}

	if val, err := mongoclient.MakeBsonM(value); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return err
	} else if err := im.q.Patch(c, domain.TableERC1155Contracts, qry, val); err == query.ErrNotFound {
		return domain.ErrNotFound
	} else if err != nil {
		c.WithField("err", err).Error("q.Patch failed")
		return err
	}

	return nil
}

func (im *contractImpl) Create(c ctx.Ctx, value erc1155.Contract) error {
	if err := im.q.Insert(c, domain.TableERC1155Contracts, value); err != nil {
		c.WithField("err", err).WithField("value", value).Error("q.Insert failed")
		return err
	}
	return nil
}
