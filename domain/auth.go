package domain

import (
	"github.com/golang-jwt/jwt"
	"github.com/x-xyz/goapi/base/ctx"
)

type JwtCustomClaims struct {
	Address string `json:"data"` // name data for backward compatibility
	jwt.StandardClaims
}

type AuthUsecase interface {
	SignToken(ctx ctx.Ctx, address Address) (string, error)
	ParseToken(ctx ctx.Ctx, token string) (address string, err error)
}
