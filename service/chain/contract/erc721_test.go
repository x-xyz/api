package contract

import (
	"testing"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/chain"
)

func TestErc721_Supports721Interface(t *testing.T) {
	req := require.New(t)
	ctx := bCtx.Background()
	urls := map[int32]string{
		1: "rpc_url",
		250: "https://rpc.ftm.tools",
	}
	chainService, err := chain.NewClient(ctx, &chain.ClientCfg{RpcUrls: urls})
	req.NoError(err)
	e := NewErc721(chainService)
	tests := []struct {
		chainId    int32
		addr       string
		res        bool
		shouldFail bool
	}{
		{
			// galaxykats
			chainId:    1,
			addr:       "0x71c4658acc7b53ee814a29ce31100ff85ca23ca7",
			res:        true,
			shouldFail: false,
		},
		{
			// non contract
			chainId:    1,
			addr:       "0x94EaD797046c7b654cab82C1c27ad223b6501f1f",
			res:        false,
			shouldFail: true,
		},
		{
			// don't support
			chainId:    1,
			addr:       "0x76BE3b62873462d2142405439777e971754E8E77",
			res:        false,
			shouldFail: false,
		},
		// {
		// 	// 721
		// 	chainId:    56,
		// 	addr:       "0x0a8901b0E25DEb55A87524f0cC164E9644020EBA",
		// 	res:        true,
		// 	shouldFail: false,
		// },
		// {
		// 	// non contract
		// 	chainId:    56,
		// 	addr:       "0x94EaD797046c7b654cab82C1c27ad223b6501f1f",
		// 	res:        false,
		// 	shouldFail: true,
		// },
		// {
		// 	// don't support
		// 	chainId:    56,
		// 	addr:       "0x76BE3b62873462d2142405439777e971754E8E77",
		// 	res:        false,
		// 	shouldFail: false,
		// },
		{
			// 721
			chainId:    250,
			addr:       "0x10B11Eb388520D9F71FAC7aeBB4A0e501bE08df6",
			res:        true,
			shouldFail: false,
		},
		{
			// non contract
			chainId:    250,
			addr:       "0x94EaD797046c7b654cab82C1c27ad223b6501f1f",
			res:        false,
			shouldFail: true,
		},
		{
			// don't support
			chainId:    250,
			addr:       "0xd19eb6f25de99a993a73a3931c94cf3b299bee03",
			res:        false,
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		supports, err := e.Supports721Interface(ctx, tt.chainId, tt.addr)
		if tt.shouldFail {
			req.Error(err)
			continue
		}
		req.NoError(err)
		req.Equal(tt.res, supports)
	}
}
