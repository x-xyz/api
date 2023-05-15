package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/erc721/contract"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type erc721contractImpl struct {
	q query.Mongo
}

func NewErc721Contract(q query.Mongo) contract.Repo {
	return &erc721contractImpl{q}
}

func (im *erc721contractImpl) FindAll(c ctx.Ctx, optFns ...contract.FindOptions) ([]*contract.Contract, error) {
	opts, err := contract.GetFindOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("contract.GetFindOptions failed")
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

	res := []*contract.Contract{}

	if err := im.q.Search(c, domain.TableERC721Contracts, offset, limit, sort, query, &res); err != nil {
		c.WithField("err", err).Error("q.Search failed")
		return res, err
	}

	return res, nil
}

func (im *erc721contractImpl) FindOne(c ctx.Ctx, optFns ...contract.FindOptions) (*contract.Contract, error) {
	opts, err := contract.GetFindOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("contract.GetFindOptions failed")
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

	res := &contract.Contract{}

	if err := im.q.FindOne(c, domain.TableERC721Contracts, qry, res); err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		c.WithField("err", err).Error("q.FindOne failed")
		return nil, err
	}

	return res, nil
}

func (im *erc721contractImpl) Update(c ctx.Ctx, value contract.UpdatePayload, optFns ...contract.FindOptions) error {
	opts, err := contract.GetFindOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("contract.GetFindOptions failed")
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
	} else if err := im.q.Patch(c, domain.TableERC721Contracts, qry, val); err == query.ErrNotFound {
		return domain.ErrNotFound
	} else if err != nil {
		c.WithField("err", err).Error("q.Patch failed")
		return err
	}

	return nil
}

func (im *erc721contractImpl) Create(c ctx.Ctx, value contract.Contract) error {
	if err := im.q.Insert(c, domain.TableERC721Contracts, value); err != nil {
		c.WithField("err", err).WithField("value", value).Error("q.Insert failed")
		return err
	}
	return nil
}
