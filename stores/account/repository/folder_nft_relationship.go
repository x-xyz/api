package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type folderNftRelationshipImpl struct {
	query query.Mongo
}

func NewFolderNftRelationshipRepo(query query.Mongo) account.FolderNftRelationshipRepo {
	return &folderNftRelationshipImpl{query}
}

func (im *folderNftRelationshipImpl) Insert(ctx ctx.Ctx, relation *account.FolderNftRelationship) error {
	if err := im.query.Insert(ctx, domain.TableFolderNftRelationships, relation); err != nil {
		ctx.WithFields(log.Fields{
			"relationship": *relation,
			"err":          err,
		}).Error("failed to Insert FolderNftRelationship")
		return err
	}
	return nil
}

func makeRelationFindQuery(opts account.RelationsQueryOptions) bson.M {
	query := bson.M{}
	if opts.FolderId != nil {
		query["folderId"] = *opts.FolderId
	}

	if opts.FolderIds != nil {
		query["folderId"] = bson.M{"$in": *opts.FolderIds}
	}

	if opts.NftitemId != nil {
		query["chainId"] = opts.NftitemId.ChainId
		query["contractAddress"] = opts.NftitemId.ContractAddress.ToLower()
		query["tokenId"] = opts.NftitemId.TokenId
	}

	return query
}

func (im *folderNftRelationshipImpl) GetAllRelations(ctx ctx.Ctx, opts ...account.RelationsQueryOptionsFunc) ([]*account.FolderNftRelationship, error) {
	option, err := account.ParseRelationsQueryOptionFunc(opts...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to ParseGetAllRelationsOptionFunc")
		return nil, err
	}

	query := makeRelationFindQuery(option)

	res := []*account.FolderNftRelationship{}
	if err := im.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "index", query, &res); err != nil {
		ctx.WithFields(log.Fields{
			"query": query,
			"err":   err,
		}).Error("failed to Search")
		return nil, err
	}
	return res, nil
}

func (im *folderNftRelationshipImpl) DeleteAllRelationsByFolderID(ctx ctx.Ctx, folderID string) error {
	q := bson.M{"folderId": folderID}
	if _, err := im.query.RemoveAll(ctx, domain.TableFolderNftRelationships, q); err != nil {
		ctx.WithFields(log.Fields{
			"folderId": folderID,
			"err":      err,
		}).Error("failed to RemoveAll")
		return err
	}
	return nil
}

func (im *folderNftRelationshipImpl) DeleteAllRelationsByNftitem(ctx ctx.Ctx, nftID nftitem.Id) error {
	q := bson.M{
		"chainId":         nftID.ChainId,
		"contractAddress": nftID.ContractAddress,
		"tokenId":         nftID.TokenId,
	}
	if _, err := im.query.RemoveAll(ctx, domain.TableFolderNftRelationships, q); err != nil {
		ctx.WithFields(log.Fields{
			"nftID": nftID,
			"err":   err,
		}).Error("failed to RemoveAll")
		return err
	}
	return nil
}

func (im *folderNftRelationshipImpl) DeleteAll(ctx ctx.Ctx, options ...account.RelationsQueryOptionsFunc) error {
	opts, err := account.ParseRelationsQueryOptionFunc(options...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to ParseGetAllRelationsOptionFunc")
		return err
	}

	query := makeRelationFindQuery(opts)

	_, err = im.query.RemoveAll(ctx, domain.TableFolderNftRelationships, query)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to query.RemoveAll")
		return err
	}

	return nil
}

func (im *folderNftRelationshipImpl) AddNftitemsToFolder(ctx ctx.Ctx, items []nftitem.Id, folderId string) error {
	ops := []query.UpsertOp{}
	for i, it := range items {
		ops = append(ops, query.UpsertOp{
			Selector: bson.M{
				"chainId":         it.ChainId,
				"contractAddress": it.ContractAddress,
				"tokenId":         it.TokenId,
				"folderId":        folderId,
			},
			Updater: bson.M{
				"chainId":         it.ChainId,
				"contractAddress": it.ContractAddress,
				"tokenId":         it.TokenId,
				"folderId":        folderId,
				"index":           i,
			},
		})
	}

	if _, _, err := im.query.BulkUpsert(ctx, domain.TableFolderNftRelationships, ops); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to BulkUpsert")
		return err
	}

	return nil
}

func (im *folderNftRelationshipImpl) MoveNftitems(ctx ctx.Ctx, items []nftitem.Id, fromFolderId, toFolderId string) error {
	// return if no item need to be moved
	if len(items) == 0 {
		return nil
	}

	// query last index
	lastRelation := []account.FolderNftRelationship{}
	if err := im.query.Search(ctx, domain.TableFolderNftRelationships, 0, 1, "-index", bson.M{"folderId": toFolderId}, &lastRelation); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to FindOne")
		return err
	}
	lastIndex := 0
	if len(lastRelation) > 0 {
		lastIndex = lastRelation[0].Index
	}

	ops := []query.UpsertOp{}
	for _, it := range items {
		lastIndex += 1
		ops = append(ops, query.UpsertOp{
			Selector: bson.M{
				"chainId":         it.ChainId,
				"contractAddress": it.ContractAddress,
				"tokenId":         it.TokenId,
				"$or": bson.A{
					bson.M{"folderId": fromFolderId},
					bson.M{"folderId": toFolderId},
				},
			},
			Updater: bson.M{
				"chainId":         it.ChainId,
				"contractAddress": it.ContractAddress,
				"tokenId":         it.TokenId,
				"folderId":        toFolderId,
				"index":           lastIndex,
			},
		})
	}

	if _, _, err := im.query.BulkUpsert(ctx, domain.TableFolderNftRelationships, ops); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to BulkUpsert")
		return err
	}

	return nil
}

func (im *folderNftRelationshipImpl) Count(ctx ctx.Ctx, opts ...account.RelationsQueryOptionsFunc) (int, error) {
	option, err := account.ParseRelationsQueryOptionFunc(opts...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to ParseGetAllRelationsOptionFunc")
		return 0, err
	}

	query := bson.M{}
	if option.FolderId != nil {
		query["folderId"] = *option.FolderId
	}

	count, err := im.query.Count(ctx, domain.TableFolderNftRelationships, query)
	if err != nil {
		ctx.WithFields(log.Fields{
			"query": query,
			"err":   err,
		}).Error("failed to Count")
		return 0, err
	}
	return count, nil
}
