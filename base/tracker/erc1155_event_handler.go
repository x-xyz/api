package tracker

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/erc1155"
)

// event TransferSingle(address indexed _operator, address indexed _from, address indexed _to, uint256 _id, uint256 _value);
// event TransferBatch(address indexed _operator, address indexed _from, address indexed _to, uint256[] _ids, uint256[] _values);
var transferSingle = abi.ERC1155TokenABI.Events["TransferSingle"].ID
var transferBatch = abi.ERC1155TokenABI.Events["TransferBatch"].ID

type Erc1155EventHandlerCfg struct {
	ChainId             int64
	Erc1155EventUseCase erc1155.Erc1155EventUseCase
}

type Erc1155EventHandler struct {
	Erc1155EventHandlerCfg
	Topics [][]common.Hash
}

func NewErc1155EventHandler(cfg *Erc1155EventHandlerCfg) EventHandler {
	return &Erc1155EventHandler{
		Erc1155EventHandlerCfg: *cfg,
		Topics: [][]common.Hash{
			{transferSingle, transferBatch},
		},
	}
}

func (h *Erc1155EventHandler) GetFilterTopics() [][]common.Hash {
	return h.Topics
}

func (h *Erc1155EventHandler) ProcessEvents(ctx bCtx.Ctx, logs []logWithBlockTime) error {
	for _, log := range logs {
		switch log.Topics[0] {
		case transferSingle:
			ctx.WithField("log", log).Info("process transfer single")
			transfer, err := toTransferSingle(&log)
			if err != nil {
				ctx.WithField("err", err).Error("toTransferSingle failed")
				return err
			}
			err = h.Erc1155EventUseCase.Transfer(ctx, domain.ChainId(h.ChainId), transfer, toLogMeta(&log))
			if err != nil {
				ctx.WithField("err", err).Error("erc721EventUC.Transfer failed")
				return err
			}
		case transferBatch:
			ctx.WithField("log", log).Info("process transfer batch")
			transfers, err := toTransferBatch(&log)
			if err != nil {
				ctx.WithField("err", err).Error("toTransferBatch failed")
				return err
			}
			// run transfer one by one
			for _, transfer := range transfers {
				err = h.Erc1155EventUseCase.Transfer(ctx, domain.ChainId(h.ChainId), transfer, toLogMeta(&log))
				if err != nil {
					ctx.WithField("err", err).Error("erc721EventUC.Transfer failed")
					return err
				}
			}
		default:
			ctx.WithField("topic", log.Topics[0]).Warn("unknown topic, skipping")
		}
	}
	return nil
}

func toTransferSingle(log *logWithBlockTime) (*erc1155.Transfer, error) {
	transferSingle, err := abi.ToErc1155TransferSingleLog(&log.Log)
	transfer := &erc1155.Transfer{
		Operator: toDomainAddress(transferSingle.Operator),
		From:     toDomainAddress(transferSingle.From),
		To:       toDomainAddress(transferSingle.To),
		Id:       domain.TokenId(transferSingle.Id.String()),
		Value:    transferSingle.Value,
	}
	return transfer, err
}

func toTransferBatch(log *logWithBlockTime) ([]*erc1155.Transfer, error) {
	transferBatch, err := abi.ToErc1155TransferBatchLog(&log.Log)
	if err != nil {
		return nil, err
	}
	var transfers []*erc1155.Transfer
	for i := range transferBatch.Ids {
		transfers = append(transfers, &erc1155.Transfer{
			Operator: toDomainAddress(transferBatch.Operator),
			From:     toDomainAddress(transferBatch.From),
			To:       toDomainAddress(transferBatch.To),
			Id:       domain.TokenId(transferBatch.Ids[i].String()),
			Value:    transferBatch.Values[i],
		})
	}
	return transfers, nil
}
