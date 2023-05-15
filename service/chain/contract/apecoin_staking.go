package contract

import (
	"math/big"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	baseabi "github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/chain"
)

type ApecoinStakingContract interface {
	NftPosition(ctx bCtx.Ctx, poolId, tokenId *big.Int) (*big.Int, *big.Int, error)
}

type apecoinStaking struct {
	chainService chain.Client
	abi          ethabi.ABI
	chainId      int32
	addr         common.Address
}

func NewApecoinStaking(chainService chain.Client, chainId int32, addr common.Address) ApecoinStakingContract {
	return &apecoinStaking{
		abi:          baseabi.ApecoinStakingABI,
		chainService: chainService,
		chainId:      chainId,
		addr:         addr,
	}
}

func (e *apecoinStaking) NftPosition(ctx bCtx.Ctx, poolId, tokenId *big.Int) (*big.Int, *big.Int, error) {
	method := "nftPosition"
	unpacked, err := e.chainService.Call(ctx, e.chainId, e.addr, nil, e.abi, method, poolId, tokenId)
	if err != nil {
		return nil, nil, err
	}
	return unpacked[0].(*big.Int), unpacked[1].(*big.Int), nil
}
