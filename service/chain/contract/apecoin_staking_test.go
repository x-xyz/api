package contract

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/chain"
)

func Test_apecoinStaking_NftPosition(t *testing.T) {
	req := require.New(t)
	ctx := bCtx.Background()
	chainService, err := chain.NewClient(ctx, &chain.ClientCfg{
		RpcUrls: map[int32]string{
			1: "rpc_url",
		},
	})
	req.NoError(err)
	staking := NewApecoinStaking(chainService, 1, common.HexToAddress("0x5954aB967Bc958940b7EB73ee84797Dc8a2AFbb9"))
	poolId := big.NewInt(1)
	tokenId := big.NewInt(10)
	stakedAmount, rewardsDept, err := staking.NftPosition(ctx, poolId, tokenId)
	req.NoError(err)
	t.Logf("%s", stakedAmount.String())
	t.Logf("%s", rewardsDept.String())
}
