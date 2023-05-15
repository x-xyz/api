package repository

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/erc1155"
	"github.com/x-xyz/goapi/service/query"
)

type holdingSuite struct {
	suite.Suite
	db   *mongoclient.Client
	q    query.Mongo
	impl *holdingImpl
}

func TestHoldingSuite(t *testing.T) {
	suite.Run(t, new(holdingSuite))
}

func (s *holdingSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.q = q
	s.db = mongoClient
	s.impl = NewHoldingRepo(q).(*holdingImpl)
}

func (s *holdingSuite) SetupTest() {
	s.db.Database("test").Drop(ctx.Background())
}

func (s *holdingSuite) TestFindAll() {
	ctx := ctx.Background()
	mockOwner := domain.Address("0x5566")
	data := []erc1155.Holding{
		{
			ChainId: 1,
			Address: "0xabc123",
			TokenId: "1",
			Owner:   mockOwner,
			Balance: 123,
		},
		{
			ChainId: 1,
			Address: "0xabc123",
			TokenId: "2",
			Owner:   mockOwner,
			Balance: 345,
		},
		{
			ChainId: 1,
			Address: "0xabc123",
			TokenId: "2",
			Owner:   domain.Address("0x1234"),
			Balance: 345,
		},
	}

	expected := []*erc1155.Holding{
		{
			ChainId: 1,
			Address: "0xabc123",
			TokenId: "1",
			Owner:   mockOwner,
			Balance: 123,
		},
		{
			ChainId: 1,
			Address: "0xabc123",
			TokenId: "2",
			Owner:   mockOwner,
			Balance: 345,
		},
	}

	for _, d := range data {
		err := s.q.Insert(ctx, domain.TableERC1155Holdings, &d)
		s.Nil(err)
	}

	output, err := s.impl.FindAll(ctx, erc1155.WithOwner(mockOwner))
	s.Nil(err)
	s.Equal(expected, output)
}
