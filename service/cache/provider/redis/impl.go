package redis

import (
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/cache/provider"
	"github.com/x-xyz/goapi/service/redis"
)

type impl struct {
	redis redis.Service
}

func NewRedis(redis redis.Service) provider.Provider {
	return &impl{redis}
}

func (im *impl) Get(c ctx.Ctx, key string) ([]byte, time.Duration, error) {
	if val, err := im.redis.Get(c, key); err != nil {
		if err == redis.ErrNotFound {
			return nil, time.Duration(0), provider.ErrNotFound
		}
		c.WithField("err", err).WithField("key", key).Error("redis.Get failed")
		return nil, time.Duration(0), err
	} else if ttl, err := im.redis.TTL(c, key); err != nil {
		c.WithField("err", err).WithField("key", key).Error("redis.TTL failed")
		return nil, time.Duration(0), err
	} else {
		return val, time.Duration(ttl) * time.Second, nil
	}
}

func (im *impl) Set(c ctx.Ctx, key string, value []byte, ttl time.Duration) error {
	if err := im.redis.Set(c, key, value, ttl); err != nil {
		c.WithField("err", err).WithField("key", key).Error("redis.Set failed")
		return err
	}
	return nil
}

func (im *impl) Incr(c ctx.Ctx, key string, val int) (int64, time.Duration, error) {
	// to perform same behavior with localecache
	if exists, err := im.redis.Exists(c, key); err != nil {
		c.WithField("err", err).WithField("key", key).Error("redis.Exists failed")
		return 0, time.Duration(0), err
	} else if !exists {
		return 0, time.Duration(0), provider.ErrNotFound
	} else if res, err := im.redis.Incrby(c, key, val); err != nil {
		c.WithField("err", err).WithField("key", key).Error("redis.Incrby failed")
		return 0, time.Duration(0), err
	} else if ttl, err := im.redis.TTL(c, key); err != nil {
		c.WithField("err", err).WithField("key", key).Error("redis.TTL failed")
		return 0, time.Duration(0), err
	} else {
		return res, time.Duration(ttl) * time.Second, nil
	}
}

func (im *impl) Del(c ctx.Ctx, key string) error {
	if _, err := im.redis.Del(c, key); err != nil {
		c.WithField("err", err).WithField("key", key).Error("redis.Del failed")
		return err
	}
	return nil
}
