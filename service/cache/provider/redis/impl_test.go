package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/cache/provider"
	"github.com/x-xyz/goapi/service/redis"
	mockRedis "github.com/x-xyz/goapi/service/redis/mocks"
)

var (
	mockCtx = ctx.Background()
)

type testsuite struct {
	suite.Suite
	im    *impl
	redis *mockRedis.Service
}

func (ts *testsuite) SetupTest() {
	ts.redis = &mockRedis.Service{}
	ts.im = NewRedis(ts.redis).(*impl)
}

func Test(t *testing.T) {
	suite.Run(t, new(testsuite))
}

func (ts *testsuite) TestSet() {
	k := "key"
	v := []byte("value")

	ts.redis.On("Set", mockCtx, k, v, time.Second).Return(nil).Once()
	ts.NoError(ts.im.Set(mockCtx, k, v, time.Second))
}

func (ts *testsuite) TestGet() {
	var (
		k   = "key"
		v   = []byte("value")
		res []byte
		ttl time.Duration
		err error
	)

	ts.redis.On("Get", mockCtx, k).Return(nil, redis.ErrNotFound).Once()
	res, _, err = ts.im.Get(mockCtx, k)
	ts.Equal([]byte(nil), res)
	ts.Equal(provider.ErrNotFound, err)

	ts.redis.On("Get", mockCtx, k).Return(v, nil).Once()
	ts.redis.On("TTL", mockCtx, k).Return(int(time.Second.Seconds()), nil).Once()
	res, ttl, err = ts.im.Get(mockCtx, k)
	ts.Equal(v, res)
	ts.Equal(time.Second, ttl)
	ts.NoError(err)
}

func (ts *testsuite) TestIncr() {
	var (
		k   = "key"
		res int64
		ttl time.Duration
		err error
	)

	ts.redis.On("Exists", mockCtx, k).Return(false, nil).Once()
	_, _, err = ts.im.Incr(mockCtx, k, 2)
	ts.Equal(provider.ErrNotFound, err)

	ts.redis.On("Exists", mockCtx, k).Return(true, nil).Once()
	ts.redis.On("Incrby", mockCtx, k, 2).Return(int64(3), nil).Once()
	ts.redis.On("TTL", mockCtx, k).Return(int(time.Second.Seconds()), nil).Once()
	res, ttl, err = ts.im.Incr(mockCtx, k, 2)
	ts.Equal(int64(3), res)
	ts.Equal(time.Second, ttl)
	ts.NoError(err)
}
