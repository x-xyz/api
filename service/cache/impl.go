package cache

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/service/cache/provider"
)

type impl struct {
	ttl         time.Duration
	pfx         string
	cache       provider.Provider
	serialize   Serializer
	deserialize Deserializer
}

func New(config ServiceConfig) Service {
	if reflect.ValueOf(config.Serialize).IsNil() {
		config.Serialize = json.Marshal
	}

	if reflect.ValueOf(config.Deserialize).IsNil() {
		config.Deserialize = json.Unmarshal
	}

	return &impl{
		ttl:         config.Ttl,
		pfx:         config.Pfx,
		cache:       config.Cache,
		serialize:   config.Serialize,
		deserialize: config.Deserialize,
	}
}

func (im *impl) GetByFunc(c ctx.Ctx, key string, container interface{}, getter OneTimeGetter) error {
	err := im.Get(c, key, container)
	if err != nil && err != ErrNotFound {
		c.WithField("err", err).WithField("key", key).Error("Get failed")
		return err
	} else if err == nil {
		// hit cache, early return
		return nil
	}

	// no cache, get and fill cache
	val, err := getter()
	if err != nil {
		c.WithField("err", err).WithField("key", key).Error("GetByFunc getter failed")
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
	key = keys.RedisKey(im.pfx, key)

	if val, _, err := im.cache.Get(c, key); err == provider.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		c.WithField("err", err).WithField("key", key).Error("cache.Get failed")
		return err
	} else if err := im.deserialize(val, container); err != nil {
		c.WithField("err", err).WithField("key", key).Error("deserialize failed")
		return err
	}

	return nil
}

func (im *impl) Set(c ctx.Ctx, key string, value interface{}) error {
	key = keys.RedisKey(im.pfx, key)

	if val, err := im.serialize(value); err != nil {
		c.WithField("err", err).WithField("key", key).Error("serialize failed")
		return err
	} else if err := im.cache.Set(c, key, val, im.ttl); err != nil {
		c.WithField("err", err).WithField("key", key).Error("cache.Set failed")
		return err
	}

	return nil
}

func (im *impl) Del(c ctx.Ctx, key string) error {
	key = keys.RedisKey(im.pfx, key)

	if err := im.cache.Del(c, key); err != nil {
		c.WithField("err", err).WithField("key", key).Error("cache.Del failed")
		return err
	}

	return nil
}
