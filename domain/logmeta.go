package domain

import "time"

type LogMeta struct {
	BlockNumber     BlockNumber
	BlockTime       time.Time
	TxHash          TxHash
	TxIndex         uint
	LogIndex        uint
	ContractAddress Address
	MsgSender       Address
}
