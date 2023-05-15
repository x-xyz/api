package repository

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/service/query"
)

type nsRepoImpl struct {
	query query.Mongo
}

// NewNotificationSettingsRepo creates notification settings repository
func NewNotificationSettingsRepo(query query.Mongo) account.NotificationSettingsRepo {
	return &nsRepoImpl{
		query: query,
	}
}

func (im *nsRepoImpl) Get(c ctx.Ctx, address domain.Address) (*account.NotificationSettings, error) {
	// If not found, return default values (i.e. all values are true)
	id := address.ToLower()
	settings := &account.NotificationSettings{
		Address:                id,
		FNotification:          true,
		FBundleCreation:        true,
		FBundleList:            true,
		FBundlePrice:           true,
		FNftAuctionPrice:       true,
		FNftList:               true,
		FNftAuction:            true,
		FNftPrice:              true,
		SNotification:          true,
		SBundleBuy:             true,
		SBundleSell:            true,
		SBundleOffer:           true,
		SBundleOfferCancel:     true,
		SNftAuctionPrice:       true,
		SNftBidToAuction:       true,
		SNftBidToAuctionCancel: true,
		SAuctionWin:            true,
		SAuctionOfBidCancel:    true,
		SNftSell:               true,
		SNftBuy:                true,
		SNftOffer:              true,
		SNftOfferCancel:        true,
	}
	if err := im.query.FindOne(c, domain.TableNotificationSettings, bson.M{"address": id}, settings); err != nil && err != query.ErrNotFound {
		c.WithFields(log.Fields{
			"address": address,
			"err":     err,
		}).Error("find notification settings failed")
		return nil, err
	}
	return settings, nil
}

func (im *nsRepoImpl) Upsert(c ctx.Ctx, settings *account.NotificationSettings) (*account.NotificationSettings, error) {
	settings.Address = settings.Address.ToLower()
	if err := im.query.Upsert(c, domain.TableNotificationSettings, bson.M{"address": settings.Address}, settings); err != nil {
		c.WithFields(log.Fields{
			"address": settings.Address,
			"err":     err,
		}).Error("upsert notification settings failed")
		return nil, err
	}
	return settings, nil
}
