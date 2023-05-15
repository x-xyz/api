package contract

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/chain"
)

func TestRoyaltyEngine_GetRoyalty(t *testing.T) {
	req := require.New(t)
	ctx := bCtx.Background()
	chainService, err := chain.NewClient(ctx, &chain.ClientCfg{
		RpcUrls: map[int32]string{
			1: "rpc_url",
		},
	})
	req.NoError(err)
	re := NewRoyaltyEngine(chainService)
	reAddr := "0xa3e5e24e075a998abdcb32eff31404e1742542d0"
	collection := "0x2d3e3def08848d405df3418bf91aa6876a057cd7"
	tokenId := big.NewInt(10)
	value := big.NewInt(10000)
	recipients, values, err := re.GetRoyalty(ctx, 1, reAddr, collection, tokenId, value)
	req.NoError(err)
	req.Equal([]string{"0xAAe7aC476b117bcCAfE2f05F582906be44bc8FF1"}, recipients)
	req.Equal([]*big.Int{big.NewInt(250)}, values)
}
