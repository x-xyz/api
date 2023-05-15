package usecase

import (
	"fmt"
	"strings"

	bctx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/erc1155"
	erc721 "github.com/x-xyz/goapi/domain/erc721/contract"
	"github.com/x-xyz/goapi/domain/nftdetector"
	chainservice "github.com/x-xyz/goapi/service/chain/contract"
)

type NFTDetectorCfg struct {
	Erc721         erc721.Repo
	Erc721Service  chainservice.Erc721Contract
	Erc1155        erc1155.Repo
	Erc1155Service chainservice.Erc1155Contract
}

type nftDetectorUseCase struct {
	seen           map[string]struct{}
	erc721         erc721.Repo
	erc721Service  chainservice.Erc721Contract
	erc1155        erc1155.Repo
	erc1155Service chainservice.Erc1155Contract
}

func NewNFTDetectorUseCase(cfg *NFTDetectorCfg) nftdetector.UseCase {
	return &nftDetectorUseCase{
		seen:           map[string]struct{}{}, // NOTE: not goroutine safe
		erc721:         cfg.Erc721,
		erc721Service:  cfg.Erc721Service,
		erc1155:        cfg.Erc1155,
		erc1155Service: cfg.Erc1155Service,
	}
}

func (n *nftDetectorUseCase) DetectNFT(ctx bctx.Ctx, chainId domain.ChainId, address domain.Address, nftType nftdetector.NFTType) error {
	if _, ok := n.seen[address.ToLowerStr()]; ok {
		return nil
	}

	var err error
	switch nftType {
	case nftdetector.Erc721Type:
		err = n.detectErc721(ctx, chainId, address)
	case nftdetector.Erc1155Type:
		err = n.detectErc1155(ctx, chainId, address)
	}

	if err == nil {
		n.seen[address.ToLowerStr()] = struct{}{}
	}
	return err
}

func (n *nftDetectorUseCase) detectErc721(ctx bctx.Ctx, chainId domain.ChainId, address domain.Address) error {
	_, err := n.erc721.FindOne(ctx,
		erc721.WithChainId(chainId),
		erc721.WithAddress(address),
	)

	// db error
	if err != nil && err != domain.ErrNotFound {
		return err
	}

	// already exists
	if err != domain.ErrNotFound {
		return nil
	}

	// check erc721
	is721, err := n.erc721Service.Supports721Interface(ctx, int32(chainId), address.ToLowerStr())
	if err != nil {
		if n.isKnownEVMError(err) {
			return nil
		}
		ctx.Error(fmt.Sprintf("detect erc721 = %s err=%s", address, err.Error()))
		return err
	}

	if !is721 {
		return nil
	}

	return n.erc721.Create(ctx, erc721.Contract{
		ChainId: chainId,
		Address: address.ToLower(),
	})
}

func (n *nftDetectorUseCase) detectErc1155(ctx bctx.Ctx, chainId domain.ChainId, address domain.Address) error {
	ctx.Info(fmt.Sprintf("detect erc1155 = %s", address))
	_, err := n.erc1155.FindOne(ctx,
		erc1155.WithChainId(chainId),
		erc1155.WithAddress(address),
	)

	// db error
	if err != nil && err != domain.ErrNotFound {
		return err
	}

	// already exists
	if err != domain.ErrNotFound {
		return nil
	}

	// check erc721
	is1155, err := n.erc1155Service.Supports1155Interface(ctx, int32(chainId), address.ToLowerStr())
	if err != nil {
		if n.isKnownEVMError(err) {
			return nil
		}
		ctx.Error(fmt.Sprintf("detect erc1155 = %s err=%s", address, err.Error()))
		return err
	}

	if !is1155 {
		return nil
	}

	return n.erc1155.Create(ctx, erc1155.Contract{
		ChainId: chainId,
		Address: address.ToLower(),
	})
}

func (n *nftDetectorUseCase) isKnownEVMError(err error) bool {
	if strings.Contains(err.Error(), "execution reverted") {
		return true
	}
	if strings.Contains(err.Error(), "abi: attempting to unmarshall an empty string while arguments are expected") {
		return true
	}
	if strings.Contains(err.Error(), "invalid opcode: INVALID") {
		return true
	}
	if strings.Contains(err.Error(), "invalid jump destination") {
		return true
	}
	return false
}
