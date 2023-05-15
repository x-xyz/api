package order

import "github.com/x-xyz/goapi/domain"

type ExchangeCfg struct {
	Address    domain.Address
	Strategies map[domain.Address]Strategy
}
