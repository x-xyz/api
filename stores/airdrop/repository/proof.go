package repository

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/airdrop"
	"github.com/x-xyz/goapi/service/query"
)

type proofRepoImpl struct {
	q query.Mongo
}

func NewProofRepo(q query.Mongo) airdrop.ProofRepo {
	return &proofRepoImpl{q: q}
}

func (r *proofRepoImpl) FindAll(ctx bCtx.Ctx, optFns ...airdrop.ProofFindAllOptionsFunc) ([]airdrop.Proof, error) {
	opts, err := airdrop.GetProofFindAllOptions(optFns...)
	if err != nil {
		ctx.WithField("err", err).Error("proof.GetProofFindAllOptions failed")
		return nil, err
	}

	var (
		offset int    = 0
		limit  int    = 0
		sort   string = "_id"
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

	query, err := mongoclient.MakeBsonM(opts)
	if err != nil {
		ctx.WithField("err", err).Error("MakeBsonM failed")
		return nil, err
	}

	proofs := []airdrop.Proof{}
	if err := r.q.Search(ctx, domain.TableProofs, offset, limit, sort, query, &proofs); err != nil {
		ctx.WithField("err", err).Error("q.Search failed")
		return nil, err
	}
	return proofs, err
}

func (r *proofRepoImpl) Create(ctx bCtx.Ctx, a *airdrop.Proof) error {
	copy := &airdrop.Proof{
		ChainId:         a.ChainId,
		ContractAddress: a.ContractAddress.ToLower(),
		Claimer:         a.Claimer.ToLower(),
		Round:           a.Round,
		Amount:          a.Amount,
		Proof:           a.Proof,
	}
	if err := r.q.Insert(ctx, domain.TableProofs, copy); err != nil {
		ctx.WithField("err", err).Error("q.Insert failed")
		return err
	}
	return nil
}
