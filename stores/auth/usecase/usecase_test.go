package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	mAccount "github.com/x-xyz/goapi/domain/account/mocks"
	"github.com/x-xyz/goapi/stores/auth/usecase"
)

func TestSignAndParseToken(t *testing.T) {
	mockAccountUC := &mAccount.Usecase{}

	mockAccountUC.On("Get", mock.Anything, domain.Address("my-address")).Return(nil, nil)

	ctx := ctx.Background()
	u := usecase.New("jwt-secret", mockAccountUC)
	tkn, err := u.SignToken(ctx, "my-address")
	assert.NoError(t, err)
	assert.NotEmpty(t, tkn)
	ads, err := u.ParseToken(ctx, tkn)
	assert.NoError(t, err)
	assert.Equal(t, "my-address", ads)
}
