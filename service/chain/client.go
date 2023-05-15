package chain

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
)

var ErrUnsupportedChain = errors.New("unsupported chain")

type ClientCfg struct {
	RpcUrls        map[int32]string
	ArchiveRpcUrls map[int32]string
}

type Client interface {
	Call(bCtx.Ctx, int32, common.Address, *big.Int, abi.ABI, string, ...interface{}) ([]interface{}, error)
}

type clientImpl struct {
	clients        map[int32]*ethclient.Client
	archiveClients map[int32]*ethclient.Client
}

func NewClient(ctx bCtx.Ctx, cfg *ClientCfg) (Client, error) {
	var (
		anyerr error
	)
	clients := make(map[int32]*ethclient.Client)
	for chainId, url := range cfg.RpcUrls {
		client, err := ethclient.DialContext(ctx, url)
		if err != nil {
			anyerr = err
			ctx.WithFields(log.Fields{
				"err":     err,
				"chainId": chainId,
				"url":     url,
			}).Warn("failed to dial rpc")
			// soft warning, still let the server start
			continue
		}
		clients[chainId] = client
	}
	archiveClients := make(map[int32]*ethclient.Client)
	for chainId, url := range cfg.ArchiveRpcUrls {
		client, err := ethclient.DialContext(ctx, url)
		if err != nil {
			anyerr = err
			ctx.WithFields(log.Fields{
				"err":     err,
				"chainId": chainId,
				"url":     url,
			}).Warn("failed to dial rpc")
			// soft warning, still let the server start
			continue
		}
		archiveClients[chainId] = client
	}
	return &clientImpl{
		clients:        clients,
		archiveClients: archiveClients,
	}, anyerr
}

func (c *clientImpl) Call(ctx bCtx.Ctx, chainId int32, addr common.Address, blk *big.Int, _abi abi.ABI, method string, params ...interface{}) ([]interface{}, error) {
	var (
		client *ethclient.Client
		ok     bool
	)
	if blk == nil {
		client, ok = c.clients[chainId]
	} else {
		client, ok = c.archiveClients[chainId]
	}
	if !ok {
		return nil, ErrUnsupportedChain
	}

	data, err := _abi.Pack(method, params...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"method": method,
			"params": params,
			"err":    err,
		}).Error("abi.Pack failed")
		return nil, err
	}
	msg := ethereum.CallMsg{
		To:   &addr,
		Data: data,
	}
	res, err := client.CallContract(ctx, msg, blk)
	if err != nil {
		ctx.WithField("err", err).Error("client.CallContract failed")
		return nil, err
	}
	unpacked, err := _abi.Unpack(method, res)
	if err != nil {
		ctx.WithField("err", err).Error("abi.Unpack failed")
		return nil, err
	}
	return unpacked, nil
}
