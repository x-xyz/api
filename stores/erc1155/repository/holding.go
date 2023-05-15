package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/erc1155"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type holdingImpl struct {
	q query.Mongo
}

func NewHoldingRepo(q query.Mongo) erc1155.HoldingRepo {
	return &holdingImpl{q}
}

func (h *holdingImpl) FindOne(c ctx.Ctx, id erc1155.HoldingId) (*erc1155.Holding, error) {
	var holding erc1155.Holding
	if err := h.q.FindOne(c, domain.TableERC1155Holdings, id, &holding); err == query.ErrNotFound {
		return nil, domain.ErrNotFound
	} else if err != nil {
		c.WithField("err", err).Error("q.FindOne failed")
		return nil, err
	}
	return &holding, nil
}

func (h *holdingImpl) FindAll(c ctx.Ctx, opts ...erc1155.FindAllOptionsFunc) ([]*erc1155.Holding, error) {
	options, err := erc1155.GetFindAllOptions(opts...)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to GetFindAllOptions")
		return nil, err
	}
	query := bson.M{}

	if options.Owner != nil && !options.Owner.IsEmpty() {
		query["owner"] = *options.Owner
	}

	if options.Address != nil {
		query["address"] = options.Address.ToLower()
	}

	if options.NftitemId != nil {
		query["chainId"] = options.NftitemId.ChainId
		query["address"] = options.NftitemId.ContractAddress.ToLower()
		query["tokenId"] = options.NftitemId.TokenId
	}

	res := []*erc1155.Holding{}

	err = h.q.Search(c, domain.TableERC1155Holdings, 0, 0, "_id", query, &res)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to Search")
		return nil, err
	}

	return res, nil
}

func (h *holdingImpl) Create(c ctx.Ctx, value erc1155.Holding) error {
	if err := h.q.Insert(c, domain.TableERC1155Holdings, value); err != nil {
		c.WithField("err", err).Error("q.Insert failed")
		return err
	}
	return nil
}

func (h *holdingImpl) Delete(c ctx.Ctx, id erc1155.HoldingId) error {
	if err := h.q.Remove(c, domain.TableERC1155Holdings, id); err != nil && err != query.ErrNotFound {
		return err
	} else if err == query.ErrNotFound {
		return domain.ErrNotFound
	}
	return nil
}

func (h *holdingImpl) Increment(c ctx.Ctx, id erc1155.HoldingId, value int64) (*erc1155.Holding, error) {
	var res erc1155.Holding
	err := h.q.Increment(c, domain.TableERC1155Holdings, id, &res, "balance", value)
	return &res, err
}

func (h *holdingImpl) CountUniqueOwner(c ctx.Ctx, chainId domain.ChainId, address domain.Address) (int64, error) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"chainId": chainId, "address": address, "balance": bson.M{"$gt": 0}}}},
		{{"$group", bson.M{"_id": "$owner"}}},
		{{"$count", "numOwner"}},
	}
	iter, close, err := h.q.Pipe(c, domain.TableERC1155Holdings, pipeline)
	if err != nil {
		c.WithField("err", err).Error("q.Pipe failed")
		return 0, err
	}
	defer close()

	var result []struct {
		NumOwner int64 `bson:"numOwner"`
	}
	if err := iter.All(c, &result); err != nil {
		c.WithField("err", err).Error("iter.Cursor.All failed")
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}
	return result[0].NumOwner, nil
}
