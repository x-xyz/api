package tracker

import (
	"fmt"
	"math/big"
)

var (
	big1 = big.NewInt(1)
	big2 = big.NewInt(2)
)

type blockRange struct {
	begin *big.Int
	end   *big.Int // inclusive
}

func newBlockRange(begin, end uint64) *blockRange {
	return &blockRange{
		begin: new(big.Int).SetUint64(begin),
		end:   new(big.Int).SetUint64(end),
	}
}

func (r *blockRange) split() (*blockRange, *blockRange) {
	mid := new(big.Int).Add(r.begin, r.end)
	mid.Div(mid, big2)
	midP1 := new(big.Int).Add(mid, big1)
	first := &blockRange{begin: r.begin, end: mid}
	second := &blockRange{begin: midP1, end: r.end}
	return first, second
}

func (r *blockRange) String() string {
	return fmt.Sprintf("blockRange{%s-%s}", r.begin.String(), r.end.String())
}
