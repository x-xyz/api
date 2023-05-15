package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/delivery"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/moderator"
)

type AuthMiddleware struct {
	auth           domain.AuthUsecase
	moderator      moderator.Usecase
	adminAddresses []string
}

func New(auth domain.AuthUsecase, moderator moderator.Usecase, adminAddresses []string) *AuthMiddleware {
	return &AuthMiddleware{
		auth:           auth,
		moderator:      moderator,
		adminAddresses: adminAddresses,
	}
}

func (m *AuthMiddleware) Auth() echo.MiddlewareFunc {
	return middleware.KeyAuth(m.validateAuthToken)
}

func (m *AuthMiddleware) OptionalAuth() echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: func(c echo.Context) bool {
			auth := c.Request().Header.Get(echo.HeaderAuthorization)
			return len(auth) == 0
		},
		Validator: m.validateAuthToken,
	})
}

func (m *AuthMiddleware) IsAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			address := c.Get("address").(domain.Address)

			for _, admin := range m.adminAddresses {
				if admin == string(address) {
					return next(c)
				}
			}

			return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, "require admin privilege")
		}
	}
}

func (m *AuthMiddleware) IsModerator() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Get("ctx").(ctx.Ctx)

			address := c.Get("address").(domain.Address)

			// skip admin
			for _, admin := range m.adminAddresses {
				if admin == string(address) {
					return next(c)
				}
			}

			if res, err := m.moderator.IsModerator(ctx, address); err != nil {
				return delivery.MakeJsonResp(c, http.StatusInternalServerError, err)
			} else if !res {
				return delivery.MakeJsonResp(c, http.StatusMethodNotAllowed, "require moderator privilege")
			} else {
				return next(c)
			}
		}
	}
}

func (m *AuthMiddleware) validateAuthToken(key string, c echo.Context) (bool, error) {
	ctx := c.Get("ctx").(ctx.Ctx)
	if ads, err := m.auth.ParseToken(ctx, key); err != nil {
		ctx.WithField("err", err).Error("auth.ParseToken failed")
		return false, err
	} else {
		c.Set("address", domain.Address(ads))
		return true, nil
	}
}
