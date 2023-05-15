package abi

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var ERC1271ABI abi.ABI

func init() {
	_abi, err := abi.JSON(strings.NewReader(erc1271ABIJson))
	if err != nil {
		panic("Failed to parse ABI")
	}
	ERC1271ABI = _abi
}

var erc1271ABIJson = `
[
  {
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "_hash",
        "type": "bytes32"
      },
      {
        "internalType": "bytes",
        "name": "_signature",
        "type": "bytes"
      }
    ],
    "name": "isValidSignature",
    "outputs": [
      {
        "internalType": "bytes4",
        "name": "magicValue",
        "type": "bytes4"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]

`
