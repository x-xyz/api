package usecase

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/service/coingecko"
	"github.com/x-xyz/goapi/service/query"
	account_repository "github.com/x-xyz/goapi/stores/account/repository"
	collection_repository "github.com/x-xyz/goapi/stores/collection/repository"
	erc1155Repository "github.com/x-xyz/goapi/stores/erc1155/repository"
	token_repository "github.com/x-xyz/goapi/stores/token/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type folderUsecaseSuite struct {
	suite.Suite

	query query.Mongo
	im    *folderUsecaseImpl
}

func TestFolderUsecaseSuite(t *testing.T) {
	suite.Run(t, new(folderUsecaseSuite))
}

func (s *folderUsecaseSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "test"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q
	folderRepo := account_repository.NewFolderRepo(q)
	relationRepo := account_repository.NewFolderNftRelationshipRepo(q)
	nftRepo := token_repository.NewNftItem(q, nil)
	collectionRepo := collection_repository.NewCollection(q)
	floorPriceRepo := collection_repository.NewFloorPriceHistoryRepo(q)
	erc1155HoldingRepo := erc1155Repository.NewHoldingRepo(q)
	coinGecko := coingecko.NewClient(&coingecko.ClientCfg{
		HttpClient: http.Client{},
		Timeout:    10 * time.Second,
	})
	s.im = NewFolderUsecase(folderRepo, relationRepo, nftRepo, collectionRepo, floorPriceRepo, coinGecko, erc1155HoldingRepo).(*folderUsecaseImpl)
}

func (s *folderUsecaseSuite) TestGetNFTsInFolder() {
	ctx := ctx.Background()
	cases := []struct {
		name            string
		folderId        string
		nftitemData     []nftitem.NftItem
		relationData    []account.FolderNftRelationship
		wantNftitemData []nftitem.NftitemWith1155Balance
	}{
		{
			name:     "find all Nfts",
			folderId: "folder1",
			nftitemData: []nftitem.NftItem{
				{
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
				},
				{
					ChainId:         1,
					ContractAddress: "0x456",
					TokenId:         "2",
				},
				{
					ChainId:         1,
					ContractAddress: "0x789",
					TokenId:         "3",
				},
			},
			relationData: []account.FolderNftRelationship{
				{
					FolderId:        "folder1",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
				},
				{
					FolderId:        "folder1",
					ChainId:         1,
					ContractAddress: "0x456",
					TokenId:         "2",
				},
				{
					FolderId:        "folder2",
					ChainId:         1,
					ContractAddress: "0x789",
					TokenId:         "3",
				},
			},
			wantNftitemData: []nftitem.NftitemWith1155Balance{
				{
					NftItem: nftitem.NftItem{
						ChainId:         1,
						ContractAddress: "0x123",
						TokenId:         "1",
					},
				},
				{
					NftItem: nftitem.NftItem{
						ChainId:         1,
						ContractAddress: "0x456",
						TokenId:         "2",
					},
				},
			},
		},
	}

	for _, c := range cases {
		_, err := s.query.RemoveAll(ctx, domain.TableFolders, bson.M{})
		s.Nil(err)
		_, err = s.query.RemoveAll(ctx, domain.TableFolderNftRelationships, bson.M{})
		s.Nil(err)
		_, err = s.query.RemoveAll(ctx, domain.TableNFTItems, bson.M{})
		s.Nil(err)

		err = s.query.Insert(ctx, domain.TableFolders, &account.Folder{
			Id:    c.folderId,
			Owner: "0x123",
		})
		s.Nil(err)

		for _, nft := range c.nftitemData {
			err := s.query.Insert(ctx, domain.TableNFTItems, &nft)
			s.Nil(err)
		}

		for _, relation := range c.relationData {
			err := s.query.Insert(ctx, domain.TableFolderNftRelationships, &relation)
			s.Nil(err)
		}

		res, err := s.im.GetNFTsInFolder(ctx, c.folderId)

		// don't care about objectID
		for i := 0; i < len(res); i++ {
			res[i].ObjectId = primitive.NilObjectID
		}

		s.Nil(err)
		s.ElementsMatch(c.wantNftitemData, res)
	}
}
