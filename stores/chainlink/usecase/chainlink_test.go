package usecase

import (
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	mockPaytoken "github.com/x-xyz/goapi/domain/mocks"
	mockChainlink "github.com/x-xyz/goapi/service/chainlink/mocks"
)

var (
	mockCtx = ctx.Background()
)

type testsuite struct {
	suite.Suite
	mockChainlink *mockChainlink.Chainlink
	mockPaytoken  *mockPaytoken.PayTokenRepo
	subject       *impl
}

func Test(t *testing.T) {
	suite.Run(t, new(testsuite))
}

func (t *testsuite) SetupTest() {
	t.mockChainlink = &mockChainlink.Chainlink{}
	t.mockPaytoken = &mockPaytoken.PayTokenRepo{}
	t.subject = &impl{
		chainlink: t.mockChainlink,
		paytoken:  t.mockPaytoken,
	}
}

func (t *testsuite) TestGetLastestAnswer() {
	var (
		chainId   = domain.ChainId(1234567)
		tokenAddr = domain.Address("abcdefg")
		feedAddr  = domain.Address("gfedcba")
	)

	t.mockPaytoken.
		On("FindOne", mockCtx, chainId, tokenAddr).
		Return(&domain.PayToken{
			ChainlinkProxyAddress: feedAddr,
			Decimals:              3,
		}, nil)

	t.mockChainlink.
		On("GetLatestAnswer", mockCtx, chainId, feedAddr).
		Return(big.NewInt(1234), nil)

	val, err := t.subject.GetLatestAnswer(mockCtx, chainId, tokenAddr)
	t.NoError(err)
	t.Equal(decimal.NewFromFloat(1.234), val)
}
