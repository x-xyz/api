package tracker

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/apecoinstaking"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/token"
	"github.com/x-xyz/goapi/service/chain/contract"
)

var (
	depositNftSig      = abi.ApecoinStakingABI.Events["DepositNft"].ID
	depositPairNftSig  = abi.ApecoinStakingABI.Events["DepositPairNft"].ID
	withdrawNftSig     = abi.ApecoinStakingABI.Events["WithdrawNft"].ID
	withdrawPairNftSig = abi.ApecoinStakingABI.Events["WithdrawPairNft"].ID
	bakcPoolId         = big.NewInt(3)
)

var poolMapping = map[string]domain.Address{
	"1": "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d", // BAYC
	"2": "0x60e4d786628fea6478f785a6d7e704777c86a7c6", // MAYC
	"3": "0xba30e5f9bb24caa003e9f2f0497ad287fdf95623", // BAKC
}

type ApecoinStakingEventHandlerCfg struct {
	ChainId                int64
	ApecoinStakingUC       apecoinstaking.UseCase
	ApecoinStakingContract contract.ApecoinStakingContract
	TokenUC                token.Usecase
}

type ApecoinStakingEventHandler struct {
	chainId                int64
	apecoinStakingUC       apecoinstaking.UseCase
	apecoinStakingContract contract.ApecoinStakingContract
	tokenUC                token.Usecase
}

func NewApecoinStakingEventHandler(cfg *ApecoinStakingEventHandlerCfg) EventHandler {
	return &ApecoinStakingEventHandler{
		chainId:                cfg.ChainId,
		apecoinStakingUC:       cfg.ApecoinStakingUC,
		apecoinStakingContract: cfg.ApecoinStakingContract,
		tokenUC:                cfg.TokenUC,
	}
}

func (h *ApecoinStakingEventHandler) GetFilterTopics() [][]common.Hash {
	return [][]common.Hash{{depositNftSig, depositPairNftSig, withdrawNftSig, withdrawPairNftSig}}
}

func (h *ApecoinStakingEventHandler) ProcessEvents(ctx bCtx.Ctx, logs []logWithBlockTime) error {
	for _, _log := range logs {
		switch _log.Topics[0] {
		case depositNftSig:
			e, err := abi.ToDepositNftLog(&_log.Log)
			if err != nil {
				ctx.WithFields(log.Fields{"l": _log.Log, "err": err}).Error("ToDepositNftLog failed")
				return err
			}
			if err := h.updateToken(ctx, e.PoolId, e.TokenId); err != nil {
				ctx.WithFields(log.Fields{"poolId": e.PoolId, "tokenId": e.TokenId, "err": err}).Error("updateToken failed")
				return err
			}
		case depositPairNftSig:
			e, err := abi.ToDepositPairNftLog(&_log.Log)
			if err != nil {
				ctx.WithFields(log.Fields{"l": _log.Log, "err": err}).Error("ToDepositPairNftLog failed")
				return err
			}
			if err := h.updateToken(ctx, e.MainTypePoolId, e.MainTokenId); err != nil {
				ctx.WithFields(log.Fields{"poolId": e.MainTypePoolId, "tokenId": e.MainTokenId, "err": err}).Error("updateToken failed")
				return err
			}
			if err := h.updateToken(ctx, bakcPoolId, e.BakcTokenId); err != nil {
				ctx.WithFields(log.Fields{"poolId": bakcPoolId, "tokenId": e.BakcTokenId, "err": err}).Error("updateToken failed")
				return err
			}
		case withdrawNftSig:
			e, err := abi.ToWithdrawNftLog(&_log.Log)
			if err != nil {
				ctx.WithFields(log.Fields{"l": _log.Log, "err": err}).Error("ToWithdrawNftLog failed")
				return err
			}
			if err := h.updateToken(ctx, e.PoolId, e.TokenId); err != nil {
				ctx.WithFields(log.Fields{"poolId": e.PoolId, "tokenId": e.TokenId, "err": err}).Error("updateToken failed")
				return err
			}
		case withdrawPairNftSig:
			e, err := abi.ToWithdrawPairNftLog(&_log.Log)
			if err != nil {
				ctx.WithFields(log.Fields{"l": _log.Log, "err": err}).Error("ToWithdrawPairNftLog failed")
				return err
			}
			if err := h.updateToken(ctx, e.MainTypePoolId, e.MainTokenId); err != nil {
				ctx.WithFields(log.Fields{"poolId": e.MainTypePoolId, "tokenId": e.MainTokenId, "err": err}).Error("updateToken failed")
				return err
			}
			if err := h.updateToken(ctx, bakcPoolId, e.BakcTokenId); err != nil {
				ctx.WithFields(log.Fields{"poolId": bakcPoolId, "tokenId": e.BakcTokenId, "err": err}).Error("updateToken failed")
				return err
			}
		default:
			ctx.WithField("topic", _log.Topics[0]).Warn("unknown topic, skipping")
		}
	}
	return nil
}

func (h *ApecoinStakingEventHandler) updateToken(ctx bCtx.Ctx, poolId, tokenId *big.Int) error {
	addr, ok := poolMapping[poolId.String()]
	if !ok {
		return nil
	}
	staked, _, err := h.apecoinStakingContract.NftPosition(ctx, poolId, tokenId)
	if err != nil {
		return err
	}
	err = h.apecoinStakingUC.Upsert(ctx, &apecoinstaking.ApecoinStaking{
		ChainId:         domain.ChainId(h.chainId),
		ContractAddress: addr,
		TokenId:         domain.TokenId(tokenId.String()),
		Staked:          staked.Cmp(common.Big0) != 0,
	})
	if err != nil {
		return err
	}

	nft, err := h.tokenUC.FindOne(ctx, nftitem.Id{
		ChainId:         domain.ChainId(h.chainId),
		ContractAddress: addr,
		TokenId:         domain.TokenId(tokenId.String()),
	})
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	if err == domain.ErrNotFound {
		return nil
	}
	if nft.IndexerState == nftitem.IndexerStateDone || nft.IndexerState == nftitem.IndexerStateFetchingAnimation {
		patchable := &nftitem.PatchableNftItem{
			IndexerState:      (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateParsingAttributes)),
			IndexerRetryCount: ptr.Int32(0),
		}
		if err := h.tokenUC.PatchNft(ctx, nft.ToId(), patchable); err != nil {
			ctx.WithField("err", err).Error("token.PatchNft failed")
			return err
		}
	}
	return nil
}
