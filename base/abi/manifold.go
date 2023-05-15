package abi

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"strings"
)

type RoyaltyOverrideLog struct {
	Owner          common.Address
	TokenAddress   common.Address
	RoyaltyAddress common.Address
}

var ManifoldABI abi.ABI

func init() {
	_abi, err := abi.JSON(strings.NewReader(manifoldABIJson))
	if err != nil {
		panic("Failed to parse ABI")
	}
	ManifoldABI = _abi
}

func ToRoyaltyOverrideLog(log *types.Log) (*RoyaltyOverrideLog, error) {
	var l RoyaltyOverrideLog
	if err := ManifoldABI.UnpackIntoInterface(&l, "RoyaltyOverride", log.Data); err != nil {
		return nil, err
	}
	return &l, nil
}

var manifoldABIJson = `
[
	{
		"anonymous":false,
		"inputs":[
		{
			"indexed":false,
            "internalType":"address",
            "name":"owner",
            "type":"address"
		},
		{
			"indexed":false,
            "internalType":"address",
            "name":"tokenAddress",
            "type":"address"
         },
		{
            "indexed":false,
            "internalType":"address",
            "name":"royaltyAddress",
            "type":"address"
		}
		],
		"name":"RoyaltyOverride",
		"type":"event"
	}
]
`
