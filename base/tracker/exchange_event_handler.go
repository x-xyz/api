package tracker

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/exchange"
)

var (
	cancelAllOrdersSig      = abi.ExchangeAPI.Events["CancelAllOrders"].ID
	cancelMultipleOrdersSig = abi.ExchangeAPI.Events["CancelMultipleOrders"].ID
	takerAskSig             = abi.ExchangeAPI.Events["TakerAsk"].ID
	takerBidSig             = abi.ExchangeAPI.Events["TakerBid"].ID
)

type ExchangeEventHandlerCfg struct {
	ChainId         int64
	ExchangeUseCase exchange.UseCase
}

type ExchangeEventHandler struct {
	chainId    int64
	exchangeUC exchange.UseCase
}

func NewExchangeEventHandler(cfg *ExchangeEventHandlerCfg) EventHandler {
	return &ExchangeEventHandler{
		chainId:    cfg.ChainId,
		exchangeUC: cfg.ExchangeUseCase,
	}
}

func (h *ExchangeEventHandler) GetFilterTopics() [][]common.Hash {
	return [][]common.Hash{
		{cancelAllOrdersSig, cancelMultipleOrdersSig, takerAskSig, takerBidSig},
	}
}

func (h *ExchangeEventHandler) ProcessEvents(ctx bCtx.Ctx, logs []logWithBlockTime) error {
	for _, log := range logs {
		switch log.Topics[0] {
		case cancelAllOrdersSig:
			e, err := toCancelAllOrdersEvent(&log)
			if err != nil {
				ctx.WithField("err", err).Error("failed to parse CancelAllOrders log")
				return err
			}
			err = h.exchangeUC.CancelAllOrders(ctx, domain.ChainId(h.chainId), e, toLogMeta(&log))
			if err != nil {
				ctx.WithField("err", err).Error("exchangeUC.CancelAllOrders failed")
				return err
			}
		case cancelMultipleOrdersSig:
			e, err := toCancelMultipleOrdersEvent(&log)
			if err != nil {
				ctx.WithField("err", err).Error("failed to parse CancelMultipleOrders log")
				return err
			}
			err = h.exchangeUC.CancelMultipleOrders(ctx, domain.ChainId(h.chainId), e, toLogMeta(&log))
			if err != nil {
				ctx.WithField("err", err).Error("exchangeUC.CancelMultipleOrders failed")
				return err
			}
		case takerAskSig:
			e, err := toTakerAskEvent(&log)
			if err != nil {
				ctx.WithField("err", err).Error("failed to parse TakerAsk log")
				return err
			}
			err = h.exchangeUC.TakerAsk(ctx, domain.ChainId(h.chainId), e, toLogMeta(&log))
			if err != nil {
				ctx.WithField("err", err).Error("exchangeUC.TakerAsk failed")
				return err
			}
		case takerBidSig:
			e, err := toTakerBidEvent(&log)
			if err != nil {
				ctx.WithField("err", err).Error("failed to parse TakerBid log")
				return err
			}
			err = h.exchangeUC.TakerBid(ctx, domain.ChainId(h.chainId), e, toLogMeta(&log))
			if err != nil {
				ctx.WithField("err", err).Error("exchangeUC.TakerBid failed")
				return err
			}
		default:
			ctx.WithField("signature", log.Topics[0]).Warn("unrecognized signature, skipping")
		}
	}
	return nil
}

func toCancelAllOrdersEvent(log *logWithBlockTime) (*exchange.CancelAllOrdersEvent, error) {
	l, err := abi.ToCancelAllOrdersLog(&log.Log)
	if err != nil {
		return nil, err
	}

	return &exchange.CancelAllOrdersEvent{
		User:     toDomainAddress(l.User),
		NewNonce: l.NewNonce,
	}, nil
}

func toCancelMultipleOrdersEvent(log *logWithBlockTime) (*exchange.CancelMultipleOrdersEvent, error) {
	l, err := abi.ToCancelMultipleOrdersLog(&log.Log)
	if err != nil {
		return nil, err
	}

	var orderItemHashes []domain.OrderHash
	for _, hash := range l.OrderItemHashes {
		orderItemHashes = append(orderItemHashes, toOrderHash(hash))
	}
	return &exchange.CancelMultipleOrdersEvent{
		OrderItemHashes: orderItemHashes,
	}, nil
}

func toTakerAskEvent(log *logWithBlockTime) (*exchange.TakerAskEvent, error) {
	l, err := abi.ToTakerAskLog(&log.Log)
	if err != nil {
		return nil, err
	}

	return &exchange.TakerAskEvent{
		Taker:         toDomainAddress(l.Taker),
		Maker:         toDomainAddress(l.Maker),
		Strategy:      toDomainAddress(l.Strategy),
		OrderHash:     toOrderHash(l.OrderHash),
		ItemIdx:       l.ItemIdx,
		OrderItemHash: toOrderHash(l.OrderItemHash),
		Fulfillment: exchange.Fulfillment{
			Collection: toDomainAddress(l.Fulfillment.Collection),
			TokenId:    l.Fulfillment.TokenId,
			Amount:     l.Fulfillment.Amount,
			Currency:   toDomainAddress(l.Fulfillment.Currency),
			Price:      l.Fulfillment.Price,
		},
		Marketplace: toMarketplaceHash(l.Marketplace),
	}, nil
}

func toTakerBidEvent(log *logWithBlockTime) (*exchange.TakerBidEvent, error) {
	l, err := abi.ToTakerBidLog(&log.Log)
	if err != nil {
		return nil, err
	}

	return &exchange.TakerBidEvent{
		Taker:         toDomainAddress(l.Taker),
		Maker:         toDomainAddress(l.Maker),
		Strategy:      toDomainAddress(l.Strategy),
		OrderHash:     toOrderHash(l.OrderHash),
		ItemIdx:       l.ItemIdx,
		OrderItemHash: toOrderHash(l.OrderItemHash),
		Fulfillment: exchange.Fulfillment{
			Collection: toDomainAddress(l.Fulfillment.Collection),
			TokenId:    l.Fulfillment.TokenId,
			Amount:     l.Fulfillment.Amount,
			Currency:   toDomainAddress(l.Fulfillment.Currency),
			Price:      l.Fulfillment.Price,
		},
		Marketplace: toMarketplaceHash(l.Marketplace),
	}, nil
}
