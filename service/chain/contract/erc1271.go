package contract

import (
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	baseabi "github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/chain"
)

type Erc1271Contract interface {
	IsValidSignature(ctx bCtx.Ctx, chainId int32, addr string, hash common.Hash, signature []byte) (bool, error)
}

type Erc1271 struct {
	chainService chain.Client
	abi          ethabi.ABI
	magicValue   [4]byte
}

func NewErc1271(chainService chain.Client) Erc1271Contract {
	var magicValue [4]byte
	copy(magicValue[:], common.Hex2Bytes("1626ba7e"))
	return &Erc1271{
		abi:          baseabi.ERC1271ABI,
		chainService: chainService,
		magicValue:   magicValue,
	}
}

func (e *Erc1271) IsValidSignature(ctx bCtx.Ctx, chainId int32, addr string, hash common.Hash, signature []byte) (bool, error) {
	method := "isValidSignature"
	unpacked, err := e.chainService.Call(ctx, chainId, common.HexToAddress(addr), nil, e.abi, method, hash, signature)
	if err != nil {
		return false, err
	}
	return unpacked[0].([4]byte) == e.magicValue, nil
}
