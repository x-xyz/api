package abi

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

func TestApecoinStakingLogParsing(t *testing.T) {
	req := require.New(t)
	ctx := bCtx.Background()
	client, err := ethclient.Dial("rpc_url")
	req.NoError(err)

	{
		filter := ethereum.FilterQuery{
			Topics:    [][]common.Hash{{ApecoinStakingABI.Events["DepositNft"].ID}},
			FromBlock: big.NewInt(16746476),
			ToBlock:   big.NewInt(16746476),
		}
		logs, err := client.FilterLogs(ctx, filter)
		req.NoError(err)
		req.Equal(1, len(logs))
		l, err := ToDepositNftLog(&logs[0])
		req.NoError(err)
		expected := &DepositNftLog{
			User:    common.HexToAddress("0x5324a98b506F3265c500f978F3943A1fC6A55fa4"),
			PoolId:  big.NewInt(1),
			Amount:  decimal.RequireFromString("62357001106192178781").BigInt(),
			TokenId: big.NewInt(9638),
		}
		req.Equal(expected, l)
	}
	{
		filter := ethereum.FilterQuery{
			Topics:    [][]common.Hash{{ApecoinStakingABI.Events["WithdrawNft"].ID}},
			FromBlock: big.NewInt(16746736),
			ToBlock:   big.NewInt(16746736),
		}
		logs, err := client.FilterLogs(ctx, filter)
		req.NoError(err)
		req.Equal(1, len(logs))
		l, err := ToWithdrawNftLog(&logs[0])
		req.NoError(err)
		expected := &WithdrawNftLog{
			User:      common.HexToAddress("0x9438c455b9fC72A71Ad3225e8625Ec66Eb74CfAD"),
			PoolId:    big.NewInt(1),
			Amount:    decimal.RequireFromString("10094000000000000000000").BigInt(),
			Recipient: common.HexToAddress("0x9438c455b9fC72A71Ad3225e8625Ec66Eb74CfAD"),
			TokenId:   big.NewInt(486),
		}
		req.Equal(expected, l)
	}
	{
		filter := ethereum.FilterQuery{
			Topics:    [][]common.Hash{{ApecoinStakingABI.Events["DepositPairNft"].ID}},
			FromBlock: big.NewInt(16746755),
			ToBlock:   big.NewInt(16746755),
		}
		logs, err := client.FilterLogs(ctx, filter)
		req.NoError(err)
		req.Equal(1, len(logs))
		l, err := ToDepositPairNftLog(&logs[0])
		req.NoError(err)
		expected := &DepositPairNftLog{
			User:           common.HexToAddress("0xfB44Bb953cd20a2db39427c6039b95B6BBa0f1C1"),
			Amount:         decimal.RequireFromString("856000000000000000000").BigInt(),
			MainTypePoolId: big.NewInt(1),
			MainTokenId:    big.NewInt(259),
			BakcTokenId:    big.NewInt(9801),
		}
		req.Equal(expected, l)
	}
	{
		filter := ethereum.FilterQuery{
			Topics:    [][]common.Hash{{ApecoinStakingABI.Events["WithdrawPairNft"].ID}},
			FromBlock: big.NewInt(16746772),
			ToBlock:   big.NewInt(16746772),
		}
		logs, err := client.FilterLogs(ctx, filter)
		req.NoError(err)
		req.Equal(1, len(logs))
		l, err := ToWithdrawPairNftLog(&logs[0])
		req.NoError(err)
		expected := &WithdrawPairNftLog{
			User:           common.HexToAddress("0x822d3c3D8ed080a041f861c2476f583E234920BB"),
			Amount:         decimal.RequireFromString("856000000000000000000").BigInt(),
			MainTypePoolId: big.NewInt(1),
			MainTokenId:    big.NewInt(3868),
			BakcTokenId:    big.NewInt(9240),
		}
		req.Equal(expected, l)
	}
}
