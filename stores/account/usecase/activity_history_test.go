package usecase

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/opensea"
	"github.com/x-xyz/goapi/service/query"
	"github.com/x-xyz/goapi/stores/account/repository"
	"go.mongodb.org/mongo-driver/bson"
)

type activityHistorySuite struct {
	suite.Suite

	query  query.Mongo
	im     *activityUsecase
	client opensea.Client
}

func TestActivityHistorySuite(t *testing.T) {
	suite.Run(t, new(activityHistorySuite))
}

func (s *activityHistorySuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.client = opensea.NewClient(&opensea.ClientCfg{
		HttpClient: http.Client{},
		Timeout:    time.Second * 60,
		Apikey:     "api_key",
	})
	s.query = q
	activityHistoryRepo := repository.NewActivityHistoryRepo(q)
	s.im = NewActivityHistoryUsecase(activityHistoryRepo).(*activityUsecase)
}

func (s *activityHistorySuite) SetupTest() {
	ctx := ctx.Background()
	s.query.RemoveAll(ctx, domain.TableActivityHistories, bson.M{})
}

func (s *activityHistorySuite) TestParseSuccessfulEvent() {
	ctx := ctx.Background()

	baycContract := domain.Address("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	data, err := s.client.GetEvent(
		ctx,
		opensea.WithContractAddress(baycContract),
		opensea.WithEventType(opensea.EventTypeSuccessful),
		opensea.WithBefore(time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC)),
	)
	s.Nil(err)

	err = s.im.ParseAndInsertOpenseaEventToActivityHistory(ctx, data.AssetEvents[0])
	s.Nil(err)

}
