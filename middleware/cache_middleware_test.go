package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/redisclient"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/service/redis"
)

type cacheMiddlewareSuite struct {
	suite.Suite

	redis redis.Service
}

func (s *cacheMiddlewareSuite) SetupSuite() {
	redisCacheName := "cache"
	redisCacheURI := "localhost:6379"
	redisCachePwd := ""
	redisCachePoolMultiplier := float64(20)
	redisCachePool := redisclient.MustConnectRedis(redisCacheURI, redisCachePwd, redisclient.RedisParam{
		PoolMultiplier: redisCachePoolMultiplier,
		Retry:          true,
	})
	redisCache := redis.New(redisCacheName, metrics.New(redisCacheName), &redis.Pools{
		Src: redisCachePool,
	})

	SetupCache(redisCache)

	s.redis = redisCache
}

func TestCacheMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(cacheMiddlewareSuite))
}

func (s *cacheMiddlewareSuite) TestCacheMiddleware() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	res := "Hello, World"
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, res)
	}

	c := e.NewContext(req, rec)
	cont := ctx.WithValue(ctx.Background(), "requestID", c.Response().Header().Get(echo.HeaderXRequestID))
	c.Set("ctx", cont)

	if s.NoError(CacheHttp(30 * time.Second)(h)(c)) {
		s.Equal(http.StatusOK, rec.Code)
		s.Equal(res, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	res2 := "Hello, again"
	h2 := func(c echo.Context) error {
		return c.String(http.StatusOK, res2)
	}
	c2 := e.NewContext(req2, rec2)
	c2.Set("ctx", cont)
	c2.Set("redis", s.redis)

	if s.NoError(CacheHttp(30 * time.Second)(h2)(c2)) {
		s.Equal(http.StatusOK, rec2.Code)
		s.Equal(res, rec2.Body.String())
	}

	key := generateKey(req.URL.String())
	_, err := s.redis.Get(cont, "httpCacheMiddleware:"+key)
	s.Nil(err)
}
