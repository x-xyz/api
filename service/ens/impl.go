package ens

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	goens "github.com/wealdtech/go-ens/v3"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/service/cache"
	compoundcache "github.com/x-xyz/goapi/service/cache/compoundCache"
	"github.com/x-xyz/goapi/service/cache/provider/primitive"
	redisCache "github.com/x-xyz/goapi/service/cache/provider/redis"
	"github.com/x-xyz/goapi/service/redis"
)

type impl struct {
	client *ethclient.Client
	cache  cache.Service
}

func New(rpc string, redis redis.Service) ENS {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		panic(err)
	}
	return &impl{
		client,
		compoundcache.NewCompoundCache([]cache.Service{
			cache.New(cache.ServiceConfig{
				Ttl:   30 * time.Second,
				Pfx:   "ensPfx",
				Cache: primitive.NewPrimitive("ens", 512),
			}),
			cache.New(cache.ServiceConfig{
				Ttl:   7 * 24 * time.Hour, // cache for 1 week
				Pfx:   "ensPfx",
				Cache: redisCache.NewRedis(redis),
			}),
		}),
	}
}

func (im *impl) Resolve(ctx ctx.Ctx, name string) (domain.Address, error) {
	res := domain.Address("")
	key := keys.RedisKey("resolve", name)
	err := im.cache.GetByFunc(ctx, key, &res, func() (interface{}, error) {
		addr, err := goens.Resolve(im.client, name)
		if fmt.Sprint(err) == "unregistered name" {
			val := domain.Address("")
			return &val, nil
		}
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
			}).Error("failed to goens.Resolve")
			return nil, err
		}
		val := domain.Address(addr.String())
		return &val, nil
	})

	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to cache.GetByFunc")
		return "", err
	}

	return res, nil
}

func (im *impl) ReverseResolve(ctx ctx.Ctx, address domain.Address) (string, error) {
	res := ""
	key := keys.RedisKey("reverse-resolve", address.ToLowerStr())
	err := im.cache.GetByFunc(ctx, key, &res, func() (interface{}, error) {
		name, err := goens.ReverseResolve(im.client, common.HexToAddress(string(address)))
		if fmt.Sprint(err) == "not a resolver" {
			return ptr.String(""), nil
		}
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
			}).Error("failed to goens.ReverseResolve")
			return nil, err
		}
		return &name, nil
	})

	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to cache.GetByFunc")
		return "", err
	}

	return res, nil
}
