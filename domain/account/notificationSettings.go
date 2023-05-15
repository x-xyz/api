package account

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

// NotificationSettings is notification settings for user's activities or user's follower activties
type NotificationSettings struct {
	Address domain.Address `json:"address" bson:"address"`

	FNotification    bool `json:"fNotification" bson:"fNotification"`
	FBundleCreation  bool `json:"fBundleCreation" bson:"fBundleCreation"`
	FBundleList      bool `json:"fBundleList" bson:"fBundleList"`
	FBundlePrice     bool `json:"fBundlePrice" bson:"fBundlePrice"`
	FNftAuctionPrice bool `json:"fNftAuctionPrice" bson:"fNftAuctionPrice"`
	FNftList         bool `json:"fNftList" bson:"fNftList"`
	FNftAuction      bool `json:"fNftAuction" bson:"fNftAuction"`
	FNftPrice        bool `json:"fNftPrice" bson:"fNftPrice"`

	SNotification          bool `json:"sNotification" bson:"sNotification"`
	SBundleBuy             bool `json:"sBundleBuy" bson:"sBundleBuy"`
	SBundleSell            bool `json:"sBundleSell" bson:"sBundleSell"`
	SBundleOffer           bool `json:"sBundleOffer" bson:"sBundleOffer"`
	SBundleOfferCancel     bool `json:"sBundleOfferCancel" bson:"sBundleOfferCancel"`
	SNftAuctionPrice       bool `json:"sNftAuctionPrice" bson:"sNftAuctionPrice"`
	SNftBidToAuction       bool `json:"sNftBidToAuction" bson:"sNftBidToAuction"`
	SNftBidToAuctionCancel bool `json:"sNftBidToAuctionCancel" bson:"sNftBidToAuctionCancel"`
	SAuctionWin            bool `json:"sAuctionWin" bson:"sAuctionWin"`
	SAuctionOfBidCancel    bool `json:"sAuctionOfBidCancel" bson:"sAuctionOfBidCancel"`
	SNftSell               bool `json:"sNftSell" bson:"sNftSell"`
	SNftBuy                bool `json:"sNftBuy" bson:"sNftBuy"`
	SNftOffer              bool `json:"sNftOffer" bson:"sNftOffer"`
	SNftOfferCancel        bool `json:"sNftOfferCancel" bson:"sNftOfferCancel"`
}

// NotificationSettingsRepo notification settings repo
type NotificationSettingsRepo interface {
	Get(c ctx.Ctx, address domain.Address) (*NotificationSettings, error)
	Upsert(c ctx.Ctx, settings *NotificationSettings) (*NotificationSettings, error)
}
