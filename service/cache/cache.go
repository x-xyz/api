package cache

import (
	"errors"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/cache/provider"
)

var (
	ErrNotFound = errors.New("Cache not found")
)

type OneTimeGetter func() (interface{}, error)

type Serializer func(interface{}) ([]byte, error)

type Deserializer func([]byte, interface{}) error

// high order cache service
type Service interface {
	GetByFunc(c ctx.Ctx, key string, container interface{}, getter OneTimeGetter) error
	Get(c ctx.Ctx, key string, container interface{}) error
	Set(c ctx.Ctx, key string, value interface{}) error
	Del(c ctx.Ctx, key string) error
}

type ServiceConfig struct {
	Ttl         time.Duration
	Pfx         string
	Cache       provider.Provider
	Serialize   Serializer
	Deserialize Deserializer
}
