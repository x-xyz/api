package compound

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/cache/provider"
	"github.com/x-xyz/goapi/service/cache/provider/primitive"
)

var (
	mockCtx = ctx.Background()
)

type testsuite struct {
	suite.Suite
	lyr0 provider.Provider
	lyr1 provider.Provider
	im   *impl
}

func (ts *testsuite) SetupTest() {
	ts.lyr0 = primitive.NewPrimitive("layer 0", 64)
	ts.lyr1 = primitive.NewPrimitive("layer 1", 64)
	ts.im = NewCompound([]provider.Provider{ts.lyr0, ts.lyr1}).(*impl)
}

func Test(t *testing.T) {
	suite.Run(t, new(testsuite))
}

func (ts *testsuite) TestSet() {
	k := "key"
	v := []byte("value")

	ts.NoError(ts.im.Set(mockCtx, k, v, time.Second))
	r0, _, e := ts.lyr0.Get(mockCtx, k)
	ts.NoError(e)
	ts.Equal(v, r0)
	r1, _, e := ts.lyr1.Get(mockCtx, k)
	ts.NoError(e)
	ts.Equal(v, r1)

	time.Sleep(time.Second)
	_, _, e = ts.lyr0.Get(mockCtx, k)
	ts.Equal(provider.ErrNotFound, e)
	_, _, e = ts.lyr1.Get(mockCtx, k)
	ts.Equal(provider.ErrNotFound, e)
}

func (ts *testsuite) TestGet() {
	cases := []struct {
		Desc  string
		Key   string
		Val   string
		Err   error
		Cache provider.Provider
	}{
		{
			Desc:  "Success from layer 0",
			Key:   "key 0",
			Val:   "value 0",
			Err:   nil,
			Cache: ts.lyr0,
		},
		{
			Desc:  "Success from layer 1",
			Key:   "key 1",
			Val:   "value 1",
			Err:   nil,
			Cache: ts.lyr1,
		},
		{
			Desc: "Not found",
			Err:  provider.ErrNotFound,
		},
	}

	for _, c := range cases {
		if len(c.Key) > 0 && c.Cache != nil {
			ts.NoError(c.Cache.Set(mockCtx, c.Key, []byte(c.Val), 10), c.Desc)
		}

		v, _, e := ts.im.Get(mockCtx, c.Key)
		ts.Equal(c.Val, string(v), c.Desc)
		ts.Equal(c.Err, e, c.Desc)
	}
}

func (ts *testsuite) TestIncr() {
	cases := []struct {
		Desc  string
		Key   string
		Val   string
		Incr  int
		Res   int64
		Err   error
		Cache provider.Provider
	}{
		{
			Desc:  "Success",
			Key:   "key 1",
			Val:   "7",
			Incr:  3,
			Res:   10,
			Cache: ts.lyr1,
		},
		{
			Desc:  "Not found in last cache",
			Key:   "key 0",
			Val:   "5",
			Incr:  1,
			Cache: ts.lyr0,
			Err:   provider.ErrNotFound,
		},
		{
			Desc: "Not found",
			Err:  provider.ErrNotFound,
		},
	}

	for _, c := range cases {
		if len(c.Key) > 0 {
			ts.NoError(c.Cache.Set(mockCtx, c.Key, []byte(c.Val), 10), c.Desc)
		}

		v, _, e := ts.im.Incr(mockCtx, c.Key, c.Incr)
		ts.Equal(c.Res, v, c.Desc)
		ts.Equal(c.Err, e, c.Desc)
	}
}
