package chainlink

import (
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/x-xyz/goapi/base/abi"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/service/cache"
	"github.com/x-xyz/goapi/service/cache/provider/primitive"
	"github.com/x-xyz/goapi/service/chain"
)

type impl struct {
	chainClient chain.Client
	cache       cache.Service
}

func New(chainClient chain.Client) Chainlink {
	return &impl{
		chainClient: chainClient,
		cache: cache.New(cache.ServiceConfig{
			Ttl:   time.Hour,
			Pfx:   "chainlink_cache",
			Cache: primitive.NewPrimitive("chainlink_cache", 32),
		}),
	}
}

func (im *impl) GetLatestAnswer(c ctx.Ctx, chainId domain.ChainId, address domain.Address) (*big.Int, error) {
	var res big.Int

	key := keys.RedisKey(strconv.Itoa(int(chainId)), string(address), "latest")

	if err := im.cache.GetByFunc(c, key, &res, func() (interface{}, error) {
		if res, err := im.getLatestAnswer(c, chainId, address, nil); err != nil {
			c.WithFields(log.Fields{
				"err":     err,
				"chainId": chainId,
				"address": address,
			}).Error("getLatestAnswer failed")
			return nil, err
		} else {
			return res, nil
		}
	}); err != nil {
		c.WithFields(log.Fields{
			"err":     err,
			"chainId": chainId,
			"address": address,
		}).Error("cache.GetByFunc failed")
		return nil, err
	}

	return &res, nil
}

func (im *impl) GetLatestAnswerAt(c ctx.Ctx, chainId domain.ChainId, address domain.Address, blk *big.Int) (*big.Int, error) {
	var res big.Int

	key := keys.RedisKey(strconv.Itoa(int(chainId)), string(address), blk.String())

	if err := im.cache.GetByFunc(c, key, &res, func() (interface{}, error) {
		if res, err := im.getLatestAnswer(c, chainId, address, blk); err != nil {
			c.WithFields(log.Fields{
				"err":     err,
				"chainId": chainId,
				"address": address,
				"blk":     blk,
			}).Error("getLatestAnswerAt failed")
			return nil, err
		} else {
			return res, nil
		}
	}); err != nil {
		c.WithFields(log.Fields{
			"err":     err,
			"chainId": chainId,
			"address": address,
			"blk":     blk,
		}).Error("cache.GetByFunc failed")
		return nil, err
	}

	return &res, nil
}

func (im *impl) getLatestAnswer(c ctx.Ctx, chainId domain.ChainId, address domain.Address, blk *big.Int) (*big.Int, error) {
	feedAddr := common.HexToAddress(string(address))

	res, err := im.chainClient.Call(c, int32(chainId), feedAddr, blk, abi.ChainlinkFeedABI, "latestAnswer")
	if err != nil {
		c.WithFields(log.Fields{
			"err":     err,
			"chainId": chainId,
			"address": address,
		}).Error("chainClient.Call failed")
		return nil, err
	}

	return res[0].(*big.Int), nil
}
