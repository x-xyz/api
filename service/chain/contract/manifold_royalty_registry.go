package contract

import (
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	baseabi "github.com/x-xyz/goapi/base/abi"
	"github.com/x-xyz/goapi/service/chain"
)

type ManifoldRoyaltyEngine struct {
	chainService chain.Client
	abi          ethabi.ABI
}

func NewManifoldRoyaltyEngine(chainService chain.Client) *ManifoldRoyaltyEngine {
	return &ManifoldRoyaltyEngine{
		abi:          baseabi.ManifoldABI,
		chainService: chainService,
	}
}
