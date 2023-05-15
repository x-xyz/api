package tracker

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/chain"
	"github.com/x-xyz/goapi/domain/exchange"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type TradingBotConfig struct {
	ChainId          domain.ChainId
	DiscordBotKey    string
	DiscordChannelId string
	Nftiem           nftitem.Repo
	Paytoken         domain.PayTokenRepo
	Account          account.Usecase
}

type tradingBotHandler struct {
	config  TradingBotConfig
	discord *discordgo.Session
}

type notifyPayload struct {
	ChainName    string
	ChainUrlPart string
	PayToken     *domain.PayToken
	Price        float64
	Collection   domain.Address
	TokenId      domain.TokenId
	ImageUrl     string
}

func NewTradingBotHandler(config TradingBotConfig) EventHandler {
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", config.DiscordBotKey))
	if err != nil {
		panic("failed to connect to discord")
	}

	return &tradingBotHandler{config, discord}
}

func (h *tradingBotHandler) GetFilterTopics() [][]common.Hash {
	return [][]common.Hash{
		{takerAskSig, takerBidSig},
	}
}

func (h *tradingBotHandler) ProcessEvents(c ctx.Ctx, logs []logWithBlockTime) error {
	for _, log := range logs {
		switch log.Topics[0] {
		case takerBidSig:
			if evt, err := toTakerBidEvent(&log); err != nil {
				c.WithField("err", err).Error("failed to parse TakerBid log")
				return err
			} else if err := h.processTakerBid(c, evt); err != nil {
				c.WithField("err", err).Error("failed to handle TakerBid event")
				return err
			}
		case takerAskSig:
			if evt, err := toTakerAskEvent(&log); err != nil {
				c.WithField("err", err).Error("failed to parse TakerAsk log")
				return err
			} else if err := h.processTakerAsk(c, evt); err != nil {
				c.WithField("err", err).Error("failed to handle TakerAsk event")
				return err
			}
		}
	}

	return nil
}

func (h *tradingBotHandler) preNotifyPayload(c ctx.Ctx, currency domain.Address, price *big.Int, collection domain.Address, tokenId domain.TokenId) (notifyPayload, error) {
	chainName, err := chain.GetChainDisplayName(h.config.ChainId)
	if err != nil {
		c.WithField("chainId", h.config.ChainId).Warn("unknown chainId")
		return notifyPayload{}, err
	}

	chainUrlPart, err := chain.GetChainUrlPart(h.config.ChainId)
	if err != nil {
		c.WithField("chainId", h.config.ChainId).Warn("unknown chainId")
		return notifyPayload{}, err
	}

	paytoken, err := h.config.Paytoken.FindOne(c, h.config.ChainId, currency)
	if err != nil {
		c.WithField("chainId", h.config.ChainId).WithField("payToken", currency).Warn("unknown token")
		return notifyPayload{}, err
	}

	formatedPrice, _ := decimal.NewFromBigInt(price, int32(-paytoken.TokenDecimals)).Float64()

	nftitem, err := h.config.Nftiem.FindOne(c, h.config.ChainId, collection, tokenId)
	if err != nil {
		c.WithField("chainId", h.config.ChainId).WithField("contract", collection).WithField("tokenId", tokenId).Warn("not found nftitem")
		return notifyPayload{}, err
	}

	imageUrl := nftitem.ImageUrl
	if strings.Contains(nftitem.ImageUrl, "ipfs://") {
		imageUrl = strings.Replace(imageUrl, "ipfs://", "https://ipfs.io/ipfs/", 1)
	}

	paylaod := notifyPayload{
		ChainName:    chainName,
		ChainUrlPart: chainUrlPart,
		PayToken:     paytoken,
		Price:        formatedPrice,
		Collection:   collection,
		TokenId:      tokenId,
		ImageUrl:     imageUrl,
	}

	return paylaod, nil
}

func (h *tradingBotHandler) processTakerBid(c ctx.Ctx, evt *exchange.TakerBidEvent) error {
	partialPayload, err := h.preNotifyPayload(c, evt.Fulfillment.Currency, evt.Fulfillment.Price, evt.Fulfillment.Collection, domain.TokenId(evt.Fulfillment.TokenId.String()))
	if err != nil {
		return err
	}
	sellerAlias := "-"
	buyerAlias := "-"

	seller, _ := h.config.Account.Get(c, evt.Maker)
	if seller != nil && len(seller.Alias) > 0 {
		sellerAlias = seller.Alias
	}

	buyer, _ := h.config.Account.Get(c, evt.Taker)
	if buyer != nil && len(buyer.Alias) > 0 {
		buyerAlias = buyer.Alias
	}

	msg := &discordgo.MessageEmbed{
		Title:       "Item sold!",
		Description: fmt.Sprintf("https://x.xyz/asset/%s/%s/%s", partialPayload.ChainUrlPart, evt.Fulfillment.Collection, evt.Fulfillment.TokenId),
		Image: &discordgo.MessageEmbedImage{
			URL: partialPayload.ImageUrl,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Seller", Value: fmt.Sprintf("%s (%s)", evt.Maker, sellerAlias)},
			{Name: "Buyer", Value: fmt.Sprintf("%s (%s)", evt.Taker, buyerAlias)},
			{Name: "Chain", Value: partialPayload.ChainName},
			{Name: "Price", Value: fmt.Sprintf("%s %s", strconv.FormatFloat(partialPayload.Price, 'f', -1, 64), partialPayload.PayToken.Symbol)},
		},
	}

	if _, err := h.discord.ChannelMessageSendEmbed(h.config.DiscordChannelId, msg); err != nil {
		return err
	}
	return nil
}

func (h *tradingBotHandler) processTakerAsk(c ctx.Ctx, evt *exchange.TakerAskEvent) error {
	partialPayload, err := h.preNotifyPayload(c, evt.Fulfillment.Currency, evt.Fulfillment.Price, evt.Fulfillment.Collection, domain.TokenId(evt.Fulfillment.TokenId.String()))
	if err != nil {
		return err
	}
	sellerAlias := "-"
	buyerAlias := "-"

	seller, _ := h.config.Account.Get(c, evt.Taker)
	if seller != nil && len(seller.Alias) > 0 {
		sellerAlias = seller.Alias
	}

	buyer, _ := h.config.Account.Get(c, evt.Maker)
	if buyer != nil && len(buyer.Alias) > 0 {
		buyerAlias = buyer.Alias
	}

	msg := &discordgo.MessageEmbed{
		Title:       "Item sold!",
		Description: fmt.Sprintf("https://x.xyz/asset/%s/%s/%s", partialPayload.ChainUrlPart, evt.Fulfillment.Collection, evt.Fulfillment.TokenId),
		Image: &discordgo.MessageEmbedImage{
			URL: partialPayload.ImageUrl,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Seller", Value: fmt.Sprintf("%s (%s)", evt.Taker, sellerAlias)},
			{Name: "Buyer", Value: fmt.Sprintf("%s (%s)", evt.Maker, buyerAlias)},
			{Name: "Chain", Value: partialPayload.ChainName},
			{Name: "Price", Value: fmt.Sprintf("%s %s", strconv.FormatFloat(partialPayload.Price, 'f', -1, 64), partialPayload.PayToken.Symbol)},
		},
	}

	if _, err := h.discord.ChannelMessageSendEmbed(h.config.DiscordChannelId, msg); err != nil {
		return err
	}

	return nil
}
