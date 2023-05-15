package primitive

import (
	"strconv"
	"time"

	"github.com/coocood/freecache"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/cache/provider"
)

type impl struct {
	name  string
	cache freecache.Cache
}

func NewPrimitive(name string, size int) provider.Provider {
	return &impl{name, *freecache.NewCache(size * 1024 * 1024)}
}

func (im *impl) Get(c ctx.Ctx, key string) ([]byte, time.Duration, error) {
	if val, ttl, err := im.cache.GetWithExpiration([]byte(key)); err != nil {
		if err == freecache.ErrNotFound {
			return nil, time.Duration(0), provider.ErrNotFound
		}
		c.WithField("err", err).WithField("key", key).Error("cache.Get failed")
		return nil, time.Duration(0), err
	} else {
		return val, time.Duration(ttl) * time.Second, nil
	}
}

func (im *impl) Set(c ctx.Ctx, key string, value []byte, ttl time.Duration) error {
	if err := im.cache.Set([]byte(key), value, int(ttl.Seconds())); err != nil {
		if err == freecache.ErrNotFound {
			return provider.ErrNotFound
		}
		c.WithField("err", err).WithField("key", key).Error("cache.Set failed")
		return err
	}
	return nil
}

func (im *impl) Incr(c ctx.Ctx, key string, val int) (int64, time.Duration, error) {
	v, exp, err := im.cache.GetWithExpiration([]byte(key))
	if err != nil {
		if err == freecache.ErrNotFound {
			return 0, time.Duration(0), provider.ErrNotFound
		}
		c.WithField("err", err).WithField("key", key).Error("cache.GetWithExpiration failed")
		return 0, time.Duration(0), err
	}

	i, err := strconv.ParseInt(string(v), 10, 64)
	if err != nil {
		c.WithField("err", err).WithField("key", key).Error("strconv.ParseInt failed")
		return 0, time.Duration(0), err
	}

	nv := i + int64(val)
	ttl := time.Duration(exp) * time.Second
	return nv, ttl, im.Set(c, key, []byte(strconv.FormatInt(nv, 10)), ttl)
}

func (im *impl) Del(c ctx.Ctx, key string) error {
	im.cache.Del([]byte(key))
	return nil
}
