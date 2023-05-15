package chainlink

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/cache"
	"github.com/x-xyz/goapi/service/chain"
)

var (
	mockCTX = ctx.Background()
)

type testsuite struct {
	suite.Suite
	chainlink   Chainlink
	chainClient chain.Client
	cache       cache.Service
}

func Test(t *testing.T) {
	suite.Run(t, new(testsuite))
}

func (t *testsuite) SetupSuite() {
	cli, err := chain.NewClient(mockCTX, &chain.ClientCfg{
		RpcUrls: map[int32]string{
			1: "rpc_url",
		},
		ArchiveRpcUrls: map[int32]string{
			1: "rpc_url",
		},
	})
	if err != nil {
		panic(err)
	}

	t.chainClient = cli
}

func (t *testsuite) SetupTest() {
	t.chainlink = New(t.chainClient)
}

func (t *testsuite) TestGetLastestAnswer() {
	chainId := domain.ChainId(1)
	ethAddr := domain.Address("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
	crvAddr := domain.Address("0xCd627aA160A6fA45Eb793D19Ef54f5062F20f33f")

	ethPrice, err := t.chainlink.GetLatestAnswer(mockCTX, chainId, ethAddr)
	t.NoError(err)
	t.NotEqual(big.NewInt(0), ethPrice)

	crvPrice, err := t.chainlink.GetLatestAnswer(mockCTX, chainId, crvAddr)
	t.NoError(err)
	t.NotEqual(big.NewInt(0), crvPrice)

	t.NotEqual(crvPrice, ethPrice)
}
func (t *testsuite) TestGetLastestAnswerAt() {
	chainId := domain.ChainId(1)
	ethAddr := domain.Address("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
	crvAddr := domain.Address("0xCd627aA160A6fA45Eb793D19Ef54f5062F20f33f")
	blk := big.NewInt(13086888)

	ethPrice, err := t.chainlink.GetLatestAnswerAt(mockCTX, chainId, ethAddr, blk)
	t.NoError(err)
	t.Equal(big.NewInt(334915000000), ethPrice)

	crvPrice, err := t.chainlink.GetLatestAnswerAt(mockCTX, chainId, crvAddr, blk)
	t.NoError(err)
	t.Equal(big.NewInt(232298188), crvPrice)
}
