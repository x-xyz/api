package nftdetector

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type NFTType int

const (
	Erc721Type NFTType = iota
	Erc1155Type
)

type UseCase interface {
	DetectNFT(ctx.Ctx, domain.ChainId, domain.Address, NFTType) error
}
