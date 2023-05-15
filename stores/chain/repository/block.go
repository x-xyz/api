package repository

import (
	"errors"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/chain"
	"github.com/x-xyz/goapi/service/query"
)

type blockRepo struct {
	q query.Mongo
}

func NewBlockRepo(q query.Mongo) chain.BlockRepo {
	return &blockRepo{
		q: q,
	}
}

func (r *blockRepo) Create(ctx bCtx.Ctx, b *chain.Block) error {
	if err := r.q.Insert(ctx, domain.TableBlocks, b); err != nil {
		ctx.WithField("err", err).Error("q.Insert failed")
		return err
	}
	return nil
}

func (r *blockRepo) Upsert(ctx bCtx.Ctx, b *chain.Block) error {
	if err := r.q.Upsert(ctx, domain.TableBlocks, b.ToId(), b); err != nil {
		ctx.WithField("err", err).Error("q.Upsertfailed")
		return err
	}
	return nil
}

func (r *blockRepo) FindOne(ctx bCtx.Ctx, id *chain.BlockId) (*chain.Block, error) {
	b := &chain.Block{}
	if err := r.q.FindOne(ctx, domain.TableBlocks, id, b); err != nil {
		if !errors.Is(err, query.ErrNotFound) {
			ctx.WithField("err", err).Error("q.FindOne failed")
		}
		return nil, err
	}
	return b, nil
}
