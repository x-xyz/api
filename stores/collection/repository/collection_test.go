package repository

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type collectionSuite struct {
	suite.Suite

	query query.Mongo
	im    *collectionImpl
}

func TestCollectionSuite(t *testing.T) {
	suite.Run(t, new(collectionSuite))
}

func (s *collectionSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q
	s.im = NewCollection(q).(*collectionImpl)
}

func (s *collectionSuite) TestFindAll() {
	ctx := ctx.Background()
	mockOwner := domain.Address("0xc37c41601bc88c91b6569c701f08d37fa0f565f0")
	cases := []struct {
		name         string
		queryOptions []collection.FindAllOptions
		data         []*collection.Collection
		want         []*collection.Collection
	}{
		{
			name:         "find",
			queryOptions: []collection.FindAllOptions{},
			data: []*collection.Collection{
				{
					ChainId:        1,
					Erc721Address:  "0x9a38dec0590abc8c883d72e52391090e948ddf12",
					Owner:          mockOwner,
					CollectionName: "collection1",
				},
			},
			want: []*collection.Collection{
				{
					ChainId:        1,
					Erc721Address:  "0x9a38dec0590abc8c883d72e52391090e948ddf12",
					Owner:          mockOwner,
					CollectionName: "collection1",
				},
			},
		},
		{
			name:         "find editable collections",
			queryOptions: []collection.FindAllOptions{collection.WithAccountEditable(mockOwner)},
			data: []*collection.Collection{
				{
					ChainId:        1,
					Erc721Address:  "0x9a38dec0590abc8c883d72e52391090e948ddf12",
					Owner:          mockOwner,
					CollectionName: "collection1",
				},
				{
					ChainId:          1,
					Erc721Address:    "0xef88c71f5be29c4b30bf89625bd9be8f263e940c",
					EditableAccounts: []domain.Address{mockOwner},
					CollectionName:   "collection2",
				},
				{
					ChainId:        1,
					Erc721Address:  "0xef88c71f5be29c4b30bf89625bd9be8f21234567",
					CollectionName: "collection3",
				},
			},
			want: []*collection.Collection{
				{
					ChainId:        1,
					Erc721Address:  "0x9a38dec0590abc8c883d72e52391090e948ddf12",
					Owner:          mockOwner,
					CollectionName: "collection1",
				},
				{
					ChainId:          1,
					Erc721Address:    "0xef88c71f5be29c4b30bf89625bd9be8f263e940c",
					EditableAccounts: []domain.Address{mockOwner},
					CollectionName:   "collection2",
				},
			},
		},
	}

	for _, c := range cases {
		s.query.RemoveAll(ctx, domain.TableCollections, bson.M{})

		for _, d := range c.data {
			s.query.Insert(ctx, domain.TableCollections, d)
		}

		output, err := s.im.FindAll(ctx, c.queryOptions...)
		s.Nil(err)

		s.ElementsMatch(c.want, output)
	}
}
