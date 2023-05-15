package tracker

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/punk"
)

var transfer = abi.PunkABI.Events["Transfer"].ID
var punkAssign = abi.PunkABI.Events["Assign"].ID
var punkTransfer = abi.PunkABI.Events["PunkTransfer"].ID
var punkBought = abi.PunkABI.Events["PunkBought"].ID

type PunkEventHandlerCfg struct {
	ChainId          int64
	PunkEventUseCase punk.PunkEventUseCase
	RpcClient        domain.EthClientRepo
}

type PunkEventHandler struct {
	PunkEventHandlerCfg
	Topics [][]common.Hash
}

func NewPunkEventHandler(cfg *PunkEventHandlerCfg) EventHandler {
	return &PunkEventHandler{
		PunkEventHandlerCfg: *cfg,
		Topics: [][]common.Hash{
			{punkAssign, punkTransfer, punkBought},
		},
	}
}

func (h *PunkEventHandler) GetFilterTopics() [][]common.Hash {
	return h.Topics
}

func (h *PunkEventHandler) ProcessEvents(ctx bCtx.Ctx, logs []logWithBlockTime) error {
	for _, log := range logs {
		switch log.Topics[0] {
		case punkAssign:
			transfer, err := punkAssignToTransfer(&log)
			if err != nil {
				ctx.WithField("err", err).Error("toPunkAssign failed")
				return err
			}
			err = h.PunkEventUseCase.Transfer(ctx, domain.ChainId(h.ChainId), transfer, toLogMeta(&log))
			if err != nil {
				ctx.WithField("err", err).Error("PunkEventUseCase.Transfer failed")
				return err
			}
		case punkTransfer:
			transfer, err := punkTransferToTransfer(&log)
			if err != nil {
				ctx.WithField("err", err).Error("toPunkAssign failed")
				return err
			}
			err = h.PunkEventUseCase.Transfer(ctx, domain.ChainId(h.ChainId), transfer, toLogMeta(&log))
			if err != nil {
				ctx.WithField("err", err).Error("PunkEventUseCase.Transfer failed")
				return err
			}
		case punkBought:
			transfer, err := punkBoughtToTransfer(&log)
			if err != nil {
				ctx.WithField("err", err).Error("toPunkAssign failed")
				return err
			}

			// special case
			// PunkBought will be emitted in buyPunk and acceptBidForPunk,
			// but the "to" field will always be 0 due to an implementation error in punk's contract.
			// In this special case, we use TransactionReceiptByHash to get the receipt of that transaction
			// and find the Transfer event emitted just before PunkBought event in acceptBidForPunk
			// and use it to get the "to" address
			if transfer.To.Equals(domain.EmptyAddress) {
				to, err := h.getPunkBoughtTo(ctx, log.TxHash, log.Index)
				if err != nil {
					ctx.WithField("err", err).Error("getPunkBoughtTo failed")
					return err
				}
				transfer.To = to
			}

			err = h.PunkEventUseCase.Transfer(ctx, domain.ChainId(h.ChainId), transfer, toLogMeta(&log))
			if err != nil {
				ctx.WithField("err", err).Error("PunkEventUseCase.Transfer failed")
				return err
			}
		default:
			ctx.WithField("topic", log.Topics[0]).Warn("unknown topic, skipping")
		}
	}
	return nil
}

func (h PunkEventHandler) getPunkBoughtTo(ctx bCtx.Ctx, txHash common.Hash, logIndex uint) (domain.Address, error) {
	ctx.Info(fmt.Sprintf("bought %s", txHash.String()))
	receipt, err := h.RpcClient.TransactionReceipt(ctx, txHash)
	if err != nil {
		return "", err
	}

	var prev *types.Log
	for _, log := range receipt.Logs {
		if log.Index == logIndex-1 {
			prev = log
		}
	}
	if prev.Topics[0] != transfer {
		ctx.WithField("err", err).Error("expect a Transfer event before PunkBought event")
		return "", err
	}

	to := common.BytesToAddress(prev.Topics[2].Bytes())
	return toDomainAddress(to), nil

}

func punkAssignToTransfer(log *logWithBlockTime) (*punk.Transfer, error) {
	assign, err := abi.ToPunkAssignLog(&log.Log)
	if err != nil {
		return nil, err
	}

	transfer := &punk.Transfer{
		From:    domain.EmptyAddress,
		To:      toDomainAddress(assign.To),
		TokenId: domain.TokenId(assign.PunkIndex.String()),
	}
	return transfer, nil
}

func punkTransferToTransfer(log *logWithBlockTime) (*punk.Transfer, error) {
	punkTransfer, err := abi.ToPunkTransferLog(&log.Log)
	if err != nil {
		return nil, err
	}

	transfer := &punk.Transfer{
		From:    toDomainAddress(punkTransfer.From),
		To:      toDomainAddress(punkTransfer.To),
		TokenId: domain.TokenId(punkTransfer.PunkIndex.String()),
	}
	return transfer, nil
}

func punkBoughtToTransfer(log *logWithBlockTime) (*punk.Transfer, error) {
	bought, err := abi.ToPunkBoughtLog(&log.Log)
	if err != nil {
		return nil, err
	}

	transfer := &punk.Transfer{
		From:    toDomainAddress(bought.From),
		To:      toDomainAddress(bought.To),
		TokenId: domain.TokenId(bought.PunkIndex.String()),
	}
	return transfer, nil
}
