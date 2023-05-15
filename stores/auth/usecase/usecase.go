package usecase

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
)

type impl struct {
	jwtSecret []byte
	account   account.Usecase
}

func New(jwtSecret string, account account.Usecase) domain.AuthUsecase {
	return &impl{
		jwtSecret: []byte(jwtSecret),
		account:   account,
	}
}

func (im *impl) SignToken(ctx ctx.Ctx, address domain.Address) (string, error) {
	_, err := im.account.Get(ctx, address)

	if err != nil && err != domain.ErrNotFound {
		return "", err
	}

	if err == domain.ErrNotFound {
		if _, err := im.account.Create(ctx, address); err != nil {
			return "", err
		}
	}

	claims := domain.JwtCustomClaims{
		Address: string(address),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	if ss, err := token.SignedString(im.jwtSecret); err != nil {
		ctx.WithField("err", err).Error("token.SignedString failed")
		return "", err
	} else {
		return ss, nil
	}
}

func (im *impl) ParseToken(ctx ctx.Ctx, str string) (string, error) {
	token, err := jwt.ParseWithClaims(str, &domain.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return im.jwtSecret, nil
	})

	if claims, ok := token.Claims.(*domain.JwtCustomClaims); ok && token.Valid {
		return claims.Address, nil
	}

	return "", err
}
