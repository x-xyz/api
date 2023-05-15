package tracker

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/erc721/contract"
)

var transferSig = abi.ERC721TokenABI.Events["Transfer"].ID

type Erc721EventHandlerCfg struct {
	ChainId            int64
	Erc721EventUseCase contract.Erc721EventUseCase
}

type Erc721EventHandler struct {
	chainId       int64
	erc721EventUC contract.Erc721EventUseCase
}

func NewErc721EventHandler(cfg *Erc721EventHandlerCfg) EventHandler {
	return &Erc721EventHandler{
		chainId:       cfg.ChainId,
		erc721EventUC: cfg.Erc721EventUseCase,
	}
}

func (h *Erc721EventHandler) GetFilterTopics() [][]common.Hash {
	return [][]common.Hash{
		{
			transferSig,
		},
	}
}

func (h *Erc721EventHandler) ProcessEvents(ctx bCtx.Ctx, logs []logWithBlockTime) error {
	for _, log := range logs {
		switch log.Topics[0] {
		case transferSig:
			e := toTransferEvent(&log)
			err := h.erc721EventUC.Transfer(ctx, domain.ChainId(h.chainId), e, toLogMeta(&log))
			if err != nil {
				ctx.WithField("err", err).Error("erc721EventUC.Transfer failed")
				return err
			}
		default:
			ctx.WithField("topic", log.Topics[0]).Warn("unknown topic, skipping")
		}
	}
	return nil
}

func toTransferEvent(log *logWithBlockTime) *contract.TransferEvent {
	transferLog := abi.ToTransferLog(&log.Log)
	return &contract.TransferEvent{
		From:    toDomainAddress(transferLog.From),
		To:      toDomainAddress(transferLog.To),
		TokenId: domain.TokenId(transferLog.TokenId.String()),
	}
}
