package usecase

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/ethereum"
	"github.com/x-xyz/goapi/base/log"
	pricefomatter "github.com/x-xyz/goapi/base/price_fomatter"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/erc1155"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/domain/token"
	"github.com/x-xyz/goapi/service/chain/contract"
)

type OrderUseCaseCfg struct {
	ExchangeCfgs        map[domain.ChainId]order.ExchangeCfg
	OrderRepo           order.OrderRepo
	OrderItemRepo       order.OrderItemRepo
	NftitemRepo         nftitem.Repo
	Erc1155HoldingRepo  erc1155.HoldingRepo
	AccountRepo         account.Repo
	PaytokenRepo        domain.PayTokenRepo
	PriceFormatter      pricefomatter.PriceFormatter
	OrderNonceUC        account.OrderNonceUseCase
	TokenUC             token.Usecase
	Erc1271             contract.Erc1271Contract
	ActivityHistoryRepo account.ActivityHistoryRepo
}

type impl struct {
	exchangeCfgs        map[domain.ChainId]order.ExchangeCfg
	orderRepo           order.OrderRepo
	orderItemRepo       order.OrderItemRepo
	nftitemRepo         nftitem.Repo
	erc1155HoldingRepo  erc1155.HoldingRepo
	accountRepo         account.Repo
	paytokenRepo        domain.PayTokenRepo
	priceFormatter      pricefomatter.PriceFormatter
	orderNonceUC        account.OrderNonceUseCase
	tokenUC             token.Usecase
	erc1271             contract.Erc1271Contract
	activityHistoryRepo account.ActivityHistoryRepo
}

func New(cfg *OrderUseCaseCfg) order.UseCase {
	return &impl{
		exchangeCfgs:        cfg.ExchangeCfgs,
		orderRepo:           cfg.OrderRepo,
		orderItemRepo:       cfg.OrderItemRepo,
		nftitemRepo:         cfg.NftitemRepo,
		erc1155HoldingRepo:  cfg.Erc1155HoldingRepo,
		accountRepo:         cfg.AccountRepo,
		paytokenRepo:        cfg.PaytokenRepo,
		priceFormatter:      cfg.PriceFormatter,
		orderNonceUC:        cfg.OrderNonceUC,
		tokenUC:             cfg.TokenUC,
		erc1271:             cfg.Erc1271,
		activityHistoryRepo: cfg.ActivityHistoryRepo,
	}
}

func (im *impl) FindAll(ctx ctx.Ctx, opts ...order.OrderItemFindAllOptionsFunc) ([]*order.OrderItem, error) {
	res, err := im.orderItemRepo.FindAll(ctx, opts...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to order.FindAll")
		return nil, err
	}

	return res, nil
}

func (im *impl) GetOrder(ctx ctx.Ctx, id order.OrderId) (*order.Order, error) {
	order, err := im.orderRepo.FindOne(ctx, id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to orderRepo.FindOneOrder")
		return nil, err
	}

	return order, nil
}

func (im *impl) MakeOrder(ctx ctx.Ctx, od order.Order) error {
	od.LowerCase()
	err := im.validateOrder(ctx, od)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to validateOrder")
		return err
	}
	orderHash, err := od.Hash()
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to od.Hash")
		return err
	}
	od.OrderHash = domain.OrderHash(hexutil.Encode(orderHash))

	if err := im.processOrderItems(ctx, &od); err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"order": od,
		}).Error("processOrderItems failed")
		im.removeRelatedOrders(ctx, od)
		return err
	}

	if err := im.orderRepo.Upsert(ctx, &od); err != nil {
		ctx.WithFields(log.Fields{
			"err":    err,
			"signer": od.Signer,
			"nonce":  od.Nonce,
		}).Error("orderRepo.Upsert failed")
		im.removeRelatedOrders(ctx, od)
		return err
	}
	orderNonceId := account.OrderNonceId{Address: od.Signer, ChainId: od.ChainId}
	if err := im.orderNonceUC.UpdateAvailableNonceIfNeeded(ctx, orderNonceId, od.Nonce); err != nil {
		ctx.WithFields(log.Fields{
			"err":          err,
			"orderNonceId": orderNonceId,
			"nonce":        od.Nonce,
		}).Error("orderNonceUC.UpdateAvailableNonceIfNeeded failed")
		im.removeRelatedOrders(ctx, od)
		return err
	}
	return nil
}

func (im *impl) removeRelatedOrders(ctx ctx.Ctx, od order.Order) {
	if err := im.orderItemRepo.RemoveAll(ctx, order.WithChainId(od.ChainId), order.WithOrderHash(od.OrderHash)); err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"chainId":   od.ChainId,
			"orderHash": od.OrderHash,
		}).Error("orderItemRepo.RemoveAll failed")
	}
	if err := im.orderRepo.RemoveAll(ctx, order.OrderWithChainId(od.ChainId), order.OrderWithOrderHash(od.OrderHash)); err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"chainId":   od.ChainId,
			"orderHash": od.OrderHash,
		}).Error("orderRepo.RemoveAll failed")
	}
}

func (im *impl) processOrderItems(ctx ctx.Ctx, od *order.Order) error {
	startTime, err := strconv.ParseInt(od.StartTime, 10, 64)
	if err != nil {
		return err
	}
	endTime, err := strconv.ParseInt(od.EndTime, 10, 64)
	if err != nil {
		return err
	}
	for idx, item := range od.Items {
		priceInt, ok := new(big.Int).SetString(item.Price, 10)
		if !ok {
			ctx.WithFields(log.Fields{
				"err":   err,
				"price": item.Price,
			}).Error("big.Int.SetString failed")
			return domain.ErrInvalidNumberFormat
		}
		displayPrice, priceInUsd, priceInNative, err := im.priceFormatter.GetPrices(ctx, od.ChainId, od.Currency, priceInt)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err":      err,
				"currency": od.Currency,
				"price":    item.Price,
			}).Error("failed to priceFormatter.GetPrices")
			return err
		}
		orderItemHash, err := od.HashOrderItem(idx)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err":       err,
				"orderHash": od.OrderHash,
			}).Error("HashOrderItem failed")
			return err
		}
		strategy := im.exchangeCfgs[od.ChainId].Strategies[od.Strategy]
		reservedBuyer := domain.Address("")
		nonce, ok := new(big.Int).SetString(od.Nonce, 10)
		if !ok {
			return domain.ErrInvalidNumberFormat
		}
		switch strategy {
		case order.StrategyPrivateSale:
			reservedBuyer = domain.Address(common.HexToAddress(od.Params).Hex()).ToLower()
		case order.StrategyCollectionOffer:
		case order.StrategyFixedPrice:
		default:
		}

		orderItem := &order.OrderItem{
			ChainId:            od.ChainId,
			Item:               item,
			ItemIdx:            idx,
			OrderHash:          od.OrderHash,
			OrderItemHash:      domain.OrderHash(hexutil.Encode(orderItemHash)),
			IsAsk:              od.IsAsk,
			Signer:             od.Signer,
			Nonce:              od.Nonce,
			HexNonce:           hexutil.Encode(math.U256Bytes(nonce)),
			Currency:           od.Currency,
			StartTime:          time.Unix(startTime, 0),
			EndTime:            time.Unix(endTime, 0),
			MinPercentageToAsk: od.MinPercentageToAsk,
			Marketplace:        od.Marketplace,
			Strategy:           strategy,
			ReservedBuyer:      reservedBuyer,
			PriceInUsd:         priceInUsd,
			PriceInNative:      priceInNative,
			DisplayPrice:       displayPrice.String(),
			IsValid:            true,
			IsUsed:             false,
		}
		err = im.orderItemRepo.Upsert(ctx, orderItem)
		if err != nil {
			// TODO: if error occured, delete all the other orderItems
			ctx.WithFields(log.Fields{
				"orderItem": orderItem,
				"err":       err,
			}).Error("failed to orderRepo.Upsert")
			return err
		}

		nftitemId := nftitem.Id{
			ChainId:         od.ChainId,
			ContractAddress: item.Collection.ToLower(),
			TokenId:         item.TokenId,
		}
		if strategy != order.StrategyCollectionOffer {
			err = im.tokenUC.RefreshListingAndOfferState(ctx, nftitemId)
			if err != nil {
				ctx.WithFields(log.Fields{
					"err": err,
					"id":  nftitemId,
				}).Error("failed to tokenUC.RefreshListingAndOfferState")
				return err
			}
		}

		now := time.Now()
		if od.IsAsk {
			patchable := &nftitem.PatchableNftItem{
				ListedAt: &now,
			}
			if err := im.tokenUC.PatchNft(ctx, &nftitemId, patchable); err != nil {
				ctx.WithFields(log.Fields{
					"err":       err,
					"id":        nftitemId,
					"patchable": patchable,
				}).Error("tokenUC.PatchNft failed")
				return err
			}
		}

		activityType := account.ActivityHistoryTypeList
		if !od.IsAsk {
			activityType = account.ActivityHistoryTypeCreateOffer
		}
		activityHistory := &account.ActivityHistory{
			ChainId:         od.ChainId,
			ContractAddress: item.Collection,
			TokenId:         item.TokenId,
			Type:            activityType,
			Account:         od.Signer,
			Quantity:        item.Amount,
			Price:           displayPrice.String(),
			PaymentToken:    od.Currency,
			PriceInUsd:      priceInUsd,
			PriceInNative:   priceInNative,
			Time:            now,
			Source:          account.SourceX,
		}
		if err := im.activityHistoryRepo.Insert(ctx, activityHistory); err != nil {
			ctx.WithFields(log.Fields{
				"err":      err,
				"activity": activityHistory,
			}).Error("activityHistoryRepo.Insert failed")
			return err
		}
	}
	return nil
}

func (im *impl) CancelOrderItemByOrderHash(ctx ctx.Ctx, chainId domain.ChainId, orderHash domain.OrderHash) error {
	orderItems, err := im.orderItemRepo.FindAll(ctx, order.WithChainId(chainId), order.WithOrderHash(orderHash))
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to orderRepo.RemoveAll")
		return err
	}

	for _, oi := range orderItems {
		err := im.orderItemRepo.Update(ctx, oi.ToId(), order.OrderItemPatchable{
			IsUsed: ptr.Bool(true),
		})
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
			}).Error("failed to orderRepo.Update")
			return err
		}
	}
	return nil
}

func (im *impl) CancelOrderItemByOrderItemHash(ctx ctx.Ctx, chainId domain.ChainId, orderItemHash domain.OrderHash, logCancelActivity bool, lMeta *domain.LogMeta) error {
	orderItems, err := im.orderItemRepo.FindAll(ctx, order.WithChainId(chainId), order.WithOrderItemHash(orderItemHash), order.WithIsUsed(false))
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to orderItemRepo.FindAll")
		return err
	}

	for _, oi := range orderItems {
		err := im.orderItemRepo.Update(ctx, oi.ToId(), order.OrderItemPatchable{
			IsUsed: ptr.Bool(true),
		})
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
			}).Error("failed to orderItemRepo.Update")
			return err
		}
		if !logCancelActivity {
			continue
		}
		typ := account.ActivityHistoryTypeCancelOffer
		if oi.IsAsk {
			typ = account.ActivityHistoryTypeCancelListing
		}

		activityHistory := &account.ActivityHistory{
			ChainId:         oi.ChainId,
			ContractAddress: oi.Collection,
			TokenId:         oi.TokenId,
			Type:            typ,
			Account:         oi.Signer,
			Quantity:        oi.Amount,
			Price:           oi.DisplayPrice,
			PaymentToken:    oi.Currency,
			PriceInUsd:      oi.PriceInUsd,
			PriceInNative:   oi.PriceInNative,
			BlockNumber:     lMeta.BlockNumber,
			TxHash:          lMeta.TxHash,
			LogIndex:        int64(lMeta.LogIndex),
			Time:            lMeta.BlockTime,
			Source:          account.SourceX,
		}
		if err := im.activityHistoryRepo.Insert(ctx, activityHistory); err != nil {
			ctx.WithFields(log.Fields{
				"err":             err,
				"activityHistory": activityHistory,
			}).Error(" failed")
			return err
		}
	}
	return nil
}

func (im *impl) CancelOrderItemByNonce(ctx ctx.Ctx, chainId domain.ChainId, signer domain.Address, nonce *big.Int, lMeta *domain.LogMeta) error {
	orderItems, err := im.orderItemRepo.FindAll(ctx,
		order.WithChainId(chainId),
		order.WithSigner(signer),
		order.WithNonceLT(nonce.String()),
	)
	if err != nil {
		return err
	}

	for _, item := range orderItems {
		if err := im.CancelOrderItemByOrderItemHash(ctx, chainId, item.OrderItemHash, true, lMeta); err != nil {
			return err
		}

	}
	return nil
}

func (im *impl) validateOrder(ctx ctx.Ctx, makerOrder order.Order) error {
	orderNonceId := account.OrderNonceId{Address: makerOrder.Signer, ChainId: makerOrder.ChainId}
	orderNonce, err := im.orderNonceUC.FindOne(ctx, orderNonceId)
	if err == domain.ErrNotFound {
		// unused account, valid nonce
	} else if err != nil {
		ctx.WithFields(log.Fields{
			"err":          err,
			"orderNonceId": orderNonceId,
		}).Error("orderNonceUC.FindOne failed")
		return err
	} else {
		minValidOrderNonce := orderNonce.MinValidOrderNonce
		nums, err := domain.ToBigInt([]string{minValidOrderNonce, makerOrder.Nonce})
		if err != nil {
			return err
		}
		if nums[0].Cmp(nums[1]) > 0 { // minValidOrderNonce > order.nonce
			return domain.ErrInvalidOrderNonce
		}
	}

	// validate strategy
	strategy := im.exchangeCfgs[makerOrder.ChainId].Strategies[makerOrder.Strategy]
	switch strategy {
	case order.StrategyPrivateSale:
		if !makerOrder.IsAsk {
			return domain.ErrInvalidOrderSideForStrategy
		}
	case order.StrategyCollectionOffer:
		if makerOrder.IsAsk {
			return domain.ErrInvalidOrderSideForStrategy
		}
	case order.StrategyFixedPrice:
	}

	cnt, err := im.orderRepo.Count(ctx,
		order.OrderWithChainId(makerOrder.ChainId),
		order.OrderWithSigner(makerOrder.Signer),
		order.OrderWithNonce(makerOrder.Nonce),
	)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to orderRepo.Count")
		return err
	}
	if cnt != 0 {
		return domain.ErrInvalidOrderNonce
	}

	exchangeCfg, ok := im.exchangeCfgs[makerOrder.ChainId]
	if !ok {
		return domain.ErrInvalidChainId
	}

	if _, ok := exchangeCfg.Strategies[makerOrder.Strategy]; !ok {
		return domain.ErrInvalidStrategy
	}

	if err := verifyOrderSignature(ctx, makerOrder, exchangeCfg.Address, im.erc1271); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to verifyOrderSignature")
		return err
	}

	if _, err := im.paytokenRepo.FindOne(ctx, makerOrder.ChainId, makerOrder.Currency); err != nil {
		return domain.ErrInvalidCurrency
	}

	return nil
}

func verifyOrderSignature(ctx ctx.Ctx, makerOrder order.Order, verifyingContract domain.Address, erc1271 contract.Erc1271Contract) error {
	typedData := apitypes.TypedData{
		Types:       order.OrderTypes,
		PrimaryType: order.PrimaryType,
		Domain:      order.GetDomainSeperator(makerOrder.ChainId, verifyingContract),
		Message:     makerOrder.ToMessage(),
	}

	domainSeperator, err := typedData.HashStruct(order.Eip712DomainName, typedData.Domain.Map())
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to HashStruct")
		return err
	}
	dataHash, err := makerOrder.Hash()
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to makerOrder.Hash")
		return err
	}
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeperator), string(dataHash)))
	hash := crypto.Keccak256(rawData)
	sig := []byte{}
	sig = append(sig, common.FromHex(makerOrder.R)...)
	sig = append(sig, common.FromHex(makerOrder.S)...)
	sig = append(sig, big.NewInt(int64(makerOrder.V)).Bytes()...)

	valid, err := ethereum.ValidateHashSignature(hash, hexutil.Encode(sig), makerOrder.Signer.ToLowerStr())
	if err == nil && valid {
		return nil
	}
	ctx.WithFields(log.Fields{
		"hash":  hash,
		"sig":   sig,
		"err":   err,
		"valid": valid,
	}).Warn("validating eoa signature failed")

	valid, err = erc1271.IsValidSignature(ctx, int32(makerOrder.ChainId), makerOrder.Signer.ToLowerStr(), common.BytesToHash(hash), sig)
	if err == nil && valid {
		return nil
	}
	ctx.WithFields(log.Fields{
		"hash":  hash,
		"sig":   sig,
		"err":   err,
		"valid": valid,
	}).Warn("validating eip1271 signature failed")

	return domain.ErrInvalidSignature
}

func (im *impl) RefreshOrders(ctx ctx.Ctx, nftitemId nftitem.Id) error {
	now := time.Now()

	token, err := im.nftitemRepo.FindOne(ctx, nftitemId.ChainId, nftitemId.ContractAddress, nftitemId.TokenId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  nftitemId,
		}).Error("failed to nftitemRepo.FindOne")
		return err
	}

	holdingMap := map[domain.Address]int{}
	if token.TokenType == 721 {
		holdingMap[token.Owner] = 1
	} else if token.TokenType == 1155 {
		holdings, err := im.erc1155HoldingRepo.FindAll(ctx, erc1155.WithNftitemId(nftitemId))
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
				"id":  nftitemId,
			}).Error("failed to erc1155HoldingRepo.FindAll")
			return err
		}

		for _, h := range holdings {
			holdingMap[h.Owner] = int(h.Balance)
		}
	}

	orders, err := im.orderItemRepo.FindAll(ctx,
		order.WithNftItemId(nftitemId),
		order.WithIsUsed(false),
		order.WithEndTimeGT(now),
		order.WithStartTimeLT(now),
	)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to orderRepo.FindAll")
		return err
	}

	for _, od := range orders {
		priceInt, ok := new(big.Int).SetString(od.Price, 10)
		if !ok {
			ctx.WithFields(log.Fields{
				"err":   err,
				"price": od.Price,
			}).Error("big.Int.SetString failed")
			return domain.ErrInvalidNumberFormat
		}
		displayPrice, priceInUsd, priceInNative, err := im.priceFormatter.GetPrices(ctx, od.ChainId, od.Currency, priceInt)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err":      err,
				"currency": od.Currency,
				"price":    od.Price,
			}).Error("failed to priceFormatter.GetPrices")
			return err
		}
		amount, err := strconv.ParseInt(od.Amount, 10, 64)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err":    err,
				"amount": od.Amount,
			}).Error("failed to strconv.ParseInt")
			return err
		}

		valid := false
		if !od.IsAsk || holdingMap[od.Signer] >= int(amount) {
			valid = true
		}

		err = im.orderItemRepo.Update(ctx, od.ToId(), order.OrderItemPatchable{
			IsValid:       &valid,
			DisplayPrice:  ptr.String(displayPrice.String()),
			PriceInUsd:    &priceInUsd,
			PriceInNative: &priceInNative,
		})
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
				"id":  od.ToId(),
			}).Error("failed to orderRepo.Update")
			return err
		}
	}

	return nil
}
