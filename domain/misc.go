package domain

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"golang.org/x/xerrors"
)

var (
	Big1  = big.NewInt(1)
	Big2  = big.NewInt(2)
	Big10 = big.NewInt(10)
)

type SortDir int8

const (
	SortDirAsc  = 1
	SortDirDesc = -1
)

type TokenType int

const (
	TokenType721  TokenType = 721
	TokenType1155 TokenType = 1155
	TokenTypePunk TokenType = 100000
)

type ChainId int32

type Address string

const EmptyAddress = Address("0x0000000000000000000000000000000000000000")
const PunkAddress = Address("0xb47e3cd837ddf8e4c57f05d70ab865de6e193bbb")
const PunkDataAddress = Address("0x16f5a35647d6f03d5d3da7b35409d65ba03af3b2")

func (a Address) ToLower() Address {
	return Address(strings.ToLower(string(a)))
}

func (a Address) ToLowerPtr() *Address {
	res := a.ToLower()
	return &res
}

func (a Address) ToLowerStr() string {
	return strings.ToLower(string(a))
}

func (a Address) IsEmpty() bool {
	return len(a) == 0
}

func (a Address) Equals(b Address) bool {
	return a.ToLowerStr() == b.ToLowerStr()
}

type TokenId string

func (i TokenId) String() string {
	return string(i)
}

func (i TokenId) ToHexString() (string, error) {
	id, ok := new(big.Int).SetString(i.String(), 10)
	if !ok {
		return "", xerrors.Errorf("invalid id %s", i)
	}
	return fmt.Sprintf("%064x", id), nil
}

type BlockNumber uint64

type TxHash string

type BlockHash string

type SourceId string

type TimePeriod string

type OrderHash string

type MarketplaceHash string

func (h OrderHash) ToLower() OrderHash {
	return OrderHash(strings.ToLower(string(h)))
}

const (
	TimePeriodDay    TimePeriod = "day"
	TimePeriodWeek              = "week"
	TimePeriod2Weeks            = "2weeks"
	TimePeriodMonth             = "month"
	TimePeriod2Month            = "2month"
	TimePeriodYear              = "year"
	TimePeriodAll               = "all"
)

var timePeriodToDuration = map[TimePeriod]time.Duration{
	TimePeriodDay:    1 * 24 * time.Hour,
	TimePeriodWeek:   7 * 24 * time.Hour,
	TimePeriod2Weeks: 14 * 24 * time.Hour,
	TimePeriodMonth:  30 * 24 * time.Hour,
	TimePeriod2Month: 60 * 24 * time.Hour,
	TimePeriodYear:   365 * 24 * time.Hour,
	TimePeriodAll:    time.Duration(1<<63 - 1), // max duration
}

func (tp TimePeriod) ToDuration() time.Duration {
	days, ok := timePeriodToDuration[tp]
	if !ok {
		return timePeriodToDuration[TimePeriodDay]
	}
	return days
}

func (tp TimePeriod) IsValid() bool {
	_, ok := timePeriodToDuration[tp]
	return ok
}

func (tp TimePeriod) IsAll() bool {
	return tp == TimePeriodAll
}

func ToBigInt(nums []string) ([]*big.Int, error) {
	var bns []*big.Int
	for _, n := range nums {
		bn, ok := new(big.Int).SetString(n, 10)
		if !ok {
			return nil, ErrInvalidNumberFormat
		}
		bns = append(bns, bn)
	}
	return bns, nil
}

var ChainIdWrappedNativeMap map[ChainId]Address = map[ChainId]Address{
	// eth
	1: "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
	// ropsten
	3: "0x0a180a76e4466bf68a7f86fb029bed3cccfaaac5",
	// goerli
	5: "0xb4fbf271143f4fbf7b91a5ded31805e42b2208d6",
	// bsc
	56: "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c",
	// bsc testnet
	97: "0xae13d989dac2f0debff460ac112a837c89baa7cd",
	// fantom
	250: "0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83",
}
