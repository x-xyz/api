package ctx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type testsuite struct {
	suite.Suite
}

func Test(t *testing.T) {
	suite.Run(t, new(testsuite))
}

func (ts *testsuite) TestWithValue() {
	bg := Background()
	ctx := WithValue(bg, "foo", "bar")
	ts.Equal("bar", ctx.Value("foo"))
}

func (ts *testsuite) TestWithValues() {
	bg := Background()
	ctx := WithValues(bg, map[string]interface{}{
		"a": "b",
		"c": "d",
	})
	ts.Equal("b", ctx.Value("a"))
	ts.Equal("d", ctx.Value("c"))
}

func (ts *testsuite) TestWithCancel() {
	bg := Background()
	ctx, cancel := WithCancel(bg)
	defer cancel()
	after100Ms := func(ctx context.Context) bool {
		for {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(100 * time.Millisecond):
				return true
			}
		}
	}
	res := true
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	res = after100Ms(ctx)
	ts.Equal(false, res)
}

func (ts *testsuite) TestTimeout() {
	bg := Background()
	ctx, cancel := WithTimeout(bg, 10*time.Millisecond)
	defer cancel()
	after100Ms := func(ctx context.Context) bool {
		for {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(100 * time.Millisecond):
				return true
			}
		}
	}
	res := after100Ms(ctx)
	ts.Equal(false, res)
	ts.Equal("context deadline exceeded", ctx.Err().Error())
}
