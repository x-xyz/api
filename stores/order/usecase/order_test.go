package usecase

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	pricefomatter "github.com/x-xyz/goapi/base/price_fomatter/mocks"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	mErc1155 "github.com/x-xyz/goapi/domain/erc1155/mocks"
	mDomain "github.com/x-xyz/goapi/domain/mocks"
	"github.com/x-xyz/goapi/domain/nftitem"
	mNftitem "github.com/x-xyz/goapi/domain/nftitem/mocks"
	"github.com/x-xyz/goapi/domain/order"
	mOrder "github.com/x-xyz/goapi/domain/order/mocks"
	"github.com/x-xyz/goapi/service/query"
)

func TestVerifyOrder(t *testing.T) {
	req := require.New(t)
	_ctx := ctx.Background()
	user1 := domain.Address("0xce4468e7ce84aceb74363f4ea64e5a038176f369")
	user2 := domain.Address("0xdf8650b0ca1260f7a2f4fdff9082aede554f65ad")
	weth := domain.Address("0xb4fbf271143f4fbf7b91a5ded31805e42b2208d6")
	exchangeAddress := domain.Address("0x1a01ecd2263a9d5b5967667e508ea22db478bc4b")
	erc20Address := domain.Address("0x07fe9ffd85b54a3a18467d3b5e91a55ecc52a268")
	erc721Address := domain.Address("0xdcf0de6b17785a143d006e1515a6afd123cde8ba")
	erc1155Address := domain.Address("0x23c0221b2b66071afdcce502a103f18ec2666a12")
	strategyFixedPrice := domain.Address("0xa7ca695b37854181f09c1c39a0cdcffc8db7a667")
	strategyPrivateSale := domain.Address("0x54a769173d97432a48371b022709117c090298e3")
	strategyCollectionOffer := domain.Address("0x2e9e733cb0394aace1226e34313f12b0764be65a")

	cases := []struct {
		VerifyingContract domain.Address
		MakerOrder        order.Order
	}{
		{
			VerifyingContract: domain.Address("0x322813fd9a801c5507c9de605d63cea4f2ce6c44"),
			MakerOrder: order.Order{
				ChainId: domain.ChainId(31337),
				IsAsk:   true,
				Signer:  domain.Address("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"),
				Items: []order.Item{
					{Collection: domain.Address("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"), TokenId: domain.TokenId("1"), Amount: "1", Price: "1"},
					{Collection: domain.Address("0x322813fd9a801c5507c9de605d63cea4f2ce6c44"), TokenId: domain.TokenId("2"), Amount: "2", Price: "2"},
				},
				Strategy:           domain.Address("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"),
				Currency:           domain.Address("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"),
				Nonce:              "1",
				StartTime:          "1",
				EndTime:            "1",
				MinPercentageToAsk: "1",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x",
				S:                  "0x0448debd2b776fb0f6bdf91d1142474d4682057d290561814172bce464110864",
				R:                  "0xfae5218f6165f30bf7d8798d6f1990fde8fea58c336b36c8cd3078b4d8dc2a9d",
				V:                  28,
			},
		},
		{
			// buy fixed price
			// https://goerli.etherscan.io/tx/0xb32c0eefd9f89677fd7e7dfe4bd8683d5a41fae3a47a5d530d75dc410c60bc81
			// https://goerli.etherscan.io/tx/0x17300504a0d26f1fb06214a72b5d18a552b201e0ba6abb6f204e16c171dca860
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   true,
				Signer:  user1,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("0"), Amount: "1", Price: decimal.New(10000, 18).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("0"), Amount: "10", Price: decimal.New(10000, 18).String()},
				},
				Strategy:           strategyFixedPrice,
				Currency:           erc20Address,
				Nonce:              "0",
				StartTime:          decimal.NewFromInt(0x62a85d50).String(),
				EndTime:            decimal.NewFromInt(0x62a86b60).String(),
				MinPercentageToAsk: "9000",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x",
				R:                  "0x6c497788bc31a643be11394f36685fc76f76e25a66b4ef9f92324f7bba046f6c",
				S:                  "0x6be1e66ed3e964a2e01599bca4c63e390ef53d399f37a6bd4378b42e4f825dc5",
				V:                  0x1c,
			},
		},
		{
			// buy fixed price with eth/weth
			// https://goerli.etherscan.io/tx/0x99fff8ae71a8a786441992ec6e5e55f2207fc48775353af696ebea7585eb0dd6
			// https://goerli.etherscan.io/tx/0xcccb175688cf54a3e24d3d073556cfed951c793f02184c31be61dfa169229160
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   true,
				Signer:  user1,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("1"), Amount: "1", Price: decimal.New(10000, 9).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("1"), Amount: "10", Price: decimal.New(10000, 9).String()},
				},
				Strategy:           strategyFixedPrice,
				Currency:           weth,
				Nonce:              "1",
				StartTime:          decimal.NewFromInt(0x62a85d6e).String(),
				EndTime:            decimal.NewFromInt(0x62a86b7e).String(),
				MinPercentageToAsk: "9000",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x",
				R:                  "0x1b496f4021176f140e1cf41f956eba853af07dde2d75e4bd66935eb283a8bf38",
				S:                  "0x3a8cceb2684728ebca4bcbca2ec7e399f53cc4dc29b4603c301c5f8646688461",
				V:                  0x1c,
			},
		},
		{
			// accept fixed price
			// https://goerli.etherscan.io/tx/0x274a57c2576f065e53e0a41550bd85a9b3ccd0530699800f310ed9c498ceb8c4
			// https://goerli.etherscan.io/tx/0x48a8c5afadb99e4be266b4c3f8c44f32dd3b896260307070798959c7cbcbb8cb
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   false,
				Signer:  user2,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("2"), Amount: "1", Price: decimal.New(10000, 18).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("2"), Amount: "10", Price: decimal.New(10000, 18).String()},
				},
				Strategy:           strategyFixedPrice,
				Currency:           erc20Address,
				Nonce:              "2",
				StartTime:          decimal.NewFromInt(0x62a85db9).String(),
				EndTime:            decimal.NewFromInt(0x62a86bc9).String(),
				MinPercentageToAsk: "0",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x",
				R:                  "0xcd716aa453477a9ad595dff19e86a29e26da37ad7153545312d07de4be47f1ea",
				S:                  "0x19604de939790e29be08201395dbc8f1a8e77c54ee43a373a5ea78d90340752a",
				V:                  0x1b,
			},
		},
		{
			// buy private sale
			// https://goerli.etherscan.io/tx/0x1ebb46b45674fd1597491a7dc8dc7d8a51cc00e1cc86df9febaa80215a2d8724
			// https://goerli.etherscan.io/tx/0xe69a8188342177f595a7f552c9e296a0ff7d58b09734ee7fa79a65c4542b0b63
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   true,
				Signer:  user1,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("3"), Amount: "1", Price: decimal.New(10000, 18).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("3"), Amount: "10", Price: decimal.New(10000, 18).String()},
				},
				Strategy:           strategyPrivateSale,
				Currency:           erc20Address,
				Nonce:              "3",
				StartTime:          decimal.NewFromInt(0x62a85dc8).String(),
				EndTime:            decimal.NewFromInt(0x62a86bd8).String(),
				MinPercentageToAsk: "9000",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x000000000000000000000000df8650b0ca1260f7a2f4fdff9082aede554f65ad",
				R:                  "0x8149fe5282ee9297bf064bc7f215c0ef458497724f11af30b0b15eb1cf6d9592",
				S:                  "0x35e4d56caeb1e6332f54b2699039b98776f2c14a8cc2035b25753eb8bc6b3535",
				V:                  0x1b,
			},
		},
		{
			// buy private sale with eth
			// https://goerli.etherscan.io/tx/0xdb8ea6726a708e9c8a25637e83a8b02d108b004df5a18d612ebe5f74a283a8a2
			// https://goerli.etherscan.io/tx/0x137115577fbae66b63521498592c034f585c12c94c92692515a54dfabf947600
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   true,
				Signer:  user1,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("4"), Amount: "1", Price: decimal.New(10000, 9).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("4"), Amount: "10", Price: decimal.New(10000, 9).String()},
				},
				Strategy:           strategyPrivateSale,
				Currency:           weth,
				Nonce:              "4",
				StartTime:          decimal.NewFromInt(0x62a85e13).String(),
				EndTime:            decimal.NewFromInt(0x62a86c23).String(),
				MinPercentageToAsk: "9000",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x000000000000000000000000df8650b0ca1260f7a2f4fdff9082aede554f65ad",
				R:                  "0x2e3245fc8e6856f5adc845f61a891458966c6bd2473df2857488bdcccc8e272b",
				S:                  "0x202a6e7a811530a23413fabb86b9232c0822806ba101a5596120e5b745a1df15",
				V:                  0x1b,
			},
		},
		{
			// accept collection offer
			// https://goerli.etherscan.io/tx/0x4c5f04400add14cda7a7bf206041b019419e8e2814431ecb253f3587d25aaac8
			// https://goerli.etherscan.io/tx/0xc1a22fbe55a9ad4abe70008756834aae04f4157587405d0b9608e36189991948
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   false,
				Signer:  user2,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("0"), Amount: "1", Price: decimal.New(10000, 18).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("0"), Amount: "10", Price: decimal.New(10000, 18).String()},
				},
				Strategy:           strategyCollectionOffer,
				Currency:           erc20Address,
				Nonce:              "5",
				StartTime:          decimal.NewFromInt(0x62a85e40).String(),
				EndTime:            decimal.NewFromInt(0x62a86c50).String(),
				MinPercentageToAsk: "0",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x",
				R:                  "0xa261ae089ef93e02622642ab894980e20e28fe7745b5f4811215d2e238901977",
				S:                  "0x60e5f4201121b7d59aee5f7f7870d3939be3a9c2506142f43bec85841cccc7a4",
				V:                  0x1b,
			},
		},
	}

	for _, c := range cases {
		err := verifyOrderSignature(_ctx, c.MakerOrder, c.VerifyingContract, nil)
		req.NoError(err)
	}
}

func TestHashOrderItem(t *testing.T) {
	req := require.New(t)
	user1 := domain.Address("0xce4468e7ce84aceb74363f4ea64e5a038176f369")
	user2 := domain.Address("0xdf8650b0ca1260f7a2f4fdff9082aede554f65ad")
	weth := domain.Address("0xb4fbf271143f4fbf7b91a5ded31805e42b2208d6")
	exchangeAddress := domain.Address("0x1a01ecd2263a9d5b5967667e508ea22db478bc4b")
	erc20Address := domain.Address("0x07fe9ffd85b54a3a18467d3b5e91a55ecc52a268")
	erc721Address := domain.Address("0xdcf0de6b17785a143d006e1515a6afd123cde8ba")
	erc1155Address := domain.Address("0x23c0221b2b66071afdcce502a103f18ec2666a12")
	strategyFixedPrice := domain.Address("0xa7ca695b37854181f09c1c39a0cdcffc8db7a667")
	strategyPrivateSale := domain.Address("0x54a769173d97432a48371b022709117c090298e3")
	strategyCollectionOffer := domain.Address("0x2e9e733cb0394aace1226e34313f12b0764be65a")

	cases := []struct {
		VerifyingContract domain.Address
		MakerOrder        order.Order
		OrderItemHashes   [][]byte
	}{
		{
			// buy fixed price
			// https://goerli.etherscan.io/tx/0xb32c0eefd9f89677fd7e7dfe4bd8683d5a41fae3a47a5d530d75dc410c60bc81
			// https://goerli.etherscan.io/tx/0x17300504a0d26f1fb06214a72b5d18a552b201e0ba6abb6f204e16c171dca860
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   true,
				Signer:  user1,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("0"), Amount: "1", Price: decimal.New(10000, 18).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("0"), Amount: "10", Price: decimal.New(10000, 18).String()},
				},
				Strategy:           strategyFixedPrice,
				Currency:           erc20Address,
				Nonce:              "0",
				StartTime:          decimal.NewFromInt(0x62a85d50).String(),
				EndTime:            decimal.NewFromInt(0x62a86b60).String(),
				MinPercentageToAsk: "9000",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x",
				R:                  "0x6c497788bc31a643be11394f36685fc76f76e25a66b4ef9f92324f7bba046f6c",
				S:                  "0x6be1e66ed3e964a2e01599bca4c63e390ef53d399f37a6bd4378b42e4f825dc5",
				V:                  0x1c,
			},
			OrderItemHashes: [][]byte{hexutil.MustDecode("0xce2e133fcf55b3f07cb4d5a6211236ba10ceaa2b3bc8ecc8249bc52c1ad6bf2e"), hexutil.MustDecode("0x18f8e86415e562f771f590a446162df2bcc58acbc0f80305360821dc9dcdc23e")},
		},
		{
			// buy fixed price with eth/weth
			// https://goerli.etherscan.io/tx/0x99fff8ae71a8a786441992ec6e5e55f2207fc48775353af696ebea7585eb0dd6
			// https://goerli.etherscan.io/tx/0xcccb175688cf54a3e24d3d073556cfed951c793f02184c31be61dfa169229160
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   true,
				Signer:  user1,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("1"), Amount: "1", Price: decimal.New(10000, 9).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("1"), Amount: "10", Price: decimal.New(10000, 9).String()},
				},
				Strategy:           strategyFixedPrice,
				Currency:           weth,
				Nonce:              "1",
				StartTime:          decimal.NewFromInt(0x62a85d6e).String(),
				EndTime:            decimal.NewFromInt(0x62a86b7e).String(),
				MinPercentageToAsk: "9000",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x",
				R:                  "0x1b496f4021176f140e1cf41f956eba853af07dde2d75e4bd66935eb283a8bf38",
				S:                  "0x3a8cceb2684728ebca4bcbca2ec7e399f53cc4dc29b4603c301c5f8646688461",
				V:                  0x1c,
			},
			OrderItemHashes: [][]byte{hexutil.MustDecode("0xd12d99f852b7b301705746184c548b128568e74a2a1d90fab40e88a004ea904e"), hexutil.MustDecode("0xe524d554d4e7c7fe74ba850ce5333c0094b2335d2b2eee45bd552583b069041f")},
		},
		{
			// accept fixed price
			// https://goerli.etherscan.io/tx/0x274a57c2576f065e53e0a41550bd85a9b3ccd0530699800f310ed9c498ceb8c4
			// https://goerli.etherscan.io/tx/0x48a8c5afadb99e4be266b4c3f8c44f32dd3b896260307070798959c7cbcbb8cb
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   false,
				Signer:  user2,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("2"), Amount: "1", Price: decimal.New(10000, 18).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("2"), Amount: "10", Price: decimal.New(10000, 18).String()},
				},
				Strategy:           strategyFixedPrice,
				Currency:           erc20Address,
				Nonce:              "2",
				StartTime:          decimal.NewFromInt(0x62a85db9).String(),
				EndTime:            decimal.NewFromInt(0x62a86bc9).String(),
				MinPercentageToAsk: "0",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x",
				R:                  "0xcd716aa453477a9ad595dff19e86a29e26da37ad7153545312d07de4be47f1ea",
				S:                  "0x19604de939790e29be08201395dbc8f1a8e77c54ee43a373a5ea78d90340752a",
				V:                  0x1b,
			},
			OrderItemHashes: [][]byte{hexutil.MustDecode("0xe231eaba29cb01088243493cff5ffa58cb96bbf0e5fb4d767a36ceb27209bf93"), hexutil.MustDecode("0xdd93d20d8e6ae6255f5e2d801d5c127a446f87c6dd554ebdd7de4b3634cc0d24")},
		},
		{
			// buy private sale
			// https://goerli.etherscan.io/tx/0x1ebb46b45674fd1597491a7dc8dc7d8a51cc00e1cc86df9febaa80215a2d8724
			// https://goerli.etherscan.io/tx/0xe69a8188342177f595a7f552c9e296a0ff7d58b09734ee7fa79a65c4542b0b63
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   true,
				Signer:  user1,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("3"), Amount: "1", Price: decimal.New(10000, 18).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("3"), Amount: "10", Price: decimal.New(10000, 18).String()},
				},
				Strategy:           strategyPrivateSale,
				Currency:           erc20Address,
				Nonce:              "3",
				StartTime:          decimal.NewFromInt(0x62a85dc8).String(),
				EndTime:            decimal.NewFromInt(0x62a86bd8).String(),
				MinPercentageToAsk: "9000",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x000000000000000000000000df8650b0ca1260f7a2f4fdff9082aede554f65ad",
				R:                  "0x8149fe5282ee9297bf064bc7f215c0ef458497724f11af30b0b15eb1cf6d9592",
				S:                  "0x35e4d56caeb1e6332f54b2699039b98776f2c14a8cc2035b25753eb8bc6b3535",
				V:                  0x1b,
			},
			OrderItemHashes: [][]byte{hexutil.MustDecode("0x79b346a730d10e9f91a30d6a21900519aac8c1e14e4404e8426ed8e0c8b125ad"), hexutil.MustDecode("0x2a6857ea7c283911409ba6cd2aa44724333d432d339c6a54fa9332402e867ed9")},
		},
		{
			// buy private sale with eth
			// https://goerli.etherscan.io/tx/0xdb8ea6726a708e9c8a25637e83a8b02d108b004df5a18d612ebe5f74a283a8a2
			// https://goerli.etherscan.io/tx/0x137115577fbae66b63521498592c034f585c12c94c92692515a54dfabf947600
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   true,
				Signer:  user1,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("4"), Amount: "1", Price: decimal.New(10000, 9).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("4"), Amount: "10", Price: decimal.New(10000, 9).String()},
				},
				Strategy:           strategyPrivateSale,
				Currency:           weth,
				Nonce:              "4",
				StartTime:          decimal.NewFromInt(0x62a85e13).String(),
				EndTime:            decimal.NewFromInt(0x62a86c23).String(),
				MinPercentageToAsk: "9000",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x000000000000000000000000df8650b0ca1260f7a2f4fdff9082aede554f65ad",
				R:                  "0x2e3245fc8e6856f5adc845f61a891458966c6bd2473df2857488bdcccc8e272b",
				S:                  "0x202a6e7a811530a23413fabb86b9232c0822806ba101a5596120e5b745a1df15",
				V:                  0x1b,
			},
			OrderItemHashes: [][]byte{hexutil.MustDecode("0x5d2119d95a891f233902f6e0d09ced9fb80cdf611943c2a22724146f1b9dcde8"), hexutil.MustDecode("0x61d509f87457293f563172bcbf332e88821b3eaaa3c84268816da29b1debbf4c")},
		},
		{
			// accept collection offer
			// https://goerli.etherscan.io/tx/0x4c5f04400add14cda7a7bf206041b019419e8e2814431ecb253f3587d25aaac8
			// https://goerli.etherscan.io/tx/0xc1a22fbe55a9ad4abe70008756834aae04f4157587405d0b9608e36189991948
			VerifyingContract: exchangeAddress,
			MakerOrder: order.Order{
				ChainId: domain.ChainId(5),
				IsAsk:   false,
				Signer:  user2,
				Items: []order.Item{
					{Collection: erc721Address, TokenId: domain.TokenId("0"), Amount: "1", Price: decimal.New(10000, 18).String()},
					{Collection: erc1155Address, TokenId: domain.TokenId("0"), Amount: "10", Price: decimal.New(10000, 18).String()},
				},
				Strategy:           strategyCollectionOffer,
				Currency:           erc20Address,
				Nonce:              "5",
				StartTime:          decimal.NewFromInt(0x62a85e40).String(),
				EndTime:            decimal.NewFromInt(0x62a86c50).String(),
				MinPercentageToAsk: "0",
				Marketplace:        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
				Params:             "0x",
				R:                  "0xa261ae089ef93e02622642ab894980e20e28fe7745b5f4811215d2e238901977",
				S:                  "0x60e5f4201121b7d59aee5f7f7870d3939be3a9c2506142f43bec85841cccc7a4",
				V:                  0x1b,
			},
			OrderItemHashes: [][]byte{hexutil.MustDecode("0x94be0941754d1a1530bac3ec3a98aa727962e5a3aa3b8a5020f185604a97a034"), hexutil.MustDecode("0x079e08ce518b3171b0b31e0d68dbf1a95282e7b17015e3537b5109e1c945345e")},
		},
	}
	for _, c := range cases {
		for i, h := range c.OrderItemHashes {
			hash, err := c.MakerOrder.HashOrderItem(i)
			req.NoError(err)
			req.Equal(h, hash)
		}
	}
}

type testSuite struct {
	suite.Suite

	query              query.Mongo
	pricefomatter      *pricefomatter.PriceFormatter
	exchangeCfgs       map[domain.ChainId]order.ExchangeCfg
	orderRepo          *mOrder.OrderRepo
	orderItemRepo      *mOrder.OrderItemRepo
	nftitemRepo        *mNftitem.Repo
	erc1155HoldingRepo *mErc1155.HoldingRepo
	accountRepo        account.Repo
	paytokenRepo       *mDomain.PayTokenRepo
	orderNonceUC       account.OrderNonceUseCase

	im *impl
}

func (s *testSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q
	s.exchangeCfgs = nil
	s.pricefomatter = new(pricefomatter.PriceFormatter)
	s.orderRepo = &mOrder.OrderRepo{}
	s.orderItemRepo = &mOrder.OrderItemRepo{}
	s.nftitemRepo = &mNftitem.Repo{}
	s.erc1155HoldingRepo = &mErc1155.HoldingRepo{}
	s.accountRepo = nil
	s.paytokenRepo = &mDomain.PayTokenRepo{}
	s.orderNonceUC = nil

	s.im = New(&OrderUseCaseCfg{
		s.exchangeCfgs,
		s.orderRepo,
		s.orderItemRepo,
		s.nftitemRepo,
		s.erc1155HoldingRepo,
		s.accountRepo,
		s.paytokenRepo,
		s.pricefomatter,
		s.orderNonceUC,
		nil,
		nil,
		nil,
	}).(*impl)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TearDownTest() {
	s.orderRepo.AssertExpectations(s.T())
	s.orderItemRepo.AssertExpectations(s.T())
	s.nftitemRepo.AssertExpectations(s.T())
	s.erc1155HoldingRepo.AssertExpectations(s.T())
	s.paytokenRepo.AssertExpectations(s.T())
}

func (s *testSuite) TestRefreshOrders() {
	mockNftitem := nftitem.NftItem{
		ChainId:         1,
		ContractAddress: "0x123",
		TokenId:         "1",
		TokenType:       domain.TokenType721,
		Owner:           "0x5566",
	}

	mockOrderItems := []*order.OrderItem{
		{
			ChainId: 1,
			Item: order.Item{
				Collection: "0x123",
				TokenId:    "1",
				Amount:     "1",
				Price:      "100",
			},
			ItemIdx:            0,
			OrderHash:          "orderhash1",
			OrderItemHash:      "orderItemhash1",
			IsAsk:              true,
			Signer:             "0x5566",
			Currency:           domain.EmptyAddress,
			StartTime:          time.Now().Add(-1 * time.Hour),
			EndTime:            time.Now().Add(1 * time.Hour),
			MinPercentageToAsk: "0",
			Marketplace:        "x",
			Strategy:           order.StrategyFixedPrice,
			ReservedBuyer:      "",
			PriceInUsd:         0,
			PriceInNative:      0,
			DisplayPrice:       "0",
		},
		{
			ChainId: 1,
			Item: order.Item{
				Collection: "0x123",
				TokenId:    "1",
				Amount:     "1",
				Price:      "200",
			},
			ItemIdx:            0,
			OrderHash:          "orderhash2",
			OrderItemHash:      "orderItemhash2",
			IsAsk:              true,
			Signer:             "0x1234",
			Currency:           domain.EmptyAddress,
			StartTime:          time.Now().Add(-1 * time.Hour),
			EndTime:            time.Now().Add(1 * time.Hour),
			MinPercentageToAsk: "0",
			Marketplace:        "x",
			Strategy:           order.StrategyFixedPrice,
			ReservedBuyer:      "",
			PriceInUsd:         0,
			PriceInNative:      0,
			DisplayPrice:       "0",
		},
		{
			ChainId: 1,
			Item: order.Item{
				Collection: "0x123",
				TokenId:    "1",
				Amount:     "1",
				Price:      "50",
			},
			ItemIdx:            0,
			OrderHash:          "orderhash3",
			OrderItemHash:      "orderItemhash3",
			IsAsk:              false,
			Signer:             "0x54321",
			Currency:           domain.EmptyAddress,
			StartTime:          time.Now().Add(-1 * time.Hour),
			EndTime:            time.Now().Add(1 * time.Hour),
			MinPercentageToAsk: "0",
			Marketplace:        "x",
			Strategy:           order.StrategyFixedPrice,
			ReservedBuyer:      "",
			PriceInUsd:         0,
			PriceInNative:      0,
			DisplayPrice:       "0",
		},
	}

	s.nftitemRepo.On("FindOne", mock.Anything, mockNftitem.ChainId, mockNftitem.ContractAddress, mockNftitem.TokenId).
		Return(&mockNftitem, nil).Once()

	s.orderItemRepo.On("FindAll", mock.Anything,
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc")).
		Return(mockOrderItems, nil).Once()

	s.pricefomatter.On("GetPrices", mock.Anything, mockOrderItems[0].ChainId, mockOrderItems[0].Currency, big.NewInt(100)).
		Return(decimal.NewFromInt(100), float64(10000), float64(100), nil).Once()
	s.pricefomatter.On("GetPrices", mock.Anything, mockOrderItems[1].ChainId, mockOrderItems[1].Currency, big.NewInt(200)).
		Return(decimal.NewFromInt(200), float64(20000), float64(200), nil).Once()
	s.pricefomatter.On("GetPrices", mock.Anything, mockOrderItems[2].ChainId, mockOrderItems[2].Currency, big.NewInt(50)).
		Return(decimal.NewFromInt(50), float64(5000), float64(50), nil).Once()

	s.orderItemRepo.On("Update", mock.Anything, mockOrderItems[0].ToId(), order.OrderItemPatchable{
		IsValid:       ptr.Bool(true),
		DisplayPrice:  ptr.String(decimal.NewFromInt(100).String()),
		PriceInUsd:    ptr.Float64(10000),
		PriceInNative: ptr.Float64(100),
	}).Return(nil).Once()

	s.orderItemRepo.On("Update", mock.Anything, mockOrderItems[1].ToId(), order.OrderItemPatchable{
		IsValid:       ptr.Bool(false),
		DisplayPrice:  ptr.String(decimal.NewFromInt(200).String()),
		PriceInUsd:    ptr.Float64(20000),
		PriceInNative: ptr.Float64(200),
	}).Return(nil).Once()

	s.orderItemRepo.On("Update", mock.Anything, mockOrderItems[2].ToId(), order.OrderItemPatchable{
		IsValid:       ptr.Bool(true),
		DisplayPrice:  ptr.String(decimal.NewFromInt(50).String()),
		PriceInUsd:    ptr.Float64(5000),
		PriceInNative: ptr.Float64(50),
	}).Return(nil).Once()

	err := s.im.RefreshOrders(ctx.Background(), *mockNftitem.ToId())
	s.Nil(err)

}
