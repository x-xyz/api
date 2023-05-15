package repository

import (
	"github.com/x-xyz/goapi/base/ctx"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type vexRepoImpl struct {
	q query.Mongo
}

func NewVexFeeDistributionHistoryRepo(q query.Mongo) domain.VexFeeDistributionHistoryRepo {
	return &vexRepoImpl{q: q}
}

func (r *vexRepoImpl) Create(ctx bCtx.Ctx, h *domain.VexFeeDistributionHistory) error {
	if err := r.q.Insert(ctx, domain.TableVexFeeDistributionHistory, h); err != nil {
		ctx.WithFields(log.Fields{
			"history": h,
			"err":     err,
		}).Error("q.Insert failed")
		return err
	}
	return nil
}

func (r *vexRepoImpl) FindLatest(ctx ctx.Ctx, limit int) ([]domain.VexFeeDistributionHistory, error) {
	history := []domain.VexFeeDistributionHistory{}
	sortBy := "-time"
	qry := bson.M{
		"time": bson.M{"$exists": true},
	}
	if err := r.q.Search(ctx, domain.TableVexFeeDistributionHistory, 0, limit, sortBy, qry, &history); err != nil {
		ctx.WithField("err", err).Error("q.Search failed")
		return nil, err
	}
	return history, nil
}
