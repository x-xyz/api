package usecase

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/airdrop"
)

type proofUseCaseImpl struct {
	repo airdrop.ProofRepo
}

func NewProofUseCase(repo airdrop.ProofRepo) airdrop.ProofUseCase {
	return &proofUseCaseImpl{repo: repo}
}

func (r *proofUseCaseImpl) FindAll(ctx bCtx.Ctx, optFns ...airdrop.ProofFindAllOptionsFunc) ([]airdrop.Proof, error) {
	proofs, err := r.repo.FindAll(ctx, optFns...)
	if err != nil {
		ctx.WithField("err", err).Error("repo.FindAll failed")
		return nil, err
	}
	return proofs, nil
}

func (r *proofUseCaseImpl) Create(ctx bCtx.Ctx, a *airdrop.Proof) error {
	if err := r.repo.Create(ctx, a); err != nil {
		ctx.WithField("err", err).Error("repo.Create failed")
		return err
	}
	return nil
}
