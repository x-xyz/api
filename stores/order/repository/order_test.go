package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type orderSuite struct {
	suite.Suite

	query query.Mongo
	im    *orderRepoImpl
}

func (s *orderSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q

	s.im = NewOrderRepo(q).(*orderRepoImpl)
}

func TestOrderSuite(t *testing.T) {
	suite.Run(t, new(orderSuite))
}

func (s *orderSuite) TestFindAll() {
	timestamp2HAgo := fmt.Sprint(time.Now().Add(-2 * time.Hour).Unix())
	timestamp1HAgo := fmt.Sprint(time.Now().Add(-1 * time.Hour).Unix())
	// timestampNow := fmt.Sprint(time.Now().Unix())
	timestamp1HLater := fmt.Sprint(time.Now().Add(1 * time.Hour).Unix())
	timestamp2HLater := fmt.Sprint(time.Now().Add(2 * time.Hour).Unix())
	cases := []struct {
		name    string
		options []order.OrderFindAllOptionsFunc
		data    []order.Order
		want    []*order.Order
	}{
		{
			name: "test find all with chainId",
			options: []order.OrderFindAllOptionsFunc{
				order.OrderWithChainId(1),
			},
			data: []order.Order{
				{
					ChainId:   1,
					OrderHash: "orderhash1",
					IsAsk:     true,
					Signer:    "signer1",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
				{
					ChainId:   2,
					OrderHash: "orderhash2",
					IsAsk:     true,
					Signer:    "signer2",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
			},
			want: []*order.Order{
				{
					ChainId:   1,
					OrderHash: "orderhash1",
					IsAsk:     true,
					Signer:    "signer1",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
			},
		},
		{
			name: "test find all with orderHash",
			options: []order.OrderFindAllOptionsFunc{
				order.OrderWithOrderHash("orderhash1"),
			},
			data: []order.Order{
				{
					ChainId:   1,
					OrderHash: "orderhash1",
					IsAsk:     true,
					Signer:    "signer1",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
				{
					ChainId:   2,
					OrderHash: "orderhash2",
					IsAsk:     true,
					Signer:    "signer2",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
			},
			want: []*order.Order{
				{
					ChainId:   1,
					OrderHash: "orderhash1",
					IsAsk:     true,
					Signer:    "signer1",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
			},
		},
		{
			name: "test find all with isAsk",
			options: []order.OrderFindAllOptionsFunc{
				order.OrderWithIsAsk(true),
			},
			data: []order.Order{
				{
					ChainId:   1,
					OrderHash: "orderhash1",
					IsAsk:     true,
					Signer:    "signer1",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
				{
					ChainId:   2,
					OrderHash: "orderhash2",
					IsAsk:     false,
					Signer:    "signer2",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
			},
			want: []*order.Order{
				{
					ChainId:   1,
					OrderHash: "orderhash1",
					IsAsk:     true,
					Signer:    "signer1",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
			},
		},
		{
			name: "test find all with time",
			options: []order.OrderFindAllOptionsFunc{
				order.OrderWithStartTimeLT(time.Now()),
				order.OrderWithEndTimeGT(time.Now()),
			},
			data: []order.Order{
				{
					ChainId:   1,
					OrderHash: "orderhash1",
					IsAsk:     true,
					Signer:    "signer1",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
				{
					ChainId:   2,
					OrderHash: "orderhash2",
					IsAsk:     true,
					Signer:    "signer2",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp1HAgo,
				},
				{
					ChainId:   2,
					OrderHash: "orderhash3",
					IsAsk:     true,
					Signer:    "signer3",
					Nonce:     "1",
					StartTime: timestamp1HLater,
					EndTime:   timestamp2HLater,
				},
			},
			want: []*order.Order{
				{
					ChainId:   1,
					OrderHash: "orderhash1",
					IsAsk:     true,
					Signer:    "signer1",
					Nonce:     "1",
					StartTime: timestamp2HAgo,
					EndTime:   timestamp2HLater,
				},
			},
		},
	}

	for _, c := range cases {
		_, err := s.query.RemoveAll(ctx.Background(), domain.TableOrders, bson.M{})
		s.Nil(err)
		for _, d := range c.data {
			err := s.query.Insert(ctx.Background(), domain.TableOrders, d)
			s.Nil(err)
		}

		res, err := s.im.FindAll(ctx.Background(), c.options...)
		s.Nil(err)
		s.Equal(c.want, res, c.name+" failed")
	}
}
