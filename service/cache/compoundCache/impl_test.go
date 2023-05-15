package compoundcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/cache"
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
	im       *impl
	service1 cache.Service
	service2 cache.Service
}

func (ts *testsuite) SetupTest() {
	cache1 := primitive.NewPrimitive("test", 64)
	cache2 := primitive.NewPrimitive("test2", 64)

	ts.service1 = cache.New(cache.ServiceConfig{
		Ttl:   time.Second,
		Pfx:   "test",
		Cache: cache1,
	})

	ts.service2 = cache.New(cache.ServiceConfig{
		Ttl:   2 * time.Second,
		Pfx:   "test",
		Cache: cache2,
	})

	ts.im = NewCompoundCache([]cache.Service{
		ts.service1,
		ts.service2,
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

	// test cache 1
	ts.Equal(cache.ErrNotFound, ts.im.Get(mockCtx, k, c))

	err := ts.service1.Set(mockCtx, k, v)
	ts.service1.Get(mockCtx, k, c)
	ts.NoError(err)
	ts.NoError(ts.im.Get(mockCtx, k, c))
	ts.Equal(v, *c)

	time.Sleep(time.Second)

	err = ts.service1.Get(mockCtx, k, c)
	ts.Equal(provider.ErrNotFound, err)

	// test cache 2
	err = ts.service2.Set(mockCtx, k, v)
	ts.NoError(err)
	ts.NoError(ts.im.Get(mockCtx, k, c))
	ts.Equal(v, *c)

	// get from cache 1
	err = ts.service1.Get(mockCtx, k, c)
	ts.NoError(err)
	ts.Equal(v, *c)

	time.Sleep(time.Second)

	err = ts.service1.Get(mockCtx, k, c)
	ts.Equal(provider.ErrNotFound, err)
}

func (ts *testsuite) TestSet() {
	var (
		k = "key"
		v = value{"value"}
		c = &value{}
	)

	ts.NoError(ts.im.Set(mockCtx, k, v))

	err := ts.service1.Get(mockCtx, k, c)
	ts.NoError(err)
	ts.Equal(v, *c)

	err = ts.service2.Get(mockCtx, k, c)
	ts.NoError(err)
	ts.Equal(v, *c)

	time.Sleep(time.Second)

	err = ts.service1.Get(mockCtx, k, c)
	ts.Equal(provider.ErrNotFound, err)

	time.Sleep(time.Second)

	err = ts.service2.Get(mockCtx, k, c)
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

	err := ts.service1.Get(mockCtx, k, c)
	ts.NoError(err)
	ts.Equal(v, *c)

	err = ts.service2.Get(mockCtx, k, c)
	ts.NoError(err)
	ts.Equal(v, *c)

	time.Sleep(time.Second)

	err = ts.service1.Get(mockCtx, k, c)
	ts.Equal(provider.ErrNotFound, err)

	time.Sleep(time.Second)

	err = ts.service2.Get(mockCtx, k, c)
	ts.Equal(provider.ErrNotFound, err)
}
