package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
	mNftitem "github.com/x-xyz/goapi/domain/nftitem/mocks"
	"github.com/x-xyz/goapi/domain/order"
	mOrder "github.com/x-xyz/goapi/domain/order/mocks"
	"github.com/x-xyz/goapi/service/query"
)

type tokenSuite struct {
	suite.Suite

	query         query.Mongo
	orderItemRepo *mOrder.OrderItemRepo
	nftitemRepo   *mNftitem.Repo
	im            *impl
}

func (s *tokenSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q
	s.orderItemRepo = &mOrder.OrderItemRepo{}
	s.nftitemRepo = &mNftitem.Repo{}
	s.im = New(&TokenUseCaseCfg{
		NftitemRepo:   s.nftitemRepo,
		OrderItemRepo: s.orderItemRepo,
	}).(*impl)
}

func (s *tokenSuite) TearDownTest() {
	s.orderItemRepo.AssertExpectations(s.T())
	s.nftitemRepo.AssertExpectations(s.T())
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(tokenSuite))
}

func (s *tokenSuite) TestGetNftitemIdsIntersection() {
	l1 := []nftitem.Id{
		{
			ChainId:         1,
			ContractAddress: "0x123",
			TokenId:         "1",
		},
		{
			ChainId:         2,
			ContractAddress: "0x456",
			TokenId:         "2",
		},
	}

	l2 := []nftitem.Id{
		{
			ChainId:         2,
			ContractAddress: "0x456",
			TokenId:         "2",
		},
	}

	expected := []nftitem.Id{
		{
			ChainId:         2,
			ContractAddress: "0x456",
			TokenId:         "2",
		},
	}

	s.Equal(expected, getNftitemIdsIntersection(l1, l2))
}

func (s *tokenSuite) TestRefreshListingAndOfferState() {
	time2HAgo := time.Now().Add(-2 * time.Hour)
	time1HAgo := time.Now().Add(-1 * time.Hour)
	time1HAfter := time.Now().Add(1 * time.Hour)
	time2HAfter := time.Now().Add(2 * time.Hour)

	mockNftitem := nftitem.NftItem{
		ChainId:         1,
		ContractAddress: "0x123",
		TokenId:         "1",
	}

	mockOrderItems := []*order.OrderItem{
		{
			IsAsk:   true,
			ChainId: 1,
			Item: order.Item{
				Collection: "0x123",
				TokenId:    "1",
				Amount:     "1",
				Price:      "100",
			},
			StartTime:     time1HAgo,
			EndTime:       time1HAfter,
			PriceInUsd:    10000,
			PriceInNative: 100,
			DisplayPrice:  "100",
			Strategy:      order.StrategyFixedPrice,
			Currency:      domain.EmptyAddress,
			Signer:        "0x5566",
			IsValid:       true,
		},
		{
			IsAsk:   true,
			ChainId: 1,
			Item: order.Item{
				Collection: "0x123",
				TokenId:    "1",
				Amount:     "1",
				Price:      "200",
			},
			StartTime:     time1HAgo,
			EndTime:       time1HAfter,
			PriceInUsd:    20000,
			PriceInNative: 200,
			DisplayPrice:  "200",
			Strategy:      order.StrategyFixedPrice,
			Currency:      domain.EmptyAddress,
			Signer:        "0x5566",
			IsValid:       true,
		},
		{
			IsAsk:   false,
			ChainId: 1,
			Item: order.Item{
				Collection: "0x123",
				TokenId:    "1",
				Amount:     "1",
				Price:      "30",
			},
			StartTime:     time2HAgo,
			EndTime:       time2HAfter,
			PriceInUsd:    3000,
			PriceInNative: 30,
			DisplayPrice:  "30",
			Strategy:      order.StrategyFixedPrice,
			Currency:      domain.EmptyAddress,
			Signer:        "0x1234",
			IsValid:       true,
		},
		{
			IsAsk:   false,
			ChainId: 1,
			Item: order.Item{
				Collection: "0x123",
				TokenId:    "1",
				Amount:     "1",
				Price:      "10",
			},
			StartTime:     time2HAgo,
			EndTime:       time2HAfter,
			PriceInUsd:    1000,
			PriceInNative: 10,
			DisplayPrice:  "10",
			Strategy:      order.StrategyFixedPrice,
			Currency:      domain.EmptyAddress,
			Signer:        "0xabcd",
			IsValid:       true,
		},
	}

	mockCollectionOfferItems := []*order.OrderItem{
		{
			IsAsk:   false,
			ChainId: 1,
			Item: order.Item{
				Collection: "0x123",
				TokenId:    "1",
				Amount:     "1",
				Price:      "50",
			},
			StartTime:     time2HAgo,
			EndTime:       time2HAfter,
			PriceInUsd:    5000,
			PriceInNative: 50,
			DisplayPrice:  "50",
			Strategy:      order.StrategyCollectionOffer,
			Currency:      domain.EmptyAddress,
			Signer:        "0xefg",
		},
	}

	s.orderItemRepo.On("FindAll",
		mock.Anything,
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
	).Return(mockOrderItems, nil).Once()

	s.orderItemRepo.On("FindAll",
		mock.Anything,
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
		mock.AnythingOfType("order.OrderItemFindAllOptionsFunc"),
	).Return(mockCollectionOfferItems, nil).Once()

	s.nftitemRepo.On("Patch",
		mock.Anything,
		*mockNftitem.ToId(),
		mock.MatchedBy(func(input nftitem.PatchableNftItem) bool {
			nftPatchable := nftitem.PatchableNftItem{
				ListingEndsAt:         &time1HAfter,
				ListingOwners:         []domain.Address{"0x5566"},
				InactiveListingOwners: []domain.Address{},
				OfferEndsAt:           &time2HAfter,
				OfferOwners:           []domain.Address{"0x1234", "0xabcd", "0xefg"},
				Price:                 ptr.Float64(100),
				PaymentToken:          domain.EmptyAddress.ToLowerPtr(),
				PriceInUsd:            ptr.Float64(10000),
				PriceSource:           (*nftitem.PriceSource)(ptr.String(nftitem.PriceSourceListing)),
				InstantLiquidityInUsd: ptr.Float64(5000),
				HasActiveListings:     ptr.Bool(true),
				HasOrder:              ptr.Bool(true),
			}
			s.ElementsMatch(nftPatchable.OfferOwners, input.OfferOwners)
			nftPatchable.OfferOwners = []domain.Address{}
			input.OfferOwners = []domain.Address{}
			s.Equal(nftPatchable, input)

			return true
		}),
	).Return(nil).Once()

	err := s.im.RefreshListingAndOfferState(ctx.Background(), *mockNftitem.ToId())
	s.Nil(err)
}
