package usecase

import (
	"github.com/x-xyz/goapi/domain/apecoinstaking"

	bCtx "github.com/x-xyz/goapi/base/ctx"
)

type uc struct {
	repo apecoinstaking.Repo
}

func New(repo apecoinstaking.Repo) apecoinstaking.UseCase {
	return &uc{repo}
}

func (u *uc) Get(ctx bCtx.Ctx, id apecoinstaking.Id) (*apecoinstaking.ApecoinStaking, error) {
	s, err := u.repo.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (u *uc) Upsert(ctx bCtx.Ctx, s *apecoinstaking.ApecoinStaking) error {
	err := u.repo.Upsert(ctx, s)
	if err != nil {
		return err
	}
	return nil
}
