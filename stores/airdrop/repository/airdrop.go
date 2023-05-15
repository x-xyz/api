package repository

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/airdrop"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type airdropRepoImpl struct {
	q query.Mongo
}

func NewAirdropRepo(q query.Mongo) airdrop.AirdropRepo {
	return &airdropRepoImpl{q: q}
}

func (r *airdropRepoImpl) FindAll(ctx bCtx.Ctx, optFns ...airdrop.AirdropFindAllOptionsFunc) ([]airdrop.Airdrop, error) {
	opts, err := airdrop.GetAirdropFindAllOptions(optFns...)
	if err != nil {
		ctx.WithField("err", err).Error("airdrop.GetAirdropFindAllOptions failed")
		return nil, err
	}

	var (
		offset int    = 0
		limit  int    = 0
		sort   string = "-deadline"
		query  bson.M = bson.M{}
	)
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
	if opts.DeadlineAfter != nil {
		query["deadline"] = bson.M{"$gt": opts.DeadlineAfter}
	}

	airdrops := []airdrop.Airdrop{}
	if err := r.q.Search(ctx, domain.TableAirdrops, offset, limit, sort, query, &airdrops); err != nil {
		ctx.WithField("err", err).Error("q.Search failed")
		return nil, err
	}
	return airdrops, err
}

func (r *airdropRepoImpl) Create(ctx bCtx.Ctx, a *airdrop.Airdrop) error {
	copy := &airdrop.Airdrop{
		Name:               a.Name,
		Image:              a.Image,
		ChainId:            a.ChainId,
		ContractAddress:    a.ContractAddress.ToLower(),
		RewardTokenAddress: a.RewardTokenAddress.ToLower(),
		Type:               a.Type,
		Deadline:           a.Deadline,
	}
	if err := r.q.Insert(ctx, domain.TableAirdrops, copy); err != nil {
		ctx.WithField("err", err).Error("q.Insert failed")
		return err
	}
	return nil
}
