package chain

import (
	"strings"

	"github.com/x-xyz/goapi/domain"
)

type TokenInfo struct {
	Symbol   string
	Decimals int
}

type TokenToText = map[domain.Address]TokenInfo

var (
	chainIdToText = map[domain.ChainId]string{
		domain.ChainId(1):   "ethereum",
		domain.ChainId(3):   "ropsten",
		domain.ChainId(5):   "goerli",
		domain.ChainId(56):  "binance-smart-chain",
		domain.ChainId(97):  "binance-smart-chain-testnet",
		domain.ChainId(250): "fantom",
	}
)

func GetChainUrlPart(chainId domain.ChainId) (string, error) {
	if val, err := GetChainDisplayName(chainId); err != nil {
		return "", err
	} else {
		return strings.ToLower(strings.Replace(val, " ", "-", 0)), nil
	}
}

func GetChainDisplayName(chainId domain.ChainId) (string, error) {
	if val, ok := chainIdToText[chainId]; !ok {
		return "", domain.ErrNotFound
	} else {
		return val, nil
	}
}
