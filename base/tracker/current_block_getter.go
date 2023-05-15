package tracker

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

type CurrentBlockGetterCfg struct {
	Client *ethclient.Client
	ErrCh  chan error
}

type CurrentBlockGetter struct {
	client    *ethclient.Client
	mutex     sync.RWMutex
	blk       uint64
	errCh     chan error
	stoppedCh chan interface{}
}

func NewCurrentBlockGetter(cfg *CurrentBlockGetterCfg) CurrentBlockProvider {
	return &CurrentBlockGetter{
		client:    cfg.Client,
		errCh:     cfg.ErrCh,
		stoppedCh: make(chan interface{}),
	}
}

func (g *CurrentBlockGetter) BlockNumber(ctx context.Context) (uint64, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.blk, nil
}

func (g *CurrentBlockGetter) Start(ctx bCtx.Ctx) error {
	blk, err := g.client.BlockNumber(ctx)
	if err != nil {
		ctx.WithField("err", err).Error("client.BlockNumber failed")
		return err
	}
	g.blk = blk
	go g.loop(ctx)
	return nil
}

func (g *CurrentBlockGetter) Wait() {
	<-g.stoppedCh
}

func (g *CurrentBlockGetter) loop(ctx bCtx.Ctx) {
	blkCh := make(chan *types.Header)
	sub, err := g.client.SubscribeNewHead(ctx, blkCh)
	if err != nil {
		ctx.WithField("err", err).Error("client.SubscribeNewHead failed")
		g.errCh <- err
		close(g.stoppedCh)
		return
	}
	defer sub.Unsubscribe()
	for {
		select {
		case <-ctx.Done():
			close(g.stoppedCh)
			return
		case blk := <-blkCh:
			g.mutex.Lock()
			g.blk = blk.Number.Uint64()
			g.mutex.Unlock()
		case err := <-sub.Err():
			ctx.WithField("err", err).Error("sub.Err()")
			g.errCh <- err
			close(g.stoppedCh)
			return
		}
	}
}
