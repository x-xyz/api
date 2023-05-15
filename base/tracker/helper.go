package tracker

import (
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/x-xyz/goapi/domain"
)

type Hexable interface {
	Hex() string
}

func ToLowerHexStr(h Hexable) string {
	return strings.ToLower(h.Hex())
}

func toDomainAddress(h Hexable) domain.Address {
	return domain.Address(ToLowerHexStr(h))
}

func toOrderHash(h Hexable) domain.OrderHash {
	return domain.OrderHash(ToLowerHexStr(h))
}

func toMarketplaceHash(h Hexable) domain.MarketplaceHash {
	return domain.MarketplaceHash(ToLowerHexStr(h))
}

func toLogMeta(l *logWithBlockTime) *domain.LogMeta {
	return &domain.LogMeta{
		BlockNumber:     domain.BlockNumber(l.BlockNumber),
		BlockTime:       l.blockTime,
		TxHash:          domain.TxHash(ToLowerHexStr(l.TxHash)),
		TxIndex:         l.TxIndex,
		LogIndex:        l.Index,
		ContractAddress: domain.Address(ToLowerHexStr(l.Address)),
		MsgSender:       l.msgSender,
	}
}

type logWithBlockTime struct {
	types.Log
	blockTime time.Time
	msgSender domain.Address
}
