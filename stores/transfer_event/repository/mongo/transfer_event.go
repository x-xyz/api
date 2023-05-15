package mongo

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/query"
)

type transferEventMongoRepo struct {
	m query.Mongo
}

func NewTransferEventMongoRepo(mCon query.Mongo) domain.TransferEventRepo {
	return &transferEventMongoRepo{m: mCon}
}

func (r *transferEventMongoRepo) Store(_ ctx.Ctx, _ *domain.TransferEvent) error {
	panic("not implemented") // TODO: Implement
}
