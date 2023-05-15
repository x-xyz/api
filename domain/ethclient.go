package domain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// just using go-ethereum/ethclient
type EthClientRepo interface {
	BlockNumber(context.Context) (uint64, error)
	BlockByNumber(context.Context, *big.Int) (*types.Block, error)
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
	FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(context.Context, ethereum.FilterQuery, chan<- types.Log) (ethereum.Subscription, error)
	CodeAt(context.Context, common.Address, *big.Int) ([]byte, error)
	CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error)
	TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)
	TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error)
}
