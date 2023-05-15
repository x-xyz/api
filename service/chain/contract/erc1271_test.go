package contract

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/chain"
)

func TestErc1271_IsValidSignature(t *testing.T) {
	req := require.New(t)
	ctx := bCtx.Background()
	goerliChainId := int32(5)
	erc1271Addr := "0xAc461fDFc10C71861f37fe42589334e021BaA1ee"
	chainService, err := chain.NewClient(ctx, &chain.ClientCfg{
		RpcUrls: map[int32]string{
			goerliChainId: "rpc_url",
		},
	})
	req.NoError(err)
	erc1271 := NewErc1271(chainService)
	type args struct {
		hash      common.Hash
		signature []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid sig",
			args: args{
				hash:      common.HexToHash("0x01f6f4c6639ea7f7d4df5425aaefe85113235810e9dd52ccf56297a16191c3ea"),
				signature: hexutil.MustDecode("0xfae5218f6165f30bf7d8798d6f1990fde8fea58c336b36c8cd3078b4d8dc2a9d0448debd2b776fb0f6bdf91d1142474d4682057d290561814172bce4641108641c"),
			},
			want: true,
		},
		{
			name: "invalid sig",
			args: args{
				hash:      common.HexToHash("0x01f6f4c6639ea7f7d4df5425aaefe85113235810e9dd52ccf56297a16191c3ea"),
				signature: hexutil.MustDecode("0xfae5218f6165f30bf7d8798d6f1990fde8fea58c336b36c8cd3078b4d8dc2a9d0448debd2b776fb0f6bdf91d1142474d4682057d290561814172bce4641108641b"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := erc1271.IsValidSignature(ctx, goerliChainId, erc1271Addr, tt.args.hash, tt.args.signature)
			req.NoError(err)
			req.Equal(tt.want, got)
		})
	}
}
