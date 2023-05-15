package contract

import (
	"math/big"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	baseabi "github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/chain"
)

type Erc721Contract interface {
	Supports721Interface(ctx bCtx.Ctx, chainId int32, addr string) (bool, error)
}

type Erc721 struct {
	chainService      chain.Client
	abi               ethabi.ABI
	erc721InterfaceId [4]byte
}

func NewErc721(chainService chain.Client) *Erc721 {
	var interfaceId [4]byte
	copy(interfaceId[:], common.Hex2Bytes("80ac58cd"))
	return &Erc721{
		abi:               baseabi.ERC721TokenABI,
		chainService:      chainService,
		erc721InterfaceId: interfaceId,
	}
}

func (e *Erc721) Supports721Interface(ctx bCtx.Ctx, chainId int32, addr string) (bool, error) {
	method := "supportsInterface"
	unpacked, err := e.chainService.Call(ctx, chainId, common.HexToAddress(addr), nil, e.abi, method, e.erc721InterfaceId)
	if err != nil {
		return false, err
	}
	return unpacked[0].(bool), nil
}

func (e *Erc721) OwnerOf(ctx bCtx.Ctx, chainId int32, addr string, tokenId *big.Int) (string, error) {
	method := "ownerOf"
	unpacked, err := e.chainService.Call(ctx, chainId, common.HexToAddress(addr), nil, e.abi, method, tokenId)
	if err != nil {
		return "", err
	}
	return unpacked[0].(common.Address).String(), nil
}
