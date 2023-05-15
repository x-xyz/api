package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ThrottledClient struct {
	*ethclient.Client
	tokens chan int
}

func NewTrottledClient(client *ethclient.Client, n int) *ThrottledClient {
	tokens := make(chan int, n)
	for i := 0; i < n; i++ {
		tokens <- i + 1
	}
	return &ThrottledClient{
		Client: client,
		tokens: tokens,
	}
}

func (c *ThrottledClient) BlockNumber(ctx context.Context) (uint64, error) {
	token := c.before(ctx)
	defer c.after(token)
	return c.Client.BlockNumber(ctx)
}

func (c *ThrottledClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	token := c.before(ctx)
	defer c.after(token)
	return c.Client.BlockByNumber(ctx, number)
}

func (c *ThrottledClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	token := c.before(ctx)
	defer c.after(token)
	return c.Client.HeaderByNumber(ctx, number)
}

func (c *ThrottledClient) FilterLogs(ctx context.Context, filter ethereum.FilterQuery) ([]types.Log, error) {
	token := c.before(ctx)
	defer c.after(token)
	return c.Client.FilterLogs(ctx, filter)
}

func (c *ThrottledClient) SubscribeFilterLogs(ctx context.Context, filter ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	token := c.before(ctx)
	defer c.after(token)
	return c.Client.SubscribeFilterLogs(ctx, filter, ch)
}

func (c *ThrottledClient) CodeAt(ctx context.Context, address common.Address, number *big.Int) ([]byte, error) {
	token := c.before(ctx)
	defer c.after(token)
	return c.Client.CodeAt(ctx, address, number)
}

func (c *ThrottledClient) CallContract(ctx context.Context, msg ethereum.CallMsg, number *big.Int) ([]byte, error) {
	token := c.before(ctx)
	defer c.after(token)
	return c.Client.CallContract(ctx, msg, number)
}

func (c *ThrottledClient) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	token := c.before(ctx)
	defer c.after(token)
	return c.Client.TransactionByHash(ctx, hash)
}

func (c *ThrottledClient) TransactionReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	token := c.before(ctx)
	defer c.after(token)
	return c.Client.TransactionReceipt(ctx, hash)
}

func (c *ThrottledClient) before(ctx context.Context) int {
	now := time.Now()
	select {
	case <-ctx.Done():
		fmt.Printf("#throttle ctx done time=%s\n", time.Since(now))
		return 0
	case token := <-c.tokens:
		fmt.Printf("#throttle token=%d len=%d time=%s\n", token, len(c.tokens), time.Since(now))
		return token
	}
}

func (c *ThrottledClient) after(token int) {
	if token != 0 {
		c.tokens <- token
	}
}
