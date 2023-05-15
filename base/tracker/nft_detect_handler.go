package tracker

import (
	"github.com/ethereum/go-ethereum/common"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftdetector"
)

type NFTDetectHandlerCfg struct {
	ChainId          int64
	NFTDetectUseCase nftdetector.UseCase
}

type NFTDetectHandler struct {
	NFTDetectHandlerCfg
	Topics [][]common.Hash
}

func NewNFTDetectHandler(cfg *NFTDetectHandlerCfg) EventHandler {
	return &NFTDetectHandler{
		NFTDetectHandlerCfg: *cfg,
		Topics: [][]common.Hash{
			{transferSig, transferSingle, transferBatch},
		},
	}
}

func (h *NFTDetectHandler) GetFilterTopics() [][]common.Hash {
	return h.Topics
}

func (h *NFTDetectHandler) ProcessEvents(ctx bCtx.Ctx, logs []logWithBlockTime) error {
	for _, log := range logs {
		meta := toLogMeta(&log)
		switch log.Topics[0] {
		case transferSig:
			err := h.NFTDetectUseCase.DetectNFT(ctx, domain.ChainId(h.ChainId), meta.ContractAddress, nftdetector.Erc721Type)
			if err != nil {
				ctx.WithField("err", err).Error("detect erc721 failed")
				return err
			}
		case transferSingle, transferBatch:
			err := h.NFTDetectUseCase.DetectNFT(ctx, domain.ChainId(h.ChainId), meta.ContractAddress, nftdetector.Erc1155Type)
			if err != nil {
				ctx.WithField("err", err).Error("detect erc1155 failed")
				return err
			}
		default:
			ctx.WithField("topic", log.Topics[0]).Warn("unknown topic, skipping")
		}
	}
	return nil
}
