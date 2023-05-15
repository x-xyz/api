package contract

import (
	"math/big"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	baseabi "github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/chain"
)

type RoyaltyEngineContract interface {
	GetRoyalty(ctx bCtx.Ctx, chainId int32, addr string, collection string, tokenId *big.Int, value *big.Int) ([]string, []*big.Int, error)
}

type RoyaltyEngine struct {
	chainService chain.Client
	abi          ethabi.ABI
}

func NewRoyaltyEngine(chainService chain.Client) RoyaltyEngineContract {
	return &RoyaltyEngine{
		abi:          baseabi.RoyaltyEngineABI,
		chainService: chainService,
	}
}

func (r *RoyaltyEngine) GetRoyalty(ctx bCtx.Ctx, chainId int32, addr string, collection string, tokenId *big.Int, value *big.Int) ([]string, []*big.Int, error) {
	method := "getRoyalty"
	unpacked, err := r.chainService.Call(ctx, chainId, common.HexToAddress(addr), nil, r.abi, method, common.HexToAddress(collection), tokenId, value)
	if err != nil {
		return nil, nil, err
	}
	recipients := unpacked[0].([]common.Address)
	var _recipeints []string
	for _, recipient := range recipients {
		_recipeints = append(_recipeints, recipient.String())
	}
	return _recipeints, unpacked[1].([]*big.Int), nil
}

// func (e *Erc721) Supports721Interface(ctx bCtx.Ctx, chainId int32, addr string) (bool, error) {
// 	method := "supportsInterface"
// 	unpacked, err := e.chainService.Call(ctx, chainId, common.HexToAddress(addr), nil, e.abi, method, e.erc721InterfaceId)
// 	if err != nil {
// 		return false, err
// 	}
// 	return unpacked[0].(bool), nil
// }

// func (e *Erc721) OwnerOf(ctx bCtx.Ctx, chainId int32, addr string, tokenId *big.Int) (string, error) {
// 	method := "ownerOf"
// 	unpacked, err := e.chainService.Call(ctx, chainId, common.HexToAddress(addr), nil, e.abi, method, tokenId)
// 	if err != nil {
// 		return "", err
// 	}
// 	return unpacked[0].(common.Address).String(), nil
// }
