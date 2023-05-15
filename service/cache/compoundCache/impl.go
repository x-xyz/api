package compoundcache

import (
	"reflect"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/cache"
)

type impl struct {
	layers []cache.Service
}

func NewCompoundCache(layers []cache.Service) cache.Service {
	return &impl{
		layers: layers,
	}
}

func (im *impl) GetByFunc(c ctx.Ctx, key string, container interface{}, getter cache.OneTimeGetter) error {
	err := im.Get(c, key, container)
	if err != nil && err != cache.ErrNotFound {
		c.WithField("err", err).WithField("key", key).Error("Get failed")
		return err
	} else if err == nil {
		// hit cache, early return
		return nil
	}

	// no cache, get and fill cache
	val, err := getter()
	if err != nil {
		// c.WithField("err", err).WithField("key", key).Info("GetByFunc getter failed")
		return err
	}

	err = im.Set(c, key, val)
	if err != nil {
		c.WithField("err", err).WithField("key", key).Error("Set failed")
	}

	reflect.ValueOf(container).Elem().Set(reflect.ValueOf(val).Elem())

	return nil
}

func (im *impl) Get(c ctx.Ctx, key string, container interface{}) error {
	var (
		err    error
		hitIdx = -1
	)

	for idx, lyr := range im.layers {
		if err = lyr.Get(c, key, container); err == cache.ErrNotFound {
			continue
		} else if err != nil {
			return err
		} else {
			hitIdx = idx
			break
		}
	}

	if hitIdx == -1 {
		return cache.ErrNotFound
	}

	// fill layers which missing cache
	for idx := 0; idx < hitIdx; idx++ {
		lyr := im.layers[idx]
		if err := lyr.Set(c, key, container); err != nil {
			return err
		}
	}

	return nil
}

func (im *impl) Set(c ctx.Ctx, key string, value interface{}) error {
	for _, lyr := range im.layers {
		if err := lyr.Set(c, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (im *impl) Del(c ctx.Ctx, key string) error {
	for _, lyr := range im.layers {
		if err := lyr.Del(c, key); err != nil {
			return err
		}
	}
	return nil
}
