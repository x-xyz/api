package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/erc1155"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type nftitemSuite struct {
	suite.Suite

	query query.Mongo
	im    *nftitemImpl
}

func TestNftitemSuite(t *testing.T) {
	suite.Run(t, new(nftitemSuite))
}

func (s *nftitemSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q
	s.im = NewNftItem(q, nil).(*nftitemImpl)
}

func (s *nftitemSuite) TestFindAll() {
	ctx := ctx.Background()
	mockOwner := "0x501fea3b37837cde179d1c38595ea6d590becf2e"
	pastTime := time.Now().Add(-1 * time.Hour)
	futureTime := time.Now().Add(1 * time.Hour)

	cases := []struct {
		name string
		opts []nftitem.FindAllOptionsFunc
		data []*nftitem.NftItem
		want []*nftitem.NftItem
	}{
		{
			name: "ignore null address",
			opts: []nftitem.FindAllOptionsFunc{},
			data: []*nftitem.NftItem{
				{
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           zeroAddress,
				},
			},
			want: []*nftitem.NftItem{},
		},
		{
			name: "find by owner",
			opts: []nftitem.FindAllOptionsFunc{
				nftitem.WithOwner(domain.Address(mockOwner)),
			},
			data: []*nftitem.NftItem{
				{
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
				},
				{
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "2",
					Owner:           zeroAddress,
				},
			},
			want: []*nftitem.NftItem{
				{
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
				},
			},
		},
		{
			name: "find by owner with holding ids",
			opts: []nftitem.FindAllOptionsFunc{
				nftitem.WithOwner(domain.Address(mockOwner)),
				nftitem.WithHoldingIds([]nftitem.Id{
					{
						ChainId:         1,
						ContractAddress: "0x123",
						TokenId:         "1",
					},
				}),
			},
			data: []*nftitem.NftItem{
				{
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
				},
				{
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "2",
					Owner:           zeroAddress,
				},
				{
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
					Owner:           "",
				},
				{
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Owner:           "",
				},
			},
			want: []*nftitem.NftItem{
				{
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
				},
				{
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
					Owner:           "",
				},
			},
		},
		{
			name: "find without erc1155",
			opts: []nftitem.FindAllOptionsFunc{
				nftitem.WithTokenType(721),
			},
			data: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
				},
				{
					TokenType:       1155,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "2",
					Owner:           zeroAddress,
				},
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Owner:           "",
				},
			},
			want: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
				},
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Owner:           "",
				},
			},
		},
		{
			name: "find with buy now",
			opts: []nftitem.FindAllOptionsFunc{
				nftitem.WithBuyNow(),
			},
			data: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
					ListingEndsAt:   &futureTime,
				},
				{
					TokenType:       1155,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "2",
					ListingEndsAt:   &pastTime,
				},
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Owner:           "",
					ListingEndsAt:   &pastTime,
				},
			},
			want: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
					ListingEndsAt:   &futureTime,
				},
			},
		},
		{
			name: "find with listing from",
			opts: []nftitem.FindAllOptionsFunc{
				nftitem.WithListingFrom(domain.Address(mockOwner)),
			},
			data: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
					ListingOwners:   []domain.Address{domain.Address(mockOwner)},
				},
				{
					TokenType:       1155,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "2",
				},
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Owner:           "",
				},
			},
			want: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
					ListingOwners:   []domain.Address{domain.Address(mockOwner)},
				},
			},
		},
		{
			name: "find with has offer",
			opts: []nftitem.FindAllOptionsFunc{
				nftitem.WithHasOffer(),
			},
			data: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
					OfferEndsAt:     &futureTime,
				},
				{
					TokenType:       1155,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "2",
					OfferEndsAt:     &pastTime,
				},
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Owner:           "",
					OfferEndsAt:     &pastTime,
				},
			},
			want: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					Owner:           domain.Address(mockOwner),
					OfferEndsAt:     &futureTime,
				},
			},
		},
		{
			name: "find with offer owners",
			opts: []nftitem.FindAllOptionsFunc{
				nftitem.WithOfferOwners([]domain.Address{domain.Address(mockOwner)}),
			},
			data: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					OfferOwners:     []domain.Address{domain.Address(mockOwner)},
				},
				{
					TokenType:       1155,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "2",
				},
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Owner:           "",
				},
			},
			want: []*nftitem.NftItem{
				{
					TokenType:       721,
					ChainId:         1,
					ContractAddress: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					TokenId:         "1",
					OfferOwners:     []domain.Address{domain.Address(mockOwner)},
				},
			},
		},
	}

	for _, c := range cases {
		s.query.RemoveAll(ctx, domain.TableNFTItems, bson.M{})
		for _, d := range c.data {
			err := s.query.Insert(ctx, domain.TableNFTItems, d)
			s.Nil(err)
		}

		res, err := s.im.FindAll(ctx, c.opts...)
		s.Nil(err)

		s.Equal(len(c.want), len(res))

		// HACK: currently testify doesn't support comparing to time.Time, wait for testify v2
		for i := 0; i < len(c.want) && i < len(res); i++ {
			if c.want[i].ListingEndsAt != res[i].ListingEndsAt {
				s.WithinDuration(*c.want[i].ListingEndsAt, *res[i].ListingEndsAt, 1*time.Millisecond, fmt.Sprintf("%s: ListingEndAt Not Equal", c.name))
				c.want[i].ListingEndsAt = nil
				res[i].ListingEndsAt = nil
			}

			if c.want[i].OfferEndsAt != res[i].OfferEndsAt {
				s.WithinDuration(*c.want[i].OfferEndsAt, *res[i].OfferEndsAt, 1*time.Millisecond, fmt.Sprintf("%s: OfferEndsAt Not Equal", c.name))
				c.want[i].OfferEndsAt = nil
				res[i].OfferEndsAt = nil
			}
		}

		// don't care objectId
		for _, item := range res {
			item.ObjectId = primitive.NilObjectID
		}

		s.ElementsMatch(c.want, res, fmt.Sprintf("test case %s failed", c.name))
	}
}

type nftitemTestSuite struct {
	suite.Suite
	dbName      string
	db          *mongoclient.Client
	nftID       nftitem.Id
	nft         nftitem.NftItem
	nftitemRepo nftitem.Repo
	holdingRepo erc1155.HoldingRepo
}

func TestNFTItem(t *testing.T) {
	suite.Run(t, new(nftitemTestSuite))
}
