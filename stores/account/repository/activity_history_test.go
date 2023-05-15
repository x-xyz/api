package repository

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type activityHistorySuite struct {
	suite.Suite

	im    *activityHistoryRepo
	query query.Mongo
}

func (s *activityHistorySuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q
	s.im = NewActivityHistoryRepo(q).(*activityHistoryRepo)
}

func TestAcitivityHistorySuite(t *testing.T) {
	suite.Run(t, new(activityHistorySuite))
}

func (s *activityHistorySuite) SetupTest() {
	s.query.RemoveAll(ctx.Background(), domain.TableActivityHistories, bson.M{})
}

func (s *activityHistorySuite) TestFind() {
	ctx := ctx.Background()
	cases := []struct {
		name  string
		query []account.FindActivityHistoryOptions
		data  []account.ActivityHistory
		want  []account.ActivityHistory
	}{
		{
			name:  "find by source",
			query: []account.FindActivityHistoryOptions{account.ActivityHistoryWithSource(account.SourceX)},
			data: []account.ActivityHistory{
				{
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
					Type:            account.ActivityHistoryTypeBuy,
					Source:          account.SourceX,
				},
				{
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Type:            account.ActivityHistoryTypeBuy,
					Source:          account.SourceOpensea,
				},
			},
			want: []account.ActivityHistory{
				{
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
					Type:            account.ActivityHistoryTypeBuy,
					Source:          account.SourceX,
				},
			},
		},
	}

	for _, c := range cases {
		s.query.RemoveAll(ctx, domain.TableActivityHistories, bson.M{})

		for _, ac := range c.data {
			err := s.query.Insert(ctx, domain.TableActivityHistories, &ac)
			s.Nil(err)
		}

		output, err := s.im.FindActivities(ctx, c.query...)
		s.Nil(err)
		s.ElementsMatch(c.want, output)
	}
}

func (s *activityHistorySuite) TestUpsert() {
	ctx := ctx.Background()

	activity := account.ActivityHistory{
		ChainId:         1,
		ContractAddress: "0xba30e5f9bb24caa003e9f2f0497ad287fdf95623",
		TokenId:         "1",
		Type:            account.ActivityHistoryTypeSold,
		Account:         "0x616413c4a4fee2d64d9f58a56b97684c0e380b37",
		Quantity:        "1",
		Price:           "69",
		PaymentToken:    "0x0000000000000000000000000000000000000000",
		PriceInUsd:      69,
		PriceInNative:   69,
		BlockNumber:     69,
		TxHash:          "0x5e448ed1b47c2fbc5fe0779cdd76cc3050122655a845234cafac3f610116cecd",
		Source:          account.SourceOpensea,
		SourceEventId:   "123",
	}

	err := s.im.UpsertBySourceEventId(ctx, activity.Source, activity.SourceEventId, activity.Type, &activity)
	s.Nil(err)

	result := account.ActivityHistory{}

	err = s.query.FindOne(ctx, domain.TableActivityHistories, bson.M{
		"source":        account.SourceOpensea,
		"sourceEventId": "123",
	}, &result)
	s.Nil(err)
	s.Equal(activity, result)

	err = s.im.UpsertBySourceEventId(ctx, activity.Source, activity.SourceEventId, activity.Type, &activity)
	s.Nil(err)

	n, err := s.query.Count(ctx, domain.TableActivityHistories, bson.M{
		"source":        account.SourceOpensea,
		"sourceEventId": "123",
	})

	s.Nil(err)
	s.Equal(1, n)
}
