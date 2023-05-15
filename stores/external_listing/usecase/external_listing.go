package usecase

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	pricefomatter "github.com/x-xyz/goapi/base/price_fomatter"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/external_listing"
	"github.com/x-xyz/goapi/service/coingecko"
	"github.com/x-xyz/goapi/service/opensea"
)

type impl struct {
	openseaClient       opensea.Client
	externalListingRepo external_listing.ExternalListingRepo
	priceFormatter      pricefomatter.PriceFormatter
	coingecko           coingecko.Client
}

func New(openseaClient opensea.Client, externalListingRepo external_listing.ExternalListingRepo, priceFormatter pricefomatter.PriceFormatter) external_listing.ExternalListingUseCase {
	return &impl{openseaClient: openseaClient, externalListingRepo: externalListingRepo, priceFormatter: priceFormatter}
}

func (im *impl) GetListings(ctx bCtx.Ctx, account domain.Address, chainId domain.ChainId) ([]external_listing.ExternalListing, error) {
	opts := []external_listing.FindAllOptionsFunc{external_listing.WithOwner(account), external_listing.WithChainId(chainId)}
	return im.externalListingRepo.FindAll(ctx, opts...)
}

func (im *impl) FetchOpenseaListings(ctx bCtx.Ctx, account domain.Address, chainId domain.ChainId) ([]external_listing.ExternalListing, error) {
	var getAssetsNext string
	var listingsCol []external_listing.ExternalListing
	ethAddress := domain.Address("0x0000000000000000000000000000000000000000")
	nowTime := time.Now()
	for {
		data, err := im.openseaClient.GetAssetByOwner(ctx, account, getAssetsNext)
		if err != nil {
			ctx.WithField("err", err).Error("GetAssetByOwner failed")
			return nil, err
		}
		for _, asset := range data.Assets {
			if len(asset.SellOrders) > 0 {
				for _, sellOrder := range asset.SellOrders {
					quantity, err := strconv.ParseInt(sellOrder.Quantity, 10, 64)
					if err != nil {
						ctx.WithField("err", err).Error("quantity ParseInt failed")
						return nil, err
					}
					paymentToken := sellOrder.PaymentTokenContract.Address
					if paymentToken != ethAddress {
						continue
					}
					currentPrice, err := decimal.NewFromString(sellOrder.CurrentPrice)
					if err != nil {
						ctx.WithFields(log.Fields{
							"err":          err,
							"currentPrice": sellOrder.CurrentPrice,
						}).Error("currentPrice NewFromString failed")
						return nil, err
					}
					i := decimal.NewFromBigInt(big.NewInt(1), -sellOrder.PaymentTokenContract.Decimals)
					price := currentPrice.Mul(i)
					priceInUsd, priceInNative, err := im.priceFormatter.GetPricesFromDisplayPrice(ctx, chainId, paymentToken, price)
					if err != nil {
						ctx.WithFields(log.Fields{
							"err":          err,
							"currentPrice": sellOrder.CurrentPrice,
						}).Error("priceFormatter GetPricesFromDisplayPrice failed")
						return nil, err
					}
					startTime, err := time.Parse(time.RFC3339, sellOrder.StartTime)
					if err != nil {
						ctx.WithField("err", err).Error("startTime time parse failed")
						return nil, err
					}
					deadline, err := time.Parse("2006-01-02T15:04:05", sellOrder.Deadline)
					if err != nil {
						ctx.WithField("err", err).Error("deadline time parse failed")
						return nil, err
					}
					openseaListing := external_listing.ExternalListing{
						Owner:           account,
						ChainId:         chainId,
						Minter:          sellOrder.Taker.Address,
						ContractAddress: sellOrder.Metadata.Asset.Address,
						TokenId:         sellOrder.Metadata.Asset.TokenId,
						Quantity:        quantity,
						PaymentToken:    paymentToken,
						Price:           fmt.Sprintf("%f", priceInNative),
						PriceInUsd:      fmt.Sprintf("%f", priceInUsd),
						StartTime:       startTime,
						Deadline:        deadline,
						Source:          "opensea",
						UpdatedTime:     nowTime,
					}
					listingsCol = append(listingsCol, openseaListing)
				}
			}

			var nativeTokenType, erc20TokenType = 0, 1
			if len(asset.SeaportSellOrders) > 0 {
				for _, sellOrder := range asset.SeaportSellOrders {
					var skipFlag bool
					for _, consideration := range sellOrder.ProtocolData.Parameters.Consideration {
						if consideration.ContractAddress != ethAddress {
							skipFlag = true
							break
						}
					}
					if len(sellOrder.ProtocolData.Parameters.Offer) > 1 || skipFlag {
						continue
					}
					for _, offer := range sellOrder.ProtocolData.Parameters.Offer {
						if offer.ItemType == nativeTokenType || offer.ItemType == erc20TokenType || offer.ContractAddress.ToLower() != asset.AssetContract.Address.ToLower() {
							continue
						}
						endAmount, err := strconv.ParseInt(offer.EndAmount, 10, 64)
						if err != nil {
							ctx.WithField("err", err).Error("quantity ParseInt failed")
							return nil, err
						}
						currentPrice, err := decimal.NewFromString(sellOrder.CurrentPrice)
						if err != nil {
							ctx.WithFields(log.Fields{
								"err":          err,
								"currentPrice": sellOrder.CurrentPrice,
							}).Error("currentPrice ParseFloat failed")
							return nil, err
						}
						i := decimal.NewFromBigInt(big.NewInt(1), -18)
						price := currentPrice.Mul(i)
						priceInUsd, priceInNative, err := im.priceFormatter.GetPricesFromDisplayPrice(ctx, chainId, ethAddress, price)
						if err != nil {
							ctx.WithFields(log.Fields{
								"err":          err,
								"currentPrice": sellOrder.CurrentPrice,
							}).Error("priceFormatter GetPricesFromDisplayPrice failed")
							return nil, err
						}
						startTime, err := time.Parse("2006-01-02T15:04:05", sellOrder.StartTime)
						if err != nil {
							ctx.WithField("err", err).Error("startTime time parse failed")
							return nil, err
						}
						deadline, err := time.Parse("2006-01-02T15:04:05", sellOrder.Deadline)
						if err != nil {
							ctx.WithField("err", err).Error("deadline time parse failed")
							return nil, err
						}
						openseaListing := external_listing.ExternalListing{
							Owner:           account,
							ChainId:         chainId,
							Minter:          sellOrder.Taker.Address,
							ContractAddress: asset.AssetContract.Address,
							TokenId:         asset.TokenId,
							Quantity:        endAmount,
							PaymentToken:    ethAddress,
							Price:           fmt.Sprintf("%f", priceInNative),
							PriceInUsd:      fmt.Sprintf("%f", priceInUsd),
							StartTime:       startTime,
							Deadline:        deadline,
							Source:          "opensea",
							UpdatedTime:     nowTime,
						}
						listingsCol = append(listingsCol, openseaListing)
					}
				}
			}
		}
		getAssetsNext = data.Next
		if data.Next == "" {
			break
		}
	}
	return listingsCol, nil
}

func (im *impl) BulkUpsert(ctx bCtx.Ctx, o []external_listing.ExternalListing) error {
	return im.externalListingRepo.BulkUpsert(ctx, o)
}

func (im *impl) DeleteListing(ctx bCtx.Ctx, id external_listing.ExternalListingId) error {
	return im.externalListingRepo.RemoveAll(ctx, external_listing.WithExternalListing(id.Owner, id.ChainId))
}
