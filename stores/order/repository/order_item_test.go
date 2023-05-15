package repository

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type testSuite struct {
	suite.Suite

	query query.Mongo
	im    *impl
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q
	s.im = NewOrderItemRepo(q).(*impl)
}

func (s *testSuite) TestFindAll() {
	now := time.Now()
	time1HAgo := time.Now().Add(-1 * time.Hour)
	time2HAgo := time.Now().Add(-2 * time.Hour)
	time1HAfter := time.Now().Add(1 * time.Hour)
	time2HAfter := time.Now().Add(2 * time.Hour)

	ctx := ctx.Background()
	cases := []struct {
		name string
		opts []order.OrderItemFindAllOptionsFunc
		data []order.OrderItem
		want []*order.OrderItem
	}{
		{
			name: "find by orderHash",
			opts: []order.OrderItemFindAllOptionsFunc{
				order.WithOrderHash("123"),
			},
			data: []order.OrderItem{

				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
				{
					OrderHash:     "123",
					OrderItemHash: "1231",
					ItemIdx:       1,
					Signer:        "0xabc",
				},
				{
					OrderHash:     "456",
					OrderItemHash: "4560",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
			},
			want: []*order.OrderItem{
				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
				{
					OrderHash:     "123",
					OrderItemHash: "1231",
					ItemIdx:       1,
					Signer:        "0xabc",
				},
			},
		},
		{
			name: "find by nftitem.Id",
			opts: []order.OrderItemFindAllOptionsFunc{
				order.WithNftItemId(nftitem.Id{
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
				}),
			},
			data: []order.OrderItem{
				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
					ChainId:       1,
					Item: order.Item{
						Collection: "0x123",
						TokenId:    "1",
					},
				},
				{
					OrderHash:     "456",
					OrderItemHash: "4561",
					ItemIdx:       1,
					Signer:        "0xabc",
					ChainId:       1,
					Item: order.Item{
						Collection: "0x123",
						TokenId:    "2",
					},
				},
				{
					OrderHash:     "789",
					OrderItemHash: "7890",
					ItemIdx:       0,
					Signer:        "0xabc",
					ChainId:       1,
					Item: order.Item{
						Collection: "0x123",
						TokenId:    "1",
					},
				},
			},
			want: []*order.OrderItem{
				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
					ChainId:       1,
					Item: order.Item{
						Collection: "0x123",
						TokenId:    "1",
					},
				},
				{
					OrderHash:     "789",
					OrderItemHash: "7890",
					ItemIdx:       0,
					Signer:        "0xabc",
					ChainId:       1,
					Item: order.Item{
						Collection: "0x123",
						TokenId:    "1",
					},
				},
			},
		},
		{
			name: "find valid order items",
			opts: []order.OrderItemFindAllOptionsFunc{
				order.WithIsValid(true),
				order.WithIsUsed(false),
				order.WithStartTimeLT(now),
				order.WithEndTimeGT(now),
			},
			data: []order.OrderItem{
				{
					ChainId:       1,
					OrderHash:     "1",
					OrderItemHash: "10",
					ItemIdx:       0,
					IsValid:       true,
					IsUsed:        false,
					StartTime:     time1HAgo,
					EndTime:       time1HAfter,
				},
				{
					ChainId:       1,
					OrderHash:     "2",
					OrderItemHash: "20",
					ItemIdx:       0,
					IsValid:       false,
					IsUsed:        false,
					StartTime:     time1HAgo,
					EndTime:       time1HAfter,
				},
				{
					ChainId:       1,
					OrderHash:     "3",
					OrderItemHash: "30",
					ItemIdx:       0,
					IsValid:       false,
					IsUsed:        true,
					StartTime:     time1HAgo,
					EndTime:       time1HAfter,
				},
				{
					ChainId:       1,
					OrderHash:     "4",
					OrderItemHash: "40",
					ItemIdx:       0,
					IsValid:       true,
					IsUsed:        false,
					StartTime:     time1HAfter,
					EndTime:       time2HAfter,
				},
				{
					ChainId:       1,
					OrderHash:     "5",
					OrderItemHash: "50",
					ItemIdx:       0,
					IsValid:       true,
					IsUsed:        false,
					StartTime:     time2HAgo,
					EndTime:       time1HAgo,
				},
			},
			want: []*order.OrderItem{
				{
					ChainId:       1,
					OrderHash:     "1",
					OrderItemHash: "10",
					ItemIdx:       0,
					IsValid:       true,
					IsUsed:        false,
					StartTime:     time1HAgo,
					EndTime:       time1HAfter,
				},
			},
		},
		{
			name: "find all by strategy",
			opts: []order.OrderItemFindAllOptionsFunc{
				order.WithStrategy(order.StrategyCollectionOffer),
			},
			data: []order.OrderItem{
				{
					ChainId:       1,
					OrderHash:     "1",
					OrderItemHash: "10",
					ItemIdx:       0,
					Strategy:      order.StrategyCollectionOffer,
					IsValid:       true,
					IsUsed:        false,
					StartTime:     time1HAgo,
					EndTime:       time1HAfter,
				},
				{
					ChainId:       1,
					OrderHash:     "2",
					OrderItemHash: "20",
					ItemIdx:       0,
					Strategy:      order.StrategyFixedPrice,
					IsValid:       false,
					IsUsed:        false,
					StartTime:     time1HAgo,
					EndTime:       time1HAfter,
				},
				{
					ChainId:       1,
					OrderHash:     "3",
					OrderItemHash: "30",
					ItemIdx:       0,
					Strategy:      order.StrategyPrivateSale,
					IsValid:       false,
					IsUsed:        true,
					StartTime:     time1HAgo,
					EndTime:       time1HAfter,
				},
			},
			want: []*order.OrderItem{
				{
					ChainId:       1,
					OrderHash:     "1",
					OrderItemHash: "10",
					ItemIdx:       0,
					Strategy:      order.StrategyCollectionOffer,
					IsValid:       true,
					IsUsed:        false,
					StartTime:     time1HAgo,
					EndTime:       time1HAfter,
				},
			},
		},
		{
			name: "find by order nonce",
			opts: []order.OrderItemFindAllOptionsFunc{
				order.WithNonceLT("123"),
			},
			data: []order.OrderItem{
				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
					HexNonce:      hexutil.Encode(math.U256Bytes(big.NewInt(10))),
				},
				{
					OrderHash:     "123",
					OrderItemHash: "1231",
					ItemIdx:       1,
					Signer:        "0xabc",
					HexNonce:      hexutil.Encode(math.U256Bytes(big.NewInt(122))),
				},
				{
					OrderHash:     "123",
					OrderItemHash: "1231",
					ItemIdx:       1,
					Signer:        "0xabc",
					HexNonce:      hexutil.Encode(math.U256Bytes(big.NewInt(123))),
				},
				{
					OrderHash:     "456",
					OrderItemHash: "4560",
					ItemIdx:       0,
					Signer:        "0xabc",
					HexNonce:      hexutil.Encode(math.U256Bytes(big.NewInt(12400000000))),
				},
			},
			want: []*order.OrderItem{
				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
					HexNonce:      hexutil.Encode(math.U256Bytes(big.NewInt(10))),
				},
				{
					OrderHash:     "123",
					OrderItemHash: "1231",
					ItemIdx:       1,
					Signer:        "0xabc",
					HexNonce:      hexutil.Encode(math.U256Bytes(big.NewInt(122))),
				},
			},
		},
	}

	for _, c := range cases {
		_, err := s.query.RemoveAll(ctx, domain.TableOrderItems, bson.M{})
		s.Nil(err)

		for _, item := range c.data {
			err := s.query.Insert(ctx, domain.TableOrderItems, &item)
			s.Nil(err)
		}

		res, err := s.im.FindAll(ctx, c.opts...)
		s.Nil(err)

		s.Equal(len(c.want), len(res))
		for i := 0; i < len(c.want) && i < len(res); i++ {
			s.WithinDuration(c.want[i].StartTime, res[i].StartTime, 1*time.Millisecond)
			s.WithinDuration(c.want[i].EndTime, res[i].EndTime, 1*time.Millisecond)
			c.want[i].StartTime = now
			c.want[i].EndTime = now
			res[i].StartTime = now
			res[i].EndTime = now
		}

		s.Equal(c.want, res, c.name+" failed")
	}
}

func (s *testSuite) TestFindOne() {
	ctx := ctx.Background()
	cases := []struct {
		name string
		id   order.OrderItemId
		data []order.OrderItem
		want *order.OrderItem
	}{
		{
			name: "findOne",
			id: order.OrderItemId{
				ChainId:   1,
				OrderHash: "123",
				ItemIdx:   0,
			},
			data: []order.OrderItem{
				{
					ChainId:       1,
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
				{
					ChainId:       1,
					OrderHash:     "123",
					OrderItemHash: "1231",
					ItemIdx:       1,
					Signer:        "0xabc",
				},
				{
					ChainId:       1,
					OrderHash:     "456",
					OrderItemHash: "4560",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
			},
			want: &order.OrderItem{
				ChainId:       1,
				OrderHash:     "123",
				OrderItemHash: "1230",
				ItemIdx:       0,
				Signer:        "0xabc",
			},
		},
	}

	for _, c := range cases {
		_, err := s.query.RemoveAll(ctx, domain.TableOrderItems, bson.M{})
		s.Nil(err)

		for _, item := range c.data {
			err := s.query.Insert(ctx, domain.TableOrderItems, &item)
			s.Nil(err)
		}

		res, err := s.im.FindOne(ctx, c.id)
		s.Nil(err)
		s.Equal(c.want, res)
	}
}

func (s *testSuite) TestUpsert() {
	ctx := ctx.Background()
	cases := []struct {
		name string
		data []order.OrderItem
		want []order.OrderItem
	}{
		{
			name: "upsert",
			data: []order.OrderItem{
				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
				{
					OrderHash:     "123",
					OrderItemHash: "1231",
					ItemIdx:       1,
					Signer:        "0xabc",
				},
			},
			want: []order.OrderItem{
				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
				{
					OrderHash:     "123",
					OrderItemHash: "1231",
					ItemIdx:       1,
					Signer:        "0xabc",
				},
			},
		},
	}

	for _, c := range cases {
		_, err := s.query.RemoveAll(ctx, domain.TableOrderItems, bson.M{})
		s.Nil(err)

		for _, item := range c.data {
			err := s.im.Upsert(ctx, &item)
			s.Nil(err)
		}

		res := []order.OrderItem{}
		err = s.query.Search(ctx, domain.TableOrderItems, 0, 0, "itemIdx", bson.M{}, &res)
		s.Nil(err)
		s.Equal(c.want, res)
	}
}

func (s *testSuite) TestUpdate() {
	ctx := ctx.Background()
	cases := []struct {
		name      string
		data      order.OrderItem
		patchable order.OrderItemPatchable
		want      order.OrderItem
	}{
		{
			name: "update",
			data: order.OrderItem{
				OrderHash:     "123",
				OrderItemHash: "1230",
				ItemIdx:       0,
				Signer:        "0xabc",

				IsValid: true,
			},
			patchable: order.OrderItemPatchable{
				IsValid: ptr.Bool(false),
			},
			want: order.OrderItem{
				OrderHash:     "123",
				OrderItemHash: "1230",
				ItemIdx:       0,
				Signer:        "0xabc",

				IsValid: false,
			},
		},
		{
			name: "update isUsed",
			data: order.OrderItem{
				OrderHash:     "123",
				OrderItemHash: "1230",
				ItemIdx:       0,
				Signer:        "0xabc",

				IsValid: true,
				IsUsed:  false,
			},
			patchable: order.OrderItemPatchable{
				IsUsed: ptr.Bool(true),
			},
			want: order.OrderItem{
				OrderHash:     "123",
				OrderItemHash: "1230",
				ItemIdx:       0,
				Signer:        "0xabc",

				IsValid: true,
				IsUsed:  true,
			},
		},
	}

	for _, c := range cases {
		_, err := s.query.RemoveAll(ctx, domain.TableOrderItems, bson.M{})
		s.Nil(err)

		err = s.query.Insert(ctx, domain.TableOrderItems, &c.data)
		s.Nil(err)

		err = s.im.Update(ctx, c.data.ToId(), c.patchable)
		s.Nil(err)

		output := order.OrderItem{}
		err = s.query.FindOne(ctx, domain.TableOrderItems, bson.M{}, &output)
		s.Nil(err)
		s.Equal(c.want, output)
	}
}

func (s *testSuite) TestRemoveAll() {
	ctx := ctx.Background()
	cases := []struct {
		name  string
		query []order.OrderItemFindAllOptionsFunc
		data  []order.OrderItem
		want  []order.OrderItem
	}{
		{
			name: "remove all by order hash",
			query: []order.OrderItemFindAllOptionsFunc{
				order.WithOrderHash("123"),
			},
			data: []order.OrderItem{
				{
					OrderHash:     "123",
					OrderItemHash: "1230",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
				{
					OrderHash:     "123",
					OrderItemHash: "1231",
					ItemIdx:       1,
					Signer:        "0xabc",
				},
				{
					OrderHash:     "456",
					OrderItemHash: "4560",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
			},
			want: []order.OrderItem{
				{
					OrderHash:     "456",
					OrderItemHash: "4560",
					ItemIdx:       0,
					Signer:        "0xabc",
				},
			},
		},
	}

	for _, c := range cases {
		_, err := s.query.RemoveAll(ctx, domain.TableOrderItems, bson.M{})
		s.Nil(err)

		for _, d := range c.data {
			err = s.query.Insert(ctx, domain.TableOrderItems, &d)
			s.Nil(err)
		}

		err = s.im.RemoveAll(ctx, c.query...)
		s.Nil(err)

		output := []order.OrderItem{}
		err = s.query.Search(ctx, domain.TableOrderItems, 0, 0, "itemIdx", bson.M{}, &output)
		s.Nil(err)
		s.Equal(c.want, output)
	}
}

func (s *testSuite) TestFindOneOrder() {
	ctx := ctx.Background()
	cases := []struct {
		name      string
		orderHash string
		data      []order.Order
		want      *order.Order
	}{
		{
			name:      "findone",
			orderHash: "orderHash1",
			data: []order.Order{
				{
					ChainId:   1,
					OrderHash: "orderHash1",
					Signer:    "0x123",
				},
				{
					ChainId:   1,
					OrderHash: "orderHash2",
					Signer:    "0x123",
				},
			},
			want: &order.Order{
				ChainId:   1,
				OrderHash: "orderHash1",
				Signer:    "0x123",
			},
		},
	}

	for _, c := range cases {
		_, err := s.query.RemoveAll(ctx, domain.TableOrders, bson.M{})
		s.Nil(err)

		for _, d := range c.data {
			err = s.query.Insert(ctx, domain.TableOrders, &d)
			s.Nil(err)
		}

		output, err := s.im.FindOneOrder(ctx, c.orderHash)
		s.Nil(err)
		s.Equal(c.want, output)
	}
}
