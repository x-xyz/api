package cache

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/service/cache/provider"
	"github.com/x-xyz/goapi/service/cache/provider/primitive"
)

var (
	mockCtx = ctx.Background()
)

type value struct {
	Value string `json:"value"`
}

type testsuite struct {
	suite.Suite
	im    *impl
	cache provider.Provider
}

func (ts *testsuite) SetupTest() {
	ts.cache = primitive.NewPrimitive("test", 64)
	ts.im = New(ServiceConfig{
		Ttl:   time.Second,
		Pfx:   "testing",
		Cache: ts.cache,
	}).(*impl)
}

func Test(t *testing.T) {
	suite.Run(t, new(testsuite))
}

func (ts *testsuite) TestGet() {
	var (
		k = "key"
		v = value{"value"}
		c = &value{}
	)

	ts.Equal(ErrNotFound, ts.im.Get(mockCtx, k, c))

	sv, err := json.Marshal(v)
	ts.NoError(err)
	ts.cache.Set(mockCtx, keys.RedisKey(ts.im.pfx, k), sv, time.Second)
	ts.NoError(ts.im.Get(mockCtx, k, c))
	ts.Equal(v, *c)

	time.Sleep(time.Second)

	_, _, err = ts.cache.Get(mockCtx, keys.RedisKey(ts.im.pfx, k))
	ts.Equal(provider.ErrNotFound, err)
}

func (ts *testsuite) TestSet() {
	var (
		k = "key"
		v = value{"value"}
		c = &value{}
	)

	ts.NoError(ts.im.Set(mockCtx, k, v))

	sv, _, err := ts.cache.Get(mockCtx, keys.RedisKey(ts.im.pfx, k))
	ts.NoError(err)

	ts.NoError(json.Unmarshal(sv, c))
	ts.Equal(v, *c)

	time.Sleep(time.Second)

	_, _, err = ts.cache.Get(mockCtx, keys.RedisKey(ts.im.pfx, k))
	ts.Equal(provider.ErrNotFound, err)
}

func (ts *testsuite) TestGetByFunc() {
	var (
		k = "key"
		v = value{"value"}
		c = &value{}
	)

	ts.NoError(ts.im.GetByFunc(mockCtx, k, c, func() (interface{}, error) {
		return &v, nil
	}))

	ts.Equal(v, *c)

	sv, _, err := ts.cache.Get(mockCtx, keys.RedisKey(ts.im.pfx, k))
	ts.NoError(err)
	ts.NoError(json.Unmarshal(sv, c))
	ts.Equal(v, *c)

	time.Sleep(time.Second)

	_, _, err = ts.cache.Get(mockCtx, keys.RedisKey(ts.im.pfx, k))
	ts.Equal(provider.ErrNotFound, err)
}
