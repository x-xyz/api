package repository

import (
	"errors"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/ip"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type impl struct {
	query query.Mongo
}

func New(query query.Mongo) ip.Repo {
	return &impl{query}
}

func (im *impl) Insert(ctx ctx.Ctx, listing *ip.IPListing) error {
	listing.ContractAddress = listing.ContractAddress.ToLower()
	listing.Owner = listing.Owner.ToLower()

	err := im.query.Insert(ctx, domain.TableIpListings, listing)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":     err,
			"listing": listing,
		}).Error("failed to query.Insert")
		return err
	}
	return nil
}

func (im *impl) FindAll(ctx ctx.Ctx, options ...ip.FindAllOptionsFunc) ([]*ip.IPListing, error) {
	opts, err := ip.GetFindAllOptions(options...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to ip.GetFindAllOptions")
		return nil, err
	}

	query := bson.M{}

	if opts.IsIpOwner != nil {
		query["isIpOwner"] = *opts.IsIpOwner
	}

	if opts.ChainId != nil && len(opts.ContractAddresses) > 0 {
		query["chainId"] = *opts.ChainId
		query["contractAddress"] = bson.M{"$in": opts.ContractAddresses}
	}

	res := []*ip.IPListing{}
	err = im.query.Search(ctx, domain.TableIpListings, 0, 0, "-createdAt", query, &res)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"query": query,
		}).Error("failed to query.Search")
		return nil, err
	}
	return res, nil
}

func (im *impl) FindOne(ctx ctx.Ctx, id string) (*ip.IPListing, error) {
	res := ip.IPListing{}
	err := im.query.FindOne(ctx, domain.TableIpListings, bson.M{"id": id}, &res)
	if errors.Is(err, query.ErrNotFound) {
		return nil, domain.ErrNotFound
	} else if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to query.FindOne")
		return nil, err
	}
	return &res, nil
}

func (im *impl) Delete(ctx ctx.Ctx, id string) error {
	err := im.query.Remove(ctx, domain.TableIpListings, bson.M{"id": id})
	if errors.Is(err, query.ErrNotFound) {
		return domain.ErrNotFound
	} else if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("failed to query.Remove")
		return err
	}
	return nil
}
