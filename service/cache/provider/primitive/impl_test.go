package primitive

import (
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/cache/provider"
)

var (
	mockCtx = ctx.Background()
)

type testsuite struct {
	suite.Suite
	im *impl
}

func (ts *testsuite) SetupTest() {
	ts.im = NewPrimitive("", 64).(*impl)
}

func (ts *testsuite) TearDownTest() {
	ts.im.cache.Clear()
}

func Test(t *testing.T) {
	suite.Run(t, new(testsuite))
}

func (ts *testsuite) TestSet() {
	k := "key"
	v := []byte("value")

	ts.NoError(ts.im.Set(mockCtx, k, v, time.Second))
	r, e := ts.im.cache.Get([]byte(k))
	ts.NoError(e)
	ts.Equal(v, r)

	time.Sleep(time.Second)
	_, e = ts.im.cache.Get([]byte(k))
	ts.Equal(freecache.ErrNotFound, e)
}

func (ts *testsuite) TestGet() {
	cases := []struct {
		Desc string
		Key  string
		Val  string
		Err  error
	}{
		{
			Desc: "Success",
			Key:  "key",
			Val:  "value",
			Err:  nil,
		},
		{
			Desc: "Not found",
			Err:  provider.ErrNotFound,
		},
	}

	for _, c := range cases {
		if len(c.Key) > 0 {
			ts.NoError(ts.im.cache.Set([]byte(c.Key), []byte(c.Val), 10), c.Desc)
		}

		v, _, e := ts.im.Get(mockCtx, c.Key)
		ts.Equal(c.Val, string(v), c.Desc)
		ts.Equal(c.Err, e, c.Desc)
	}
}

func (ts *testsuite) TestIncr() {
	cases := []struct {
		Desc string
		Key  string
		Val  string
		Incr int
		Res  int64
		Err  error
	}{
		{
			Desc: "Success",
			Key:  "key",
			Val:  "5",
			Incr: 1,
			Res:  6,
		},
		{
			Desc: "Not found",
			Err:  provider.ErrNotFound,
		},
	}

	for _, c := range cases {
		if len(c.Key) > 0 {
			ts.NoError(ts.im.cache.Set([]byte(c.Key), []byte(c.Val), 10), c.Desc)
		}

		v, _, e := ts.im.Incr(mockCtx, c.Key, c.Incr)
		ts.Equal(c.Res, v, c.Desc)
		ts.Equal(c.Err, e, c.Desc)
	}

}
