package chain

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type Block struct {
	ChainId domain.ChainId     `bson:"chainId"`
	Hash    domain.BlockHash   `bson:"hash"`
	Number  domain.BlockNumber `bson:"number"`
	Time    time.Time          `bson:"time"`
}

func (b *Block) ToId() *BlockId {
	return &BlockId{
		ChainId: b.ChainId,
		Number:  b.Number,
	}
}

type BlockId struct {
	ChainId domain.ChainId     `bson:"chainId"`
	Number  domain.BlockNumber `bson:"number"`
}

type BlockRepo interface {
	Create(ctx.Ctx, *Block) error
	Upsert(ctx.Ctx, *Block) error
	FindOne(ctx.Ctx, *BlockId) (*Block, error)
}

type BlockUseCase interface {
	Create(ctx.Ctx, *Block) error
	Upsert(ctx.Ctx, *Block) error
	FindOne(ctx.Ctx, *BlockId) (*Block, error)
}
