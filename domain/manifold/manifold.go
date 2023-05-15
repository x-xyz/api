package manifold

import (
	"github.com/x-xyz/goapi/domain"
)

type RoyaltyOverrideEvent struct {
	Owner          domain.Address
	TokenAddress   domain.Address
	RoyaltyAddress domain.Address
}
