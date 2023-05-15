package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/base/validator"
	"github.com/x-xyz/goapi/service/redis"
)

// GoMiddleware represent the data-struct for middleware
type GoMiddleware struct {
	// another stuff , may be needed by middleware
}

// InitMiddleware initialize the middleware
func InitMiddleware() *GoMiddleware {
	return &GoMiddleware{}
}

// CORS will handle the CORS middleware
func (m *GoMiddleware) CORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		return next(c)
	}
}

// AddContexte adds custome context into echo
func (m *GoMiddleware) AddContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			cont := ctx.WithValue(ctx.Background(), "requestID", c.Response().Header().Get(echo.HeaderXRequestID))
			c.Set("ctx", cont)
			return next(c)
		}
	}
}

func (m *GoMiddleware) AddRedis(redis redis.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("redis", redis)
			return next(c)
		}
	}
}

// ResponseLogger logs response for every request
func (m *GoMiddleware) ResponseLogger() echo.MiddlewareFunc {
	met := metrics.New("http")
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer met.BumpTime("request.time", "method", c.Request().Method, "path", c.Path()).End()

			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			fields := log.Fields{
				"ms":             time.Since(start).Seconds() * 1000,
				"httpStatus":     c.Response().Status,
				"host":           req.Host,
				"remoteIP":       c.RealIP(),
				"uri":            c.Request().URL.Path,
				"httpMethod":     c.Request().Method,
				"size":           res.Size,
				"userAgent":      req.UserAgent(),
				"acceptEncoding": c.Request().Header.Get("Accept-Encoding"),
				"referer":        c.Request().Header.Get("Referer"),
				"ip_country":     c.Request().Header.Get("Cf-Ipcountry"),
				"imei":           c.Request().Header.Get("imei"),
				"platform":       c.Request().Header.Get("platform"),
				"version":        c.Request().Header.Get("version"),
				"device_model":   c.Request().Header.Get("device_model"),
				"deviceID":       c.Request().Header.Get("deviceID"),
			}

			n := res.Status
			switch {
			case n >= 400:
				fields["nextErr"] = err
			default:
			}

			c.Get("ctx").(ctx.Ctx).WithFields(fields).Info("response")
			return nil
		}
	}
}

func IsValidAddress(param string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !validator.IsValidAddress(c.Param(param)) {
				return delivery.MakeJsonResp(c, http.StatusBadRequest, "invalid address")
			}
			return next(c)
		}
	}
}
