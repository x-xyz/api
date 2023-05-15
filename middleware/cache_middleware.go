package middleware

import (
	"bufio"
	"bytes"
	"hash/fnv"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/service/cache"
	compoundcache "github.com/x-xyz/goapi/service/cache/compoundCache"
	"github.com/x-xyz/goapi/service/cache/provider"
	"github.com/x-xyz/goapi/service/cache/provider/primitive"
	redisCache "github.com/x-xyz/goapi/service/cache/provider/redis"
	"github.com/x-xyz/goapi/service/redis"
)

var (
	cacheMiddlewareLocalCache provider.Provider
	cacheMiddlewareRedisCache provider.Provider

	cacheMiddlewarePfx = "httpCacheMiddleware"

	once = sync.Once{}
)

func SetupCache(redis redis.Service) {
	once.Do(func() {
		cacheMiddlewareLocalCache = primitive.NewPrimitive("httpCacheMiddleware", 1024)
		cacheMiddlewareRedisCache = redisCache.NewRedis(redis)
	})
}

// Response is the cached response data structure.
type Response struct {
	// Value is the cached response value.
	Value []byte

	// Header is the cached response header.
	Header http.Header
}

type bodyDumpResponseWriter struct {
	statusCode int
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func sortURLParams(URL *url.URL) {
	params := URL.Query()
	for _, param := range params {
		sort.Slice(param, func(i, j int) bool {
			return param[i] < param[j]
		})
	}
	URL.RawQuery = params.Encode()
}

func generateKey(URL string) string {
	hash := fnv.New64a()
	hash.Write([]byte(URL))

	return strconv.FormatUint(hash.Sum64(), 36)
}

func CacheHttp(ttl time.Duration) echo.MiddlewareFunc {
	if cacheMiddlewareLocalCache == nil {
		panic("need SetupCache before using CacheHttp")
	}

	if cacheMiddlewareRedisCache == nil {
		panic("need SetupCache before using CacheHttp")
	}

	primitiveTTL := 10 * time.Second
	if ttl < primitiveTTL {
		primitiveTTL = ttl
	}

	cacheService := compoundcache.NewCompoundCache([]cache.Service{
		// TODO: Some api's response size are very large, which may cause error on local cache.
		//       So temporarily disable local cache layer.
		//       Enable this after adding restriction on api response size.
		// cache.New(cache.ServiceConfig{
		// 	Ttl:   primitiveTTL,
		// 	Pfx:   cacheMiddlewarePfx,
		// 	Cache: cacheMiddlewareLocalCache,
		// }),
		cache.New(cache.ServiceConfig{
			Ttl:   ttl,
			Pfx:   cacheMiddlewarePfx,
			Cache: cacheMiddlewareRedisCache,
		}),
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Get("ctx").(ctx.Ctx)

			sortURLParams(c.Request().URL)
			key := generateKey(c.Request().URL.String())

			response := Response{}
			err := cacheService.Get(ctx, key, &response)
			if err == nil {
				// cache hit
				for k, v := range response.Header {
					c.Response().Header().Set(k, strings.Join(v, ","))
				}
				c.Response().WriteHeader(http.StatusOK)
				c.Response().Write(response.Value)
				return nil
			} else if err != nil && err != cache.ErrNotFound {
				ctx.WithFields(log.Fields{
					"err": err,
				}).Error("failed to cacheService.Get")
			}

			// cache miss
			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, resBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
			c.Response().Writer = writer
			if err := next(c); err != nil {
				c.Error(err)
			}

			statusCode := writer.statusCode
			value := resBody.Bytes()
			if statusCode < 400 {
				response := Response{
					Value:  value,
					Header: writer.Header(),
				}

				err := cacheService.Set(ctx, key, response)
				if err != nil {
					ctx.WithFields(log.Fields{
						"err": err,
					}).Error("failed to cacheService.Set")
				}
			}

			return nil
		}
	}
}
