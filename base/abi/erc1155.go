package abi

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var ERC1155TokenABI abi.ABI

var erc1155ABI = `[{"type":"event","anonymous":false,"name":"TransferSingle","inputs":[{"type":"address","name":"_operator","indexed":true},{"type":"address","name":"_from","indexed":true},{"type":"address","name":"_to","indexed":true},{"type":"uint256","name":"_id"},{"type":"uint256","name":"_value"}]},{"type":"event","anonymous":false,"name":"TransferBatch","inputs":[{"type":"address","name":"_operator","indexed":true},{"type":"address","name":"_from","indexed":true},{"type":"address","name":"_to","indexed":true},{"type":"uint256[]","name":"_ids"},{"type":"uint256[]","name":"_values"}]},{"type":"function","name":"supportsInterface","constant":true,"stateMutability":"view","payable":false,"inputs":[{"type":"bytes4","name":"interfaceID"}],"outputs":[{"type":"bool"}]},{"type":"function","name":"uri","constant":true,"stateMutability":"view","payable":false,"inputs":[{"type":"uint256","name":"_id"}],"outputs":[{"type":"string"}]}]`

func init() {
	_abi, err := abi.JSON(strings.NewReader(erc1155ABI))
	if err != nil {
		panic("Failed to parse erc1155 abi")
	}
	ERC1155TokenABI = _abi
}

type Erc1155TransferSingleLog struct {
	Operator common.Address // indexed
	From     common.Address // indexed
	To       common.Address // indexed
	Id       *big.Int
	Value    *big.Int
}

type Erc1155TransferBatchLog struct {
	Operator common.Address // indexed
	From     common.Address // indexed
	To       common.Address // indexed
	Ids      []*big.Int
	Values   []*big.Int
}

func ToErc1155TransferSingleLog(log *types.Log) (*Erc1155TransferSingleLog, error) {
	var transferSingle Erc1155TransferSingleLog
	if err := ERC1155TokenABI.UnpackIntoInterface(&transferSingle, "TransferSingle", log.Data); err != nil {
		return nil, err
	}
	transferSingle.Operator = common.BytesToAddress(log.Topics[1].Bytes())
	transferSingle.From = common.BytesToAddress(log.Topics[2].Bytes())
	transferSingle.To = common.BytesToAddress(log.Topics[3].Bytes())
	return &transferSingle, nil
}

func ToErc1155TransferBatchLog(log *types.Log) (*Erc1155TransferBatchLog, error) {
	var transferBatch Erc1155TransferBatchLog
	if err := ERC1155TokenABI.UnpackIntoInterface(&transferBatch, "TransferBatch", log.Data); err != nil {
		return nil, err
	}
	transferBatch.Operator = common.BytesToAddress(log.Topics[1].Bytes())
	transferBatch.From = common.BytesToAddress(log.Topics[2].Bytes())
	transferBatch.To = common.BytesToAddress(log.Topics[3].Bytes())
	return &transferBatch, nil
}
