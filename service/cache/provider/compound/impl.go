package compound

import (
	"strconv"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/cache/provider"
)

type impl struct {
	layers []provider.Provider
}

// order of layers is matter, compound cache only handle forward filling
// and return immediately once cache hit
func NewCompound(layers []provider.Provider) provider.Provider {
	return &impl{layers}
}

func (im *impl) Get(c ctx.Ctx, key string) ([]byte, time.Duration, error) {
	var (
		val    []byte
		ttl    time.Duration
		err    error
		hitIdx = -1
	)

	for idx, lyr := range im.layers {
		if val, ttl, err = lyr.Get(c, key); err == provider.ErrNotFound {
			continue
		} else if err != nil {
			return nil, time.Duration(0), err
		} else {
			hitIdx = idx
			break
		}
	}

	if hitIdx == -1 {
		return nil, time.Duration(0), provider.ErrNotFound
	}

	// fill layers which missing cache
	for idx := 0; idx < hitIdx; idx++ {
		lyr := im.layers[idx]
		if err := lyr.Set(c, key, val, ttl); err != nil {
			return nil, time.Duration(0), err
		}
	}

	return val, ttl, nil
}

func (im *impl) Set(c ctx.Ctx, key string, value []byte, ttl time.Duration) error {
	for _, lyr := range im.layers {
		if err := lyr.Set(c, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// incr to last cache and fill all caches front
func (im *impl) Incr(c ctx.Ctx, key string, val int) (int64, time.Duration, error) {
	l := len(im.layers)
	last := im.layers[l-1]
	res, ttl, err := last.Incr(c, key, val)

	if err != nil {
		return 0, time.Duration(0), err
	}

	for _, lyr := range im.layers {
		if lyr == last {
			break
		}
		if err := lyr.Set(c, key, []byte(strconv.FormatInt(res, 10)), ttl); err != nil {
			return 0, time.Duration(0), err
		}
	}

	return res, ttl, err
}

func (im *impl) Del(c ctx.Ctx, key string) error {
	for _, lyr := range im.layers {
		if err := lyr.Del(c, key); err != nil {
			return err
		}
	}
	return nil
}
