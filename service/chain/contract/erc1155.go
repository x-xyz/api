package contract

import (
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	baseabi "github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/chain"
)

type Erc1155Contract interface {
	Supports1155Interface(ctx bCtx.Ctx, chainId int32, addr string) (bool, error)
}

type Erc1155 struct {
	chainService       chain.Client
	abi                ethabi.ABI
	erc1155InterfaceId [4]byte
}

func NewErc1155(chainService chain.Client) *Erc1155 {
	var interfaceId [4]byte
	copy(interfaceId[:], common.Hex2Bytes("d9b67a26"))
	return &Erc1155{
		abi:                baseabi.ERC1155TokenABI,
		chainService:       chainService,
		erc1155InterfaceId: interfaceId,
	}
}

func (e *Erc1155) Supports1155Interface(ctx bCtx.Ctx, chainId int32, addr string) (bool, error) {
	method := "supportsInterface"
	unpacked, err := e.chainService.Call(ctx, chainId, common.HexToAddress(addr), nil, e.abi, method, e.erc1155InterfaceId)
	if err != nil {
		return false, err
	}
	return unpacked[0].(bool), nil
}
