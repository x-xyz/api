package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

func makeFindQuery(optFns ...collection.FindAllOptions) (bson.M, error) {
	opts, err := collection.GetFindAllOptions(optFns...)
	if err != nil {
		return nil, err
	}

	query := bson.M{}

	if opts.ChainId != nil {
		query["chainId"] = *opts.ChainId
	}

	if opts.Addresses != nil {
		query["erc721Address"] = bson.M{"$in": *opts.Addresses}
	}

	if opts.Category != nil {
		query["categories"] = *opts.Category
	}

	if opts.Status != nil {
		query["status"] = *opts.Status
	}

	if opts.IsAppropriate != nil {
		query["isAppropriate"] = *opts.IsAppropriate
	}

	if opts.IsInternal != nil {
		query["isInternal"] = *opts.IsInternal
	}

	if opts.IsOwnerble != nil {
		query["isOwnerble"] = *opts.IsOwnerble
	}

	if opts.Owner != nil {
		query["owner"] = *opts.Owner
	}

	if opts.FloorPriceGTE != nil || opts.FloorPriceLTE != nil {
		subQuery := bson.M{}
		if opts.FloorPriceGTE != nil {
			subQuery["$gte"] = *opts.FloorPriceGTE
		}
		if opts.FloorPriceLTE != nil {
			subQuery["$lte"] = *opts.FloorPriceLTE
		}
		query["floorPrice"] = subQuery
	}

	if opts.UsdFloorPriceGTE != nil || opts.UsdFloorPriceLTE != nil {
		subQuery := bson.M{}
		if opts.UsdFloorPriceGTE != nil {
			subQuery["$gte"] = *opts.UsdFloorPriceGTE
		}
		if opts.UsdFloorPriceLTE != nil {
			subQuery["$lte"] = *opts.UsdFloorPriceLTE
		}
		query["usdFloorPrice"] = subQuery
	}

	queryOr := bson.A{}

	if opts.AccountEditable != nil {
		queryOr = append(queryOr, bson.M{
			"editableAccounts": *opts.AccountEditable,
		})
		queryOr = append(queryOr, bson.M{
			"owner": *opts.AccountEditable,
		})
	}

	if len(queryOr) > 0 {
		query["$or"] = queryOr
	}

	return query, nil
}

type collectionImpl struct {
	q query.Mongo
}

func NewCollection(q query.Mongo) collection.Repo {
	return &collectionImpl{q}
}

func (im *collectionImpl) FindAll(c ctx.Ctx, optFns ...collection.FindAllOptions) ([]*collection.Collection, error) {
	res := []*collection.Collection{}

	opts, err := collection.GetFindAllOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("collection.GetFindAllOptions failed")
		return res, err
	}

	offset := int(0)

	limit := int(0)

	sort := []string{"_id"}

	query, err := makeFindQuery(optFns...)
	if err != nil {
		return res, err
	}

	if opts.Offset != nil {
		offset = int(*opts.Offset)
	}

	if opts.Limit != nil {
		limit = int(*opts.Limit)
	}

	if opts.SortBy != nil && opts.SortDir != nil {
		sort = []string{}
		sortBy := *opts.SortBy

		// to keep empty data has lowest order by any way
		switch sortBy {
		case "floorPrice":
			fallthrough
		case "usdFloorPrice":
			sort = append(sort, "-hasFloorPrice")
		case "lastSoldAt":
			sort = append(sort, "-hasBeenSold")
		case "lastListedAt":
			sort = append(sort, "-hasBeenListed")
		}

		if *opts.SortDir == domain.SortDirDesc {
			sortBy = "-" + sortBy
		}
		sort = append(sort, sortBy)

		if len(query) == 0 {
			query[*opts.SortBy] = bson.M{"$exists": true}
		}
	}

	if err := im.q.SearchNSorts(c, domain.TableCollections, offset, limit, sort, query, &res); err != nil {
		c.WithField("err", err).Error("q.Search failed")
		return res, err
	}

	return res, nil
}

func (im *collectionImpl) Count(c ctx.Ctx, opts ...collection.FindAllOptions) (int, error) {
	qry, err := makeFindQuery(opts...)
	if err != nil {
		return 0, err
	}

	res, err := im.q.Count(c, domain.TableCollections, qry)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (im *collectionImpl) FindOne(c ctx.Ctx, id collection.CollectionId) (*collection.Collection, error) {
	res := &collection.Collection{}

	qry, err := mongoclient.MakeBsonM(collection.Collection{ChainId: id.ChainId, Erc721Address: id.Address})
	if err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return nil, err
	}
	if err := im.q.FindOne(c, domain.TableCollections, qry, res); err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		c.WithField("err", err).Error("q.FindOne failed")
		return nil, err
	}

	return res, nil
}

func (im *collectionImpl) Create(c ctx.Ctx, value collection.CreatePayload) error {
	if err := im.q.Insert(c, domain.TableCollections, value); err != nil {
		c.WithField("err", err).Error("q.Insert failed")
		return err
	}

	return nil
}

func (im *collectionImpl) Upsert(c ctx.Ctx, value collection.CreatePayload) error {
	id := collection.CollectionId{ChainId: value.ChainId, Address: value.Erc721Address}

	if err := im.q.Upsert(c, domain.TableCollections, id, value); err != nil {
		c.WithField("err", err).Error("q.Upsert failed")
		return err
	}

	return nil
}

func (im *collectionImpl) Update(c ctx.Ctx, id collection.CollectionId, value collection.UpdatePayload) error {
	if slt, err := mongoclient.MakeBsonM(id); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return err
	} else if val, err := mongoclient.MakeBsonM(value); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return err
	} else if err := im.q.Patch(c, domain.TableCollections, slt, val); err == query.ErrNotFound {
		return domain.ErrNotFound
	} else if err != nil {
		c.WithField("err", err).Error("q.Patch failed")
		return err
	}

	return nil
}

func (im *collectionImpl) IncreaseViewCount(c ctx.Ctx, id collection.CollectionId, count int) (int32, error) {
	res := &collection.Collection{}
	if err := im.q.Increment(c, domain.TableCollections, id, res, "viewCount", count); err != nil {
		return 0, err
	}
	return res.ViewCount, nil
}

func (im *collectionImpl) IncreaseLikeCount(c ctx.Ctx, id collection.CollectionId, count int) (int32, error) {
	res := &collection.Collection{}
	if err := im.q.Increment(c, domain.TableCollections, id, res, "liked", count); err != nil {
		return 0, err
	}
	return res.Liked, nil
}
