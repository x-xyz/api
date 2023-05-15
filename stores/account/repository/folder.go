package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type folderReopImpl struct {
	query query.Mongo
}

func NewFolderRepo(query query.Mongo) account.FolderRepo {
	return &folderReopImpl{query}
}

func (im *folderReopImpl) Insert(ctx ctx.Ctx, f *account.Folder) error {
	f.Owner = f.Owner.ToLower()
	if err := im.query.Insert(ctx, domain.TableFolders, f); err != nil {
		ctx.WithFields(log.Fields{
			"folder": *f,
			"err":    err,
		}).Error("insert folder failed")
		return err
	}

	return nil
}

func (im *folderReopImpl) Get(ctx ctx.Ctx, Id string) (*account.Folder, error) {
	q := bson.M{"id": Id}
	res := account.Folder{}
	if err := im.query.FindOne(ctx, domain.TableFolders, q, &res); err != nil {
		ctx.WithFields(log.Fields{
			"id":  Id,
			"err": err,
		}).Error("failed to FindOne")
		return nil, err
	}
	return &res, nil
}

func (im *folderReopImpl) GetFolders(ctx ctx.Ctx, opts ...account.GetFoldersOptionsFunc) ([]*account.Folder, error) {
	opt, err := account.ParseGetFoldersOptionFunc(opts...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to ParseGetFoldersOptionFunc")
		return nil, err
	}

	offset := int(0)
	limit := int(0)

	if opt.Offset != nil {
		offset = int(*opt.Offset)
	}

	if opt.Limit != nil {
		limit = int(*opt.Limit)
	}

	q := bson.M{}

	if opt.Owner != nil {
		q["owner"] = opt.Owner.ToLowerStr()
	}

	if opt.IsBuiltIn != nil {
		q["isBuiltIn"] = *opt.IsBuiltIn
	}

	if opt.IsPrivate != nil {
		q["isPrivate"] = *opt.IsPrivate
	}

	if len(q) == 0 {
		q["_id"] = bson.M{"$exists": true}
	}

	res := []*account.Folder{}
	if err := im.query.Search(ctx, domain.TableFolders, offset, limit, "_id", q, &res); err != nil {
		ctx.WithFields(log.Fields{
			"query": q,
			"err":   err,
		}).Error("search folder failed")
		return nil, err
	}
	return res, nil
}

func (im *folderReopImpl) Update(ctx ctx.Ctx, Id string, patchable *account.FolderUpdater) error {
	q := bson.M{"id": Id}
	updateBson, err := mongoclient.MakeBsonM(patchable)
	if err != nil {
		ctx.WithFields(log.Fields{
			"patchable": *patchable,
			"err":       err,
		})
		return err
	}
	if err := im.query.Patch(ctx, domain.TableFolders, q, updateBson); err != nil {
		ctx.WithFields(log.Fields{
			"id":  Id,
			"err": err,
		}).Error("patch folder failed")
		return err
	}
	return nil
}

func (im *folderReopImpl) Delete(ctx ctx.Ctx, Id string) error {
	q := bson.M{"id": Id}
	if err := im.query.Remove(ctx, domain.TableFolders, q); err != nil {
		ctx.WithFields(log.Fields{
			"id":  Id,
			"err": err,
		}).Error("remove folder failed")
		return err
	}

	return nil
}
