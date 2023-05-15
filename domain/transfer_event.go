package domain

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
)

type TransferEvent struct {
	ChainId         int64     `bson:"chain_id"`
	ContractAddress string    `bson:"contract_address"`
	BlockNumber     uint64    `bson:"block_number"`
	BlockHash       string    `bson:"block_hash"`
	BlockTime       time.Time `bson:"block_time"`
	TxHash          string    `bson:"tx_hash"`
	LogIndex        uint      `bson:"log_index"`
	From            string    `bson:"from"`
	To              string    `bson:"to"`
	TokenId         uint64    `bson:"token_id"`
}

type TransferEventRepo interface {
	Store(ctx.Ctx, *TransferEvent) error
}

type TransferEventUseCase interface {
	Store(ctx.Ctx, *TransferEvent) error
}
