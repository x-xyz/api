package usecase

import (
	"fmt"
	"math/big"
	"time"

	"github.com/shopspring/decimal"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/airdrop"
	"github.com/x-xyz/goapi/domain/collection_promotion"
	"github.com/x-xyz/goapi/domain/promotion"
	"golang.org/x/xerrors"
)

type CollPromotionCfg struct {
	CollPromotionRepo collection_promotion.CollPromotionRepo
	PromotionUsecase  promotion.PromotionUsecase
	ListingRecordRepo airdrop.ListingRecordRepo
	OpenseaDataRepo   domain.OpenseaDataRepo
}

type collPromotionImpl struct {
	collPromotionRepo collection_promotion.CollPromotionRepo
	promotionUsecase  promotion.PromotionUsecase
	listingRecordRepo airdrop.ListingRecordRepo
	openseaData       domain.OpenseaDataRepo
}

func NewCollPromotion(cfg *CollPromotionCfg) collection_promotion.CollPromotionUsecase {
	return &collPromotionImpl{
		collPromotionRepo: cfg.CollPromotionRepo,
		promotionUsecase:  cfg.PromotionUsecase,
		listingRecordRepo: cfg.ListingRecordRepo,
		openseaData:       cfg.OpenseaDataRepo,
	}
}

func (im *collPromotionImpl) CreateCollPromotion(c ctx.Ctx, collPromos []collection_promotion.CollPromotion, promotionId *string) error {
	if err := im.collPromotionRepo.Create(c, collPromos, promotionId); err != nil {
		c.WithField("err", err).Error("collPromotion.Create failed")
		return err
	}
	return nil
}

func (im *collPromotionImpl) GetCollPromotions(c ctx.Ctx, promotionIds *[]string) ([]*collection_promotion.CollPromotion, error) {
	collections, err := im.collPromotionRepo.FindAll(c, collection_promotion.WithPromotionIds(promotionIds))
	if err != nil {
		c.WithField("err", err).Error("collPromotion.FindAll failed")
		return nil, err
	}
	return collections, nil
}

func (im *collPromotionImpl) GetPromotedCollections(c ctx.Ctx, ts *time.Time) (*promotion.Promotion, []*collection_promotion.CollPromotion, error) {
	promotions, err := im.promotionUsecase.GetActivatedPromotions(c, ts)
	if err != nil {
		c.WithFields(log.Fields{
			"ts":  ts,
			"err": err,
		}).Error("promotionUsecase.GetActivatedPromotions failed")
		return nil, nil, err
	}
	if len(promotions) == 0 {
		return nil, nil, domain.ErrNotFound
	}
	if len(promotions) != 1 {
		return nil, nil, xerrors.Errorf("Multiple active promotions not supported")
	}
	activePromotion := promotions[0]
	actPromotionIds := []string{activePromotion.Id}
	chainId := domain.ChainId(1)
	collections, err := im.collPromotionRepo.FindAll(c, collection_promotion.WithPromotionIds(&actPromotionIds), collection_promotion.WithChainId(&chainId))
	if err != nil {
		c.WithField("err", err).Error("collPromotion.FindAll failed")
		return nil, nil, err
	}
	return activePromotion, collections, nil
}

func (im *collPromotionImpl) CalculateListingRewardsFlat(c ctx.Ctx, begin time.Time, end time.Time) (*collection_promotion.ListingRewardDistribution, error) {
	_, collections, err := im.GetPromotedCollections(c, &begin)
	if err != nil {
		c.WithFields(log.Fields{
			"begin": begin,
			"err":   err,
		}).Error("GetPromotedCollections failed")
		return nil, err
	}

	accountListings := make(map[domain.Address]map[domain.Address]int)
	collectionListings := make(map[domain.Address]int)
	accountRewards := make(map[domain.Address]*big.Int)

	for _, collection := range collections {
		rewardPerListing, ok := new(big.Int).SetString(collection.Reward, 10)
		if !ok {
			// expected to be set
			return nil, xerrors.Errorf("big.Int.SetString(%s) failed", collection.Reward)
		}
		subRewardPerListing, ok := new(big.Int).SetString(collection.SubReward, 10)
		if !ok {
			// optional
			subRewardPerListing = big.NewInt(0)
		}
		lrs, err := im.listingRecordRepo.FindAll(c,
			airdrop.ListingRecordWithChainId(collection.ChainId),
			airdrop.ListingRecordWithContractAddress(collection.Address),
			airdrop.ListingRecordWithSnapshotTime(begin, end),
		)
		if err != nil {
			c.WithFields(log.Fields{
				"collection": collection,
				"begin":      begin,
				"end":        end,
			}).Error("listingRecordRepo.FindAll failed")
			return nil, err
		}
		for _, lr := range lrs {
			if accountListings[lr.Owner] == nil {
				accountListings[lr.Owner] = make(map[domain.Address]int)
			}
			if accountRewards[lr.Owner] == nil {
				accountRewards[lr.Owner] = big.NewInt(0)
			}
			if lr.Count > 0 {
				accountListings[lr.Owner][collection.Address] += lr.Count
				reward := new(big.Int).Mul(rewardPerListing, big.NewInt(int64(lr.Count)))
				accountRewards[lr.Owner] = accountRewards[lr.Owner].Add(accountRewards[lr.Owner], reward)
				collectionListings[collection.Address] += lr.Count
			}
			if lr.SubCount > 0 {
				collectionSubAddress := domain.Address(fmt.Sprintf("%s-sub", collection.Address))
				accountListings[lr.Owner][collectionSubAddress] += lr.SubCount
				reward := new(big.Int).Mul(subRewardPerListing, big.NewInt(int64(lr.SubCount)))
				accountRewards[lr.Owner] = accountRewards[lr.Owner].Add(accountRewards[lr.Owner], reward)
				collectionListings[collectionSubAddress] += lr.SubCount
			}
		}
	}

	rewards := make(map[domain.Address]string)
	for a, r := range accountRewards {
		rewards[a] = r.String()
	}

	distribution := &collection_promotion.ListingRewardDistribution{
		Rewards:            rewards,
		CollectionListings: collectionListings,
		AccountListings:    accountListings,
	}

	return distribution, nil
}

func (im *collPromotionImpl) CalculateLastHourAverageRewardPerListing(c ctx.Ctx) (string, error) {
	lastHour := time.Now().Truncate(time.Hour)
	activePromo, collections, err := im.GetPromotedCollections(c, &lastHour)
	if err == domain.ErrNotFound {
		return "0", nil
	} else if err != nil {
		return "", err
	}

	rewardPerHour, ok := new(big.Int).SetString(activePromo.RewardPerDistribution, 10)
	if !ok {
		return "", domain.ErrInvalidNumberFormat
	}
	rewardPerHour = rewardPerHour.Div(rewardPerHour, big.NewInt(24))

	totalListings := 0
	for _, collection := range collections {
		lrs, err := im.listingRecordRepo.FindAll(c,
			airdrop.ListingRecordWithChainId(collection.ChainId),
			airdrop.ListingRecordWithContractAddress(collection.Address),
			airdrop.ListingRecordWithSnapshotTime(lastHour, lastHour.Add(time.Hour)),
		)
		if err != nil {
			return "", err
		}
		for _, lr := range lrs {
			totalListings += lr.Count
		}
	}
	if totalListings == 0 {
		return "0", nil
	}
	return rewardPerHour.Div(rewardPerHour, big.NewInt(int64(totalListings))).String(), nil
}

func (im *collPromotionImpl) CalculateListingRewardsFixedTotal(c ctx.Ctx, begin time.Time, end time.Time) (*collection_promotion.ListingRewardDistribution, error) {
	if end.Sub(begin) != 24*time.Hour {
		return nil, xerrors.Errorf("Supporting only 24hr-distribution")
	}
	activePromo, collections, err := im.GetPromotedCollections(c, &begin)
	if err != nil {
		c.WithFields(log.Fields{
			"begin": begin,
			"err":   err,
		}).Error("GetPromotedCollections failed")
		return nil, err
	}

	rewardPerHour, ok := new(big.Int).SetString(activePromo.RewardPerDistribution, 10)
	if !ok {
		return nil, domain.ErrInvalidNumberFormat
	}
	rewardPerHour = rewardPerHour.Div(rewardPerHour, big.NewInt(24))
	accountListings := make(map[domain.Address]map[domain.Address]int)
	collectionListings := make(map[domain.Address]int)
	accountRewards := make(map[domain.Address]*big.Int)

	for t := begin; t.Before(end); t = t.Add(time.Hour) {
		_end := t.Add(time.Hour)
		totalMultipliers := int64(0)
		accountMultiplier := make(map[domain.Address]int64)
		for _, collection := range collections {
			lrs, err := im.listingRecordRepo.FindAll(c,
				airdrop.ListingRecordWithChainId(collection.ChainId),
				airdrop.ListingRecordWithContractAddress(collection.Address),
				airdrop.ListingRecordWithSnapshotTime(t, _end),
			)
			if err != nil {
				c.WithFields(log.Fields{
					"collection": collection,
					"begin":      t,
					"end":        _end,
				}).Error("listingRecordRepo.FindAll failed")
				return nil, err
			}
			for _, lr := range lrs {
				if lr.Count > 0 {
					if accountListings[lr.Owner] == nil {
						accountListings[lr.Owner] = make(map[domain.Address]int)
					}
					accountListings[lr.Owner][collection.Address] += lr.Count
					collectionListings[collection.Address] += lr.Count
					multiplier := int64(lr.Count * collection.Multiplier)
					totalMultipliers += multiplier
					accountMultiplier[lr.Owner] += multiplier
				}
			}
		}

		for account, multiplier := range accountMultiplier {
			if accountRewards[account] == nil {
				accountRewards[account] = big.NewInt(0)
			}
			reward := new(big.Int).Mul(rewardPerHour, big.NewInt(multiplier))
			reward = reward.Div(reward, big.NewInt(totalMultipliers))
			accountRewards[account] = accountRewards[account].Add(accountRewards[account], reward)
		}
	}

	rewards := make(map[domain.Address]string)
	for a, r := range accountRewards {
		rewards[a] = r.String()
	}

	distribution := &collection_promotion.ListingRewardDistribution{
		Rewards:            rewards,
		CollectionListings: collectionListings,
		AccountListings:    accountListings,
	}

	return distribution, nil
}

func (im *collPromotionImpl) CreateWeeklyPromotion(c ctx.Ctx, name string, startTime *time.Time, endTime *time.Time, topK int32, reward decimal.Decimal) ([]collection_promotion.CollPromotion, error) {
	topCollections, err := im.openseaData.FindAll(c,
		domain.OpenseaDataWithChainId(1),
		domain.OpenseaDataWithPagination(0, topK),
		domain.OpenseaDataWithSort("sevenDayVolume", domain.SortDirDesc),
	)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("openseaData.FindAll failed")
		return nil, err
	}
	promotion := &promotion.Promotion{
		Name:      name,
		StartTime: startTime,
		EndTime:   endTime,
	}
	promotion, err = im.promotionUsecase.CreatePromotion(c, promotion)
	if err != nil {
		c.WithFields(log.Fields{
			"promotion": promotion,
			"err":       err,
		}).Error("promotionUC.CreatePromotion failed")
		return nil, err
	}

	var collections []collection_promotion.CollPromotion
	for _, collection := range topCollections {
		collections = append(collections, collection_promotion.CollPromotion{
			ChainId: 1,
			Address: collection.Address,
			Reward:  reward.String(),
		})
	}
	if err := im.CreateCollPromotion(c, collections, &promotion.Id); err != nil {
		c.WithFields(log.Fields{
			"err":         err,
			"collections": collections,
			"promotionId": promotion.Id,
		}).Error(" failed")
		return nil, err
	}
	return collections, nil
}
