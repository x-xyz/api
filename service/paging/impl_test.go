package paging

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/redisclient"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/service/redis"
)

type PagingSuite struct {
	suite.Suite

	// im         *impl
	redisCache redis.Service
}

func TestPagingSuite(t *testing.T) {
	suite.Run(t, new(PagingSuite))
}

func (s *PagingSuite) SetupSuite() {
	redisCacheName := "cache"
	redisCacheURI := "localhost:6379"
	redisCachePwd := ""
	redisCachePool := redisclient.MustConnectRedis(redisCacheURI, redisCachePwd, redisclient.RedisParam{
		PoolMultiplier: 20,
		Retry:          true,
	})

	s.redisCache = redis.New(redisCacheName, metrics.New(redisCacheName), &redis.Pools{
		Src: redisCachePool,
	})
}

func (s *PagingSuite) SetupTest() {
	conn, _ := s.redisCache.GetConn()
	conn.Do("FLUSHALL")
}

func (s *PagingSuite) TestEncodeCursor() {
	cases := []struct {
		name string
		data cursorStruct
		want string
	}{
		{
			name: "caes 1",
			data: cursorStruct{
				createTs:   123,
				totalCount: 1000,
				offset:     456,
			},
			want: "123:1000:456",
		},
		{
			name: "case 2",
			data: cursorStruct{
				createTs:   234,
				totalCount: 2000,
				offset:     789,
			},
			want: "234:2000:789",
		},
	}

	for _, c := range cases {
		base64String := encodeCursor(c.data)
		data, err := base64.StdEncoding.DecodeString(base64String)
		s.Nil(err)
		s.Equal(c.want, string(data))
	}
}

func (s *PagingSuite) TestDecodeCursor() {
	// successful case
	base64String := base64.StdEncoding.EncodeToString([]byte("123:1000:456"))
	cur, err := decodeCursor(base64String)
	s.Nil(err)
	s.Equal(&cursorStruct{
		createTs:   123,
		totalCount: 1000,
		offset:     456,
	}, cur)

	// bad cursor
	base64String = base64.StdEncoding.EncodeToString([]byte("123:ABC:456"))
	cur, err = decodeCursor(base64String)
	s.NotNil(err)
	s.Nil(cur)
}

func (s *PagingSuite) TestGet1() {
	mockCtx := ctx.Background()
	mockKey := "mockKey"

	type D struct {
		Value int
	}

	mockGetter := func(ctx ctx.Ctx, key string) (wholeList interface{}, err error) {
		return []D{
			{
				Value: 1,
			},
			{
				Value: 2,
			},
		}, nil
	}

	im := New(&PagingConfig{
		RedisCache:    s.redisCache,
		KeyPfx:        "test",
		Getter:        mockGetter,
		RenewDuration: 10 * time.Second,
		CacheDuration: 10 * time.Second,
	}).(*impl)

	output := []D{}
	nextCursor, totalCount, err := im.Get(mockCtx, mockKey, "", 1, &output)
	s.Equal([]D{{Value: 1}}, output)
	s.Equal(2, totalCount)
	s.Nil(err)

	nextCursor, totalCount, err = im.Get(mockCtx, mockKey, nextCursor, 1, &output)
	s.Equal([]D{{Value: 2}}, output)
	s.Equal(2, totalCount)
	s.Nil(err)
	s.Equal("", nextCursor)
}

func (s *PagingSuite) TestGet2() {
	mockCtx := ctx.Background()
	mockKey := "mockKey"

	type D struct {
		Value int
	}

	mockGetter := func(ctx ctx.Ctx, key string) (interface{}, error) {
		return []D{
			{
				Value: 1,
			},
			{
				Value: 2,
			},
			{
				Value: 3,
			},
		}, nil
	}

	im := New(&PagingConfig{
		RedisCache:    s.redisCache,
		KeyPfx:        "test",
		Getter:        mockGetter,
		RenewDuration: 10 * time.Second,
		CacheDuration: 10 * time.Second,
	}).(*impl)

	output := []*D{}
	_, totalCount, err := im.Get(mockCtx, mockKey, "", 5, &output)
	s.Equal([]*D{{Value: 1}, {Value: 2}, {Value: 3}}, output)
	s.Equal(3, totalCount)
	s.Nil(err)

	nextCursor, totalCount, err := im.Get(mockCtx, mockKey, "", 2, &output)
	s.Equal([]*D{{Value: 1}, {Value: 2}}, output)
	s.Equal(3, totalCount)
	s.Nil(err)

	nextCursor, totalCount, err = im.Get(mockCtx, mockKey, nextCursor, 2, &output)
	s.Equal([]*D{{Value: 3}}, output)
	s.Equal(3, totalCount)
	s.Nil(err)
	s.Equal("", nextCursor)

	// bad container
	_, _, err = im.Get(mockCtx, mockKey, "", 2, output)
	s.Equal(ErrBadContainer, err)
}
