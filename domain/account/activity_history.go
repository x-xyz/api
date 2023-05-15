package account

import (
	"errors"
	"strconv"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/opensea"
)

type ActivityHistoryType string

var ErrNotFoundActivityType = errors.New("not found activity type")

const (
	// marketplace
	ActivityHistoryTypeList          ActivityHistoryType = "list"
	ActivityHistoryTypeUpdateListing ActivityHistoryType = "updateListing"
	ActivityHistoryTypeCancelListing ActivityHistoryType = "cancelListing"
	ActivityHistoryTypeBuy           ActivityHistoryType = "buy"
	ActivityHistoryTypeSold          ActivityHistoryType = "sold"
	ActivityHistoryTypeCreateOffer   ActivityHistoryType = "createOffer"
	ActivityHistoryTypeAcceptOffer   ActivityHistoryType = "acceptOffer"
	ActivityHistoryTypeOfferTaken    ActivityHistoryType = "offerTaken"
	ActivityHistoryTypeCancelOffer   ActivityHistoryType = "cancelOffer"

	// auction
	ActivityHistoryTypeCreateAuction             ActivityHistoryType = "createAuction"
	ActivityHistoryTypePlaceBid                  ActivityHistoryType = "placeBid"
	ActivityHistoryTypeWithdrawBid               ActivityHistoryType = "withdrawBid"
	ActivityHistoryTypeBidRefunded               ActivityHistoryType = "bidRefunded"
	ActivityHistoryTypeResultAuction             ActivityHistoryType = "resultAuction"
	ActivityHistoryTypeWonAuction                ActivityHistoryType = "wonAuction"
	ActivityHistoryTypeCancelAuction             ActivityHistoryType = "cancelAuction"
	ActivityHistoryTypeUpdateAuctionReservePrice ActivityHistoryType = "updateAuctionReservePrice"
	ActivityHistoryTypeUpdateAuctionStartTime    ActivityHistoryType = "updateAuctionStartTime"
	ActivityHistoryTypeUpdateAuctionEndTime      ActivityHistoryType = "updateAuctionEndTime"

	// exchange
	ActivityHistoryTypeSale ActivityHistoryType = "sale"

	// erc721, erc1155 transfer/mint
	ActivityHistoryTypeTransfer ActivityHistoryType = "transfer"
	ActivityHistoryTypeMint     ActivityHistoryType = "mint"
)

type SourceType string

const (
	SourceX       SourceType = "x"
	SourceOpensea SourceType = "opensea"
)

type ActivityHistory struct {
	ChainId         domain.ChainId      `json:"chainId" bson:"chainId"`
	ContractAddress domain.Address      `json:"contractAddress" bson:"contractAddress"`
	TokenId         domain.TokenId      `json:"tokenId" bson:"tokenId"`
	Type            ActivityHistoryType `json:"type" bson:"type"`
	Account         domain.Address      `json:"account" bson:"account"`
	To              domain.Address      `json:"to" bson:"to"`
	Quantity        string              `json:"quantity" bson:"quantity"`
	Price           string              `json:"price" bson:"price"`
	PaymentToken    domain.Address      `json:"paymentToken" bson:"paymentToken"`
	PriceInUsd      float64             `json:"priceInUsd" bson:"priceInUsd"`
	PriceInNative   float64             `json:"priceInNative" bson:"priceInNative"`
	BlockNumber     domain.BlockNumber  `json:"blockNumber" bson:"blockNumber"`
	TxHash          domain.TxHash       `json:"txHash" bson:"txHash"`
	LogIndex        int64               `json:"logIndex" bson:"logIndex"`
	Time            time.Time           `json:"time" bson:"time"`
	Source          SourceType          `json:"source" bson:"source"`
	SourceEventId   string              `json:"sourceEventId" bson:"sourceEventId"`
}

var activityHistoryTypeToActivityType = map[ActivityHistoryType]ActivityType{
	ActivityHistoryTypeCreateOffer:   ActivityTypeOffer,
	ActivityHistoryTypeCancelOffer:   ActivityTypeCancelOffer,
	ActivityHistoryTypeList:          ActivityTypeList,
	ActivityHistoryTypeCancelListing: ActivityTypeCancelListing,
	ActivityHistoryTypePlaceBid:      ActivityTypeBid,
	ActivityHistoryTypeBuy:           ActivityTypeBuy,
	ActivityHistoryTypeSold:          ActivityTypeSale,
	ActivityHistoryTypeTransfer:      ActivityTypeTransfer,
	ActivityHistoryTypeMint:          ActivityTypeMint,
	ActivityHistoryTypeSale:          ActivityTypeSale2,
}

// Return of `ToActivity`` did not including type token and owner
func (ah *ActivityHistory) ToActivity() (*Activity, error) {
	if ah.Quantity == "" {
		ah.Quantity = "0"
	}
	qty, err := strconv.Atoi(ah.Quantity)
	if err != nil {
		return nil, err
	}

	if ah.Price == "" {
		ah.Price = "0"
	}
	price, err := strconv.ParseFloat(ah.Price, 64)
	if err != nil {
		return nil, err
	}

	typ, ok := activityHistoryTypeToActivityType[ah.Type]
	if !ok {
		return nil, ErrNotFoundActivityType
	}

	return &Activity{
		Type:         typ,
		Quantity:     int32(qty),
		Price:        price,
		PriceInUsd:   ah.PriceInUsd,
		PaymentToken: ah.PaymentToken,
		CreatedAt:    ah.Time,
	}, nil
}

type TransferID struct {
	ChainId         domain.ChainId      `json:"chainId" bson:"chainId"`
	ContractAddress domain.Address      `json:"contractAddress" bson:"contractAddress"`
	Type            ActivityHistoryType `json:"type" bson:"type"`
	TxHash          domain.TxHash       `json:"txHash" bson:"txHash"`
	LogIndex        int64               `json:"logIndex" bson:"logIndex"`
}

type findActivityHistoryOptions struct {
	Offset   *int
	Limit    *int
	Account  *domain.Address
	ChainId  *domain.ChainId
	Contract *domain.Address
	TokenId  *domain.TokenId
	Types    []ActivityHistoryType
	TimeGTE  *time.Time
	Source   *SourceType
}

type FindActivityHistoryOptions func(*findActivityHistoryOptions) error

func GetFindActivityHistoryOptions(opts ...FindActivityHistoryOptions) (*findActivityHistoryOptions, error) {
	res := &findActivityHistoryOptions{}
	for _, opt := range opts {
		if err := opt(res); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func ActivityHistoryWithPagination(offset, limit int) FindActivityHistoryOptions {
	return func(opts *findActivityHistoryOptions) error {
		opts.Offset = &offset
		opts.Limit = &limit
		return nil
	}
}

func ActivityHistoryWithAccount(account domain.Address) FindActivityHistoryOptions {
	return func(opts *findActivityHistoryOptions) error {
		opts.Account = account.ToLowerPtr()
		return nil
	}
}

func ActivityHistoryWithChainId(chainId domain.ChainId) FindActivityHistoryOptions {
	return func(opts *findActivityHistoryOptions) error {
		opts.ChainId = &chainId
		return nil
	}
}

func ActivityHistoryWithContract(contract domain.Address) FindActivityHistoryOptions {
	return func(opts *findActivityHistoryOptions) error {
		opts.Contract = contract.ToLowerPtr()
		return nil
	}
}

func ActivityHistoryWithCollection(chainId domain.ChainId, contract domain.Address) FindActivityHistoryOptions {
	return func(opts *findActivityHistoryOptions) error {
		opts.ChainId = &chainId
		opts.Contract = contract.ToLowerPtr()
		return nil
	}
}

func ActivityHistoryWithToken(chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) FindActivityHistoryOptions {
	return func(opts *findActivityHistoryOptions) error {
		opts.ChainId = &chainId
		opts.Contract = contract.ToLowerPtr()
		opts.TokenId = &tokenId
		return nil
	}
}

func ActivityHistoryWithTypes(types ...ActivityHistoryType) FindActivityHistoryOptions {
	return func(opts *findActivityHistoryOptions) error {
		opts.Types = types
		return nil
	}
}

func ActivityHistoryWithTimeGTE(time time.Time) FindActivityHistoryOptions {
	return func(opts *findActivityHistoryOptions) error {
		opts.TimeGTE = &time
		return nil
	}
}

func ActivityHistoryWithSource(source SourceType) FindActivityHistoryOptions {
	return func(opts *findActivityHistoryOptions) error {
		opts.Source = &source
		return nil
	}
}

type ActivityHistoryRepo interface {
	Insert(ctx.Ctx, *ActivityHistory) error
	FindActivities(c ctx.Ctx, opts ...FindActivityHistoryOptions) ([]ActivityHistory, error)
	CountActivities(c ctx.Ctx, opts ...FindActivityHistoryOptions) (int, error)
	// UpsertBySourceEventId use source, sourceEventId to upsert to prevent duplication
	//
	// Example:
	//
	//     source: opensea
	//     sourceEventId: <opensea_event_id>
	// 	   type: <ActivityHistoryType>
	UpsertBySourceEventId(ctx ctx.Ctx, source SourceType, sourceEventId string, t ActivityHistoryType, ah *ActivityHistory) error

	InsertTransferActivityIfNotExists(ctx ctx.Ctx, ah *ActivityHistory) error
}

type ActivityHistoryUseCase interface {
	Insert(ctx.Ctx, *ActivityHistory) error
	ParseAndInsertOpenseaEventToActivityHistory(ctx ctx.Ctx, ev opensea.AssetEvent) error
}
