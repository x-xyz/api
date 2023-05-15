package repository

import (
	"errors"
	"strconv"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/service/cache"
	compoundcache "github.com/x-xyz/goapi/service/cache/compoundCache"
	"github.com/x-xyz/goapi/service/cache/provider/primitive"
	redisCache "github.com/x-xyz/goapi/service/cache/provider/redis"
	"github.com/x-xyz/goapi/service/query"
	"github.com/x-xyz/goapi/service/redis"
	"go.mongodb.org/mongo-driver/bson"
)

const zeroAddress = "0x0000000000000000000000000000000000000000"

func makeFindQuery(opts nftitem.FindAllOptions) (query bson.M) {
	query = bson.M{}
	orQueries := bson.A{}
	andExprs := []bson.M{}
	offersCondi := bson.M{}
	offersPriceCondi := bson.M{}

	if opts.SortBy != nil {
		switch *opts.SortBy {
		case "lastSalePriceInUSD":
			query[*opts.SortBy] = bson.M{"$gt": 0}
		case "saleEndsAt":
			query[*opts.SortBy] = bson.M{"$gt": time.Now()}
		}
	}

	if len(opts.ContractAddresses) > 1 {
		query["contractAddress"] = bson.M{"$in": opts.ContractAddresses}
	} else if len(opts.ContractAddresses) == 1 {
		query["contractAddress"] = opts.ContractAddresses[0]
	}

	if opts.ChainId != nil {
		query["chainId"] = *opts.ChainId
	}

	if opts.Owner != nil && opts.HoldingIds != nil {
		subQuery := bson.A{}
		if *opts.NotOwner {
			subQuery = append(subQuery, bson.M{"$ne": opts.Owner.ToLowerStr()})
		} else {
			subQuery = append(subQuery, bson.M{"owner": opts.Owner.ToLower()})
		}
		for _, id := range *opts.HoldingIds {
			subQuery = append(subQuery, bson.M{
				"chainId":         id.ChainId,
				"contractAddress": id.ContractAddress.ToLower(),
				"tokenID":         id.TokenId,
			})
		}
		andExprs = append(andExprs, bson.M{"$or": subQuery})
	} else if opts.Owner != nil {
		if *opts.NotOwner {
			query["owner"] = bson.M{"$ne": opts.Owner.ToLowerStr()}
		} else {
			query["owner"] = opts.Owner.ToLowerStr()
		}
	} else {
		// default ignore owner is zero address
		query["owner"] = bson.M{"$ne": zeroAddress}
	}

	if opts.ListingFrom != nil {
		query["listingOwners"] = opts.ListingFrom.ToLowerStr()
	}

	if opts.InactiveListingFrom != nil {
		query["inactiveListingOwners"] = opts.InactiveListingFrom.ToLowerStr()
	}

	if opts.IsAppropriate != nil {
		query["isAppropriate"] = *opts.IsAppropriate
	}

	if len(opts.IndexerStates) == 1 {
		query["indexerState"] = opts.IndexerStates[0]
	} else if len(opts.IndexerStates) > 1 {
		query["indexerState"] = bson.M{"$in": opts.IndexerStates}
	}

	if opts.IndexerRetryCountLT != nil {
		query["indexerRetryCount"] = bson.M{"$lt": opts.IndexerRetryCountLT}
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusBuyNow) {
		query["listingEndsAt"] = bson.M{"$gt": time.Now()}
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusHasOffer) {
		query["offerEndsAt"] = bson.M{"$gt": time.Now()}
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusOnAuction) {
		// query["auction"] = bson.M{"$exists": true, "$not": bson.M{"$eq": nil}}
		// query["auction.blockNumber"] = bson.M{"$not": bson.M{"$eq": 0}}
		query["saleEndsAt"] = bson.M{"$gt": time.Now()}
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusHasBid) {
		query["highestBid"] = bson.M{"$exists": true, "$not": bson.M{"$eq": nil}}
		query["highestBid.blockNumber"] = bson.M{"$not": bson.M{"$eq": 0}}
	}

	if nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusHasTraded) {
		query["soldAt"] = bson.M{"$exists": true, "$not": bson.M{"$eq": nil}}
	}

	if len(opts.OfferOwners) > 0 {
		query["offerOwners"] = bson.M{"$in": opts.OfferOwners}
	}

	if opts.PriceGTE != nil || opts.PriceLTE != nil {
		subquery := bson.M{"$ne": 0}
		if opts.PriceGTE != nil {
			subquery["$gte"] = *opts.PriceGTE
		}
		if opts.PriceLTE != nil {
			subquery["$lte"] = *opts.PriceLTE
		}
		query["price"] = subquery
	}

	if opts.PriceInUsdGTE != nil || opts.PriceInUsdLTE != nil {
		subquery := bson.M{"$ne": 0}
		if opts.PriceInUsdGTE != nil {
			subquery["$gte"] = *opts.PriceInUsdGTE
		}
		if opts.PriceInUsdLTE != nil {
			subquery["$lte"] = *opts.PriceInUsdLTE
		}
		query["priceInUSD"] = subquery
	}

	if opts.OfferPriceInUsdGTE != nil {
		offersPriceCondi["$gte"] = *opts.OfferPriceInUsdGTE
		offersCondi["priceInUsd"] = offersPriceCondi
	}

	if opts.OfferPriceInUsdLTE != nil {
		offersPriceCondi["$lte"] = *opts.OfferPriceInUsdLTE
		offersCondi["priceInUsd"] = offersPriceCondi
	}

	if len(offersCondi) > 0 || nftitem.HasSaleStatus(opts.SaleStatus, nftitem.SaleStatusHasOfferWithExpired) {
		query["offers"] = bson.M{
			"$exists":    true,
			"$not":       bson.M{"$size": 0},
			"$elemMatch": offersCondi,
		}
	}

	if opts.TokenType != nil {
		query["tokenType"] = *opts.TokenType
	}

	if opts.BidOwner != nil {
		query["highestBid.owner"] = opts.BidOwner.ToLower()
	}

	if opts.ObjectIdLT != nil {
		query["_id"] = bson.M{"$lt": *opts.ObjectIdLT}
	}

	if opts.HasOrder != nil {
		query["hasOrder"] = *opts.HasOrder
	}

	// example:
	// {
	//    "contractAddress":"0xd03e287a677b015a649ef9fbd7267554fa4dd2d8",
	//    "$and":[
	//       {
	//          "$or":[
	//             {
	//                "attributes":{
	//                   "$elemMatch":{
	//                      "trait_type":"address",
	//                      "value":"0x35bcf180358e74d09dfe6c96f6ddc74262be506e"
	//                   }
	//                }
	//             },
	//             {
	//                "attributes":{
	//                   "$elemMatch":{
	//                      "trait_type":"address",
	//                      "value":"xxx"
	//                   }
	//                }
	//             }
	//          ]
	//       },
	//       {
	//          "attributes":{
	//             "$elemMatch":{
	//                "trait_type":"recipient",
	//                "value":"0x35bcf180358e74d09dfe6c96f6ddc74262be506e"
	//             }
	//          }
	//       }
	//    ]
	// }
	if len(opts.Attributes) > 0 {

		for _, attr := range opts.Attributes {
			orExprs := []bson.M{}
			for _, v := range attr.Values {
				orExprs = append(orExprs, bson.M{
					"attributes": bson.M{
						"$elemMatch": bson.M{
							"trait_type": attr.Name,
							"value":      v,
						},
					},
				})
			}
			andExprs = append(andExprs, bson.M{
				"$or": orExprs,
			})
		}
	}

	if opts.Ids != nil {
		orExprs := []bson.M{}
		for _, id := range *opts.Ids {
			orExprs = append(orExprs, bson.M{
				"chainId":         id.ChainId,
				"contractAddress": id.ContractAddress,
				"tokenID":         id.TokenId,
			})
		}
		andExprs = append(andExprs, bson.M{
			"$or": orExprs,
		})
	}

	if opts.Search != nil {
		orExprs := []bson.M{
			{
				"name": bson.M{"$regex": opts.Search, "$options": "i"},
			},
			{
				"attributes.value": bson.M{"$regex": opts.Search, "$options": "i"},
			},
		}
		andExprs = append(andExprs, bson.M{
			"$or": orExprs,
		})
	}

	if len(andExprs) > 0 {
		query["$and"] = andExprs
	}

	if len(orQueries) > 0 {
		query["$or"] = orQueries
	}

	if opts.Name != nil {
		query["name"] = bson.M{"$regex": opts.Name}
	}

	return query
}

type nftitemImpl struct {
	q            query.Mongo
	nftitemCache cache.Service
}

func NewNftItem(q query.Mongo, redis redis.Service) nftitem.Repo {
	cacheServices := []cache.Service{
		cache.New(cache.ServiceConfig{
			Ttl:   10 * time.Second,
			Pfx:   "nftitem",
			Cache: primitive.NewPrimitive("nftitem", 512),
		}),
	}

	if redis != nil {
		cacheServices = append(cacheServices, cache.New(cache.ServiceConfig{
			Ttl:   10 * time.Minute,
			Pfx:   "nftitem",
			Cache: redisCache.NewRedis(redis),
		}))
	}

	return &nftitemImpl{
		q:            q,
		nftitemCache: compoundcache.NewCompoundCache(cacheServices),
	}
}

func (im *nftitemImpl) FindAll(c ctx.Ctx, optFns ...nftitem.FindAllOptionsFunc) ([]*nftitem.NftItem, error) {
	opts, err := nftitem.GetFindAllOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("nftitem.GetFindAllOptions failed")
		return nil, err
	}

	offset := int(0)

	limit := int(0)

	// default sort by -createdAt
	// but there are some items without `createdAt` in dev db
	sort := []string{"-_id"}

	if opts.Offset != nil {
		offset = int(*opts.Offset)
	}

	if opts.Limit != nil {
		limit = int(*opts.Limit)
	}

	query := makeFindQuery(opts)

	if opts.SortBy != nil && opts.SortDir != nil {
		sort = []string{}
		sortBy := *opts.SortBy
		switch sortBy {
		case "listedAt":
			fallthrough
		case "priceInUSD":
			sort = append(sort, "-hasActiveListings")
		}

		if *opts.SortDir == domain.SortDirDesc {
			sortBy = "-" + sortBy
		}
		sort = append(sort, sortBy, "-_id")
	}

	if opts.Sorts != nil {
		sort = *opts.Sorts
	}

	res := []*nftitem.NftItem{}

	if err := im.q.SearchNSorts(c, domain.TableNFTItems, offset, limit, sort, query, &res); err != nil {
		c.WithFields(log.Fields{
			"err":   err,
			"query": query,
			"sort":  sort,
		}).Error("q.Search failed")
		return nil, err
	}

	return res, nil
}

func (im *nftitemImpl) Count(c ctx.Ctx, optFns ...nftitem.FindAllOptionsFunc) (int, error) {
	opts, err := nftitem.GetFindAllOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("nftitem.GetFindAllOptions failed")
		return 0, err
	}

	query := makeFindQuery(opts)

	if cnt, err := im.q.EstimateCount(c, domain.TableNFTItems, query); err != nil {
		return 0, nil
	} else {
		return cnt, nil
	}
}

func (im *nftitemImpl) FindOne(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) (*nftitem.NftItem, error) {
	key := keys.RedisKey(strconv.Itoa(int(chainId)), string(contract), string(tokenId))

	res := &nftitem.NftItem{}

	if err := im.nftitemCache.GetByFunc(c, key, res, func() (interface{}, error) {
		return im.findOne(c, chainId, contract, tokenId)
	}); err != nil {
		// c.WithFields(log.Fields{
		// 	"err":      err,
		// 	"chainId":  chainId,
		// 	"contract": contract,
		// 	"tokenId":  tokenId,
		// }).Error("nftitemCache.GetByFunc failed")
		return nil, err
	}

	return res, nil
}

func (im *nftitemImpl) findOne(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) (*nftitem.NftItem, error) {
	res := &nftitem.NftItem{}

	if err := im.q.FindOne(c, domain.TableNFTItems, bson.M{
		"chainId":         chainId,
		"contractAddress": contract,
		"tokenID":         tokenId,
	}, res); errors.Is(err, query.ErrNotFound) {
		return nil, domain.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return res, nil
}

func (im *nftitemImpl) Patch(c ctx.Ctx, id nftitem.Id, value nftitem.PatchableNftItem) error {
	ptrId := struct {
		ChainId         *domain.ChainId `bson:"chainId"`
		ContractAddress *domain.Address `bson:"contractAddress"`
		TokenId         *domain.TokenId `bson:"tokenID"`
	}{
		ChainId:         &id.ChainId,
		ContractAddress: &id.ContractAddress,
		TokenId:         &id.TokenId,
	}

	if slr, err := mongoclient.MakeBsonM(ptrId); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM for id failed")
		return err
	} else if val, err := mongoclient.MakeBsonM(value); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM for value failed")
		return err
	} else if err := im.q.Patch(c, domain.TableNFTItems, slr, val); err != nil {
		c.WithField("err", err).Error("q.Patch failed")
		return err
	}

	key := keys.RedisKey(strconv.Itoa(int(id.ChainId)), string(id.ContractAddress), string(id.TokenId))

	if err := im.nftitemCache.Del(c, key); err != nil {
		c.WithFields(log.Fields{
			"err":      err,
			"chainId":  id.ChainId,
			"contract": id.ContractAddress,
			"tokenId":  id.TokenId,
		}).Error("nftitemCache.Del failed")
		return nil
	}

	return nil
}

func (im *nftitemImpl) IncreaseViewCount(c ctx.Ctx, id nftitem.Id, count int) (int32, error) {
	res := &nftitem.NftItem{}

	if slr, err := mongoclient.MakeBsonM(id); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return 0, err
	} else if err := im.q.Increment(c, domain.TableNFTItems, slr, res, "viewed", count); err != nil {
		c.WithField("err", err).Error("q.Increment failed")
		return 0, err
	}

	key := keys.RedisKey(strconv.Itoa(int(id.ChainId)), string(id.ContractAddress), string(id.TokenId))

	if err := im.nftitemCache.Del(c, key); err != nil {
		c.WithFields(log.Fields{
			"err":      err,
			"chainId":  id.ChainId,
			"contract": id.ContractAddress,
			"tokenId":  id.TokenId,
		}).Error("nftitemCache.Del failed")
		return res.Viewed, nil
	}

	return res.Viewed, nil
}

func (im *nftitemImpl) IncreaseLikeCount(c ctx.Ctx, id nftitem.Id, count int) (int32, error) {
	res := &nftitem.NftItem{}

	id.ContractAddress = id.ContractAddress.ToLower()

	if slr, err := mongoclient.MakeBsonM(id); err != nil {
		c.WithField("err", err).Error("mongoclient.MakeBsonM failed")
		return 0, err
	} else if err := im.q.Increment(c, domain.TableNFTItems, slr, res, "liked", count); err != nil {
		c.WithField("err", err).Error("q.Increment failed")
		return 0, err
	}

	key := keys.RedisKey(strconv.Itoa(int(id.ChainId)), string(id.ContractAddress), string(id.TokenId))

	if err := im.nftitemCache.Del(c, key); err != nil {
		c.WithFields(log.Fields{
			"err":      err,
			"chainId":  id.ChainId,
			"contract": id.ContractAddress,
			"tokenId":  id.TokenId,
		}).Error("nftitemCache.Del failed")
		return res.Liked, nil
	}

	return res.Liked, nil
}

func (im *nftitemImpl) Create(c ctx.Ctx, nft *nftitem.NftItem) error {
	if err := im.q.Insert(c, domain.TableNFTItems, nft); err != nil {
		c.WithField("err", err).Error("q.Insert failed")
		return err
	}
	return nil
}

func (im *nftitemImpl) IncreaseSupply(c ctx.Ctx, id nftitem.Id, n int) error {
	var res nftitem.NftItem
	return im.q.Increment(c, domain.TableNFTItems, id, &res, "supply", n)
}

func (im *nftitemImpl) DecreaseSupply(c ctx.Ctx, id nftitem.Id, n int) error {
	return im.IncreaseSupply(c, id, -n)
}
