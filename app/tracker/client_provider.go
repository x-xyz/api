package main

import (
	"github.com/ethereum/go-ethereum/ethclient"

	bCtx "github.com/x-xyz/goapi/base/ctx"
)

type clientProvider struct {
	url    string
	limit  int
	count  int
	client *ethclient.Client
}

func newClientProvider(ctx bCtx.Ctx, limit int, url string) *clientProvider {
	client, err := ethclient.DialContext(ctx, url)
	if err != nil {
		ctx.WithField("err", err).Panic("ethclient.Dail failed")
	}
	return &clientProvider{url: url, limit: limit, client: client}
}

func (p *clientProvider) consume(ctx bCtx.Ctx) *ethclient.Client {
	p.count++
	if p.count%p.limit == 0 {
		client, err := ethclient.DialContext(ctx, p.url)
		if err != nil {
			ctx.WithField("err", err).Panic("ethclient.Dail failed")
		}
		p.client = client
	}
	return p.client
}
