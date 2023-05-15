package tracker

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"

	"github.com/x-xyz/goapi/base/abi"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/manifold"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/service/chain"
	"github.com/x-xyz/goapi/service/chain/contract"
)

var royaltyOverrideSig = abi.ManifoldABI.Events["RoyaltyOverride"].ID

type ManifoldEventHandlerCfg struct {
	ChainId        domain.ChainId
	CollectionRepo collection.Repo
	NftItemRepo    nftitem.Repo
	ChainService   chain.Client
	RoyaltyEngine  common.Address
}

type ManifoldEventHandler struct {
	ManifoldEventHandlerCfg
	royaltyEngine contract.RoyaltyEngineContract
	topics        [][]common.Hash
}

func NewManifoldEventHandler(cfg *ManifoldEventHandlerCfg) *ManifoldEventHandler {
	royaltyEngine := contract.NewRoyaltyEngine(cfg.ChainService)
	return &ManifoldEventHandler{
		ManifoldEventHandlerCfg: *cfg,
		royaltyEngine:           royaltyEngine,
		topics: [][]common.Hash{
			{royaltyOverrideSig},
		},
	}
}

func (m *ManifoldEventHandler) GetFilterTopics() [][]common.Hash {
	return m.topics
}

func (m *ManifoldEventHandler) ProcessEvents(c ctx.Ctx, logs []logWithBlockTime) error {
	for _, log := range logs {
		switch log.Topics[0] {
		case royaltyOverrideSig:
			if evt, err := toRoyaltyOverrideEvent(&log); err != nil {
				c.WithField("err", err).Error("failed to parse RoyaltyOverride log")
				return err
			} else if err := m.processRoyaltyOverride(c, evt); err != nil {
				c.WithField("err", err).Error("failed to handle RoyaltyOverride event")
				return err
			}
		}
	}

	return nil
}

func (m *ManifoldEventHandler) processRoyaltyOverride(c ctx.Ctx, evt *manifold.RoyaltyOverrideEvent) error {
	coll, err := m.CollectionRepo.FindOne(c, collection.CollectionId{ChainId: m.ChainId, Address: evt.TokenAddress})
	if err != nil {
		if err == domain.ErrNotFound {
			c.WithField("err", err).Warn("Item not found")
			return nil
		}
		c.WithField("err", err).Error("FindOne failed")
		return err
	}

	if coll == nil {
		return nil
	}

	var contractAddresses []domain.Address
	contractAddresses = append(contractAddresses, evt.TokenAddress)
	nftItemOpts := []nftitem.FindAllOptionsFunc{
		nftitem.WithContractAddresses(contractAddresses),
		nftitem.WithChainId(m.ChainId),
		nftitem.WithPagination(0, 1),
	}
	nftItems, err := m.NftItemRepo.FindAll(c, nftItemOpts...)
	if err != nil {
		c.WithField("err", err).Error("FindAll failed")
		return err
	}
	if len(nftItems) == 0 {
		return nil
	}
	nftItem := nftItems[0]
	baseNum := decimal.New(1, 5) // assessed value
	bi := new(big.Int)
	tokenId, ok := bi.SetString(string(nftItem.TokenId), 10)
	if !ok {
		return errors.New("SetString failed")
	}
	recipients, originalRoyalties, err := m.royaltyEngine.GetRoyalty(c, int32(m.ChainId), m.RoyaltyEngine.String(), evt.RoyaltyAddress.ToLowerStr(), tokenId, baseNum.BigInt())
	if err != nil {
		c.WithField("err", err).Error("GetRoyalty failed")
		return err
	}
	if len(recipients) == 0 || len(originalRoyalties) == 0 {
		return nil
	}
	r := decimal.NewFromBigInt(originalRoyalties[0], 0)
	royaltyValue := r.Div(baseNum).Mul(decimal.NewFromInt(100))
	royalty, err := strconv.ParseFloat(royaltyValue.String(), 64)
	if err != nil {
		c.WithField("err", err).Error("ParseFloat failed")
		return err
	}
	if coll.Royalty == royalty {
		return nil
	}

	updatedPayload := collection.UpdatePayload{
		FeeRecipient: recipients[0],
		Royalty:      royalty,
	}
	err = m.CollectionRepo.Update(c, coll.ToId(), updatedPayload)
	if err != nil {
		c.WithField("err", err).Error("Update failed")
		return err
	}

	return nil
}

func toRoyaltyOverrideEvent(log *logWithBlockTime) (*manifold.RoyaltyOverrideEvent, error) {
	l, err := abi.ToRoyaltyOverrideLog(&log.Log)
	if err != nil {
		return nil, err
	}

	return &manifold.RoyaltyOverrideEvent{
		Owner:          toDomainAddress(l.Owner),
		TokenAddress:   toDomainAddress(l.TokenAddress),
		RoyaltyAddress: toDomainAddress(l.RoyaltyAddress),
	}, nil
}
