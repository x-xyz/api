package repository

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type relationSuite struct {
	suite.Suite

	query query.Mongo
	im    *folderNftRelationshipImpl
}

func TestRelationSuite(t *testing.T) {
	suite.Run(t, new(relationSuite))
}

func (s *relationSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "myFirstDatabase"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q
	s.im = NewFolderNftRelationshipRepo(q).(*folderNftRelationshipImpl)
}

func (s *relationSuite) SetupTest() {
	s.query.RemoveAll(ctx.Background(), domain.TableFolderNftRelationships, bson.M{})
}

func (s *relationSuite) TestInsert() {
	ctx := ctx.Background()
	folderID := "testfolderid"
	relation := account.FolderNftRelationship{
		FolderId:        folderID,
		ChainId:         1,
		ContractAddress: "0x3C20083Ebfb04AFA103b8f4C0A9fd9879367742E",
		TokenId:         "1",
		Index:           1,
	}

	err := s.im.Insert(ctx, &relation)
	s.Nil(err)

	output := account.FolderNftRelationship{}
	err = s.query.FindOne(ctx, domain.TableFolderNftRelationships, bson.M{"folderId": folderID}, &output)
	s.Nil(err)
	s.Equal(relation, output)
}

func (s *relationSuite) TestGetAllRelationsByFolderID() {
	ctx := ctx.Background()
	mockFolderID1 := "mockFolderID1"
	mockFolderID2 := "mockFolderID2"
	mockRelations := []account.FolderNftRelationship{
		{
			FolderId:        mockFolderID1,
			ChainId:         1,
			ContractAddress: "0x3C20083Ebfb04AFA103b8f4C0A9fd9879367742E",
			TokenId:         "1",
			Index:           1,
		},
		{
			FolderId:        mockFolderID1,
			ChainId:         1,
			ContractAddress: "0x3C20083Ebfb04AFA103b8f4C0A9fd9879367742E",
			TokenId:         "2",
			Index:           2,
		},
		{
			FolderId:        mockFolderID2,
			ChainId:         1,
			ContractAddress: "0x3C20083Ebfb04AFA103b8f4C0A9fd9879367742E",
			TokenId:         "3",
			Index:           1,
		},
	}

	bulkOps := []query.UpsertOp{}
	for _, r := range mockRelations {
		m, err := mongoclient.MakeBsonM(r)
		s.Nil(err)

		bulkOps = append(bulkOps, query.UpsertOp{
			Selector: m,
			Updater:  m,
		})
	}

	_, _, err := s.query.BulkUpsert(ctx, domain.TableFolderNftRelationships, bulkOps)
	s.Nil(err)

	r1, err := s.im.GetAllRelations(ctx, account.WithFolderId(mockFolderID1))
	s.Nil(err)
	s.Equal(2, len(r1))
	s.Equal(mockRelations[0], *r1[0])
	s.Equal(mockRelations[1], *r1[1])

	r2, err := s.im.GetAllRelations(ctx, account.WithFolderId(mockFolderID2))
	s.Nil(err)
	s.Equal(1, len(r2))
	s.Equal(mockRelations[2], *r2[0])
}

func (s *relationSuite) TestDeleteAllRelationsByFolderID() {
	ctx := ctx.Background()
	mockFolderID1 := "mockFolderID1"
	mockFolderID2 := "mockFolderID2"
	mockRelations := []account.FolderNftRelationship{
		{
			FolderId:        mockFolderID1,
			ChainId:         1,
			ContractAddress: "0x3C20083Ebfb04AFA103b8f4C0A9fd9879367742E",
			TokenId:         "1",
			Index:           1,
		},
		{
			FolderId:        mockFolderID1,
			ChainId:         1,
			ContractAddress: "0x3C20083Ebfb04AFA103b8f4C0A9fd9879367742E",
			TokenId:         "2",
			Index:           2,
		},
		{
			FolderId:        mockFolderID2,
			ChainId:         1,
			ContractAddress: "0x3C20083Ebfb04AFA103b8f4C0A9fd9879367742E",
			TokenId:         "3",
			Index:           1,
		},
	}

	bulkOps := []query.UpsertOp{}
	for _, r := range mockRelations {
		m, err := mongoclient.MakeBsonM(r)
		s.Nil(err)

		bulkOps = append(bulkOps, query.UpsertOp{
			Selector: m,
			Updater:  m,
		})
	}

	_, _, err := s.query.BulkUpsert(ctx, domain.TableFolderNftRelationships, bulkOps)
	s.Nil(err)

	res := []*account.FolderNftRelationship{}
	err = s.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "", bson.M{}, &res)
	s.Nil(err)
	s.Equal(3, len(res))

	err = s.im.DeleteAllRelationsByFolderID(ctx, mockFolderID1)
	s.Nil(err)

	err = s.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "", bson.M{}, &res)
	s.Nil(err)
	s.Equal(1, len(res))
}

func (s *relationSuite) TestDeleteAllRelationsByNftitemID() {
	ctx := ctx.Background()
	mockFolderID1 := "mockFolderID1"
	mockFolderID2 := "mockFolderID2"
	mockNftitemID1 := nftitem.Id{ChainId: 1, ContractAddress: "0x3C20083Ebfb04AFA103b8f4C0A9fd9879367742E", TokenId: "1"}
	mockNftitemID2 := nftitem.Id{ChainId: 1, ContractAddress: "0x3C20083Ebfb04AFA103b8f4C0A9fd9879367742E", TokenId: "2"}
	mockRelations := []account.FolderNftRelationship{
		{
			FolderId:        mockFolderID1,
			ChainId:         mockNftitemID1.ChainId,
			ContractAddress: mockNftitemID1.ContractAddress,
			TokenId:         mockNftitemID1.TokenId,
			Index:           1,
		},
		{
			FolderId:        mockFolderID1,
			ChainId:         mockNftitemID2.ChainId,
			ContractAddress: mockNftitemID2.ContractAddress,
			TokenId:         mockNftitemID2.TokenId,
			Index:           2,
		},
		{
			FolderId:        mockFolderID2,
			ChainId:         mockNftitemID1.ChainId,
			ContractAddress: mockNftitemID1.ContractAddress,
			TokenId:         mockNftitemID1.TokenId,
			Index:           1,
		},
		{
			FolderId:        mockFolderID2,
			ChainId:         mockNftitemID2.ChainId,
			ContractAddress: mockNftitemID2.ContractAddress,
			TokenId:         mockNftitemID2.TokenId,
			Index:           2,
		},
	}

	bulkOps := []query.UpsertOp{}
	for _, r := range mockRelations {
		m, err := mongoclient.MakeBsonM(r)
		s.Nil(err)

		bulkOps = append(bulkOps, query.UpsertOp{
			Selector: m,
			Updater:  m,
		})
	}

	_, _, err := s.query.BulkUpsert(ctx, domain.TableFolderNftRelationships, bulkOps)
	s.Nil(err)

	res := []*account.FolderNftRelationship{}
	err = s.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "", bson.M{}, &res)
	s.Nil(err)
	s.Equal(4, len(res))

	err = s.im.DeleteAllRelationsByNftitem(ctx, mockNftitemID1)
	s.Nil(err)

	err = s.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "", bson.M{}, &res)
	s.Nil(err)
	s.Equal(2, len(res))
}

func (s *relationSuite) TestDeleteAll() {
	ctx := ctx.Background()
	cases := []struct {
		name    string
		options []account.RelationsQueryOptionsFunc
		data    []account.FolderNftRelationship
		want    []account.FolderNftRelationship
	}{
		{
			name: "delete with folderIds",
			options: []account.RelationsQueryOptionsFunc{
				account.WithFolderIds([]string{
					"folderId1",
					"folderId2",
				}),
			},
			data: []account.FolderNftRelationship{
				{
					FolderId:        "folderId1",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
					Index:           0,
				},
				{
					FolderId:        "folderId2",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Index:           0,
				},
				{
					FolderId:        "folderId3",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "3",
					Index:           0,
				},
			},
			want: []account.FolderNftRelationship{
				{
					FolderId:        "folderId3",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "3",
					Index:           0,
				},
			},
		},
		{
			name: "delete with folderIds with nftitemId",
			options: []account.RelationsQueryOptionsFunc{
				account.WithFolderIds([]string{
					"folderId1",
					"folderId2",
				}),
				account.WithNftitemId(nftitem.Id{
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
				}),
			},
			data: []account.FolderNftRelationship{
				{
					FolderId:        "folderId1",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
					Index:           0,
				},
				{
					FolderId:        "folderId2",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "1",
					Index:           0,
				},
				{
					FolderId:        "folderId1",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Index:           0,
				},
				{
					FolderId:        "folderId3",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "3",
					Index:           0,
				},
			},
			want: []account.FolderNftRelationship{
				{
					FolderId:        "folderId1",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "2",
					Index:           0,
				},
				{
					FolderId:        "folderId3",
					ChainId:         1,
					ContractAddress: "0x123",
					TokenId:         "3",
					Index:           0,
				},
			},
		},
	}

	for _, c := range cases {
		_, err := s.query.RemoveAll(ctx, domain.TableFolderNftRelationships, bson.M{})
		s.Nil(err)

		for _, d := range c.data {
			err := s.query.Insert(ctx, domain.TableFolderNftRelationships, &d)
			s.Nil(err)
		}

		err = s.im.DeleteAll(ctx, c.options...)
		s.Nil(err)

		res := []account.FolderNftRelationship{}
		err = s.im.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "", bson.M{}, &res)
		s.Nil(err)

		s.ElementsMatch(c.want, res)
	}
}

func (s *relationSuite) TestAddNftitemsToFolder() {
	ctx := ctx.Background()

	mockNftItemIds := []nftitem.Id{
		{
			ChainId:         1,
			ContractAddress: "0xtest",
			TokenId:         "1",
		},
		{
			ChainId:         1,
			ContractAddress: "0xtest",
			TokenId:         "2",
		},
		{
			ChainId:         1,
			ContractAddress: "0xtest",
			TokenId:         "3",
		},
	}
	mockFolderId1 := "mockFolderId1"
	mockFolderId2 := "mockFolderId2"

	s.query.BulkUpsert(ctx, domain.TableFolderNftRelationships, []query.UpsertOp{
		{
			Selector: bson.M{
				"chainId":         1,
				"contractAddress": "0xtest",
				"tokenId":         "1",
			},
			Updater: bson.M{
				"folderId":        mockFolderId1,
				"chainId":         1,
				"contractAddress": "0xtest",
				"tokenId":         "1",
				"index":           0,
			},
		},
		{
			Selector: bson.M{
				"chainId":         1,
				"contractAddress": "0xtest",
				"tokenId":         "2",
			},
			Updater: bson.M{
				"folderId":        mockFolderId1,
				"chainId":         1,
				"contractAddress": "0xtest",
				"tokenId":         "2",
				"index":           1,
			},
		},
		{
			Selector: bson.M{
				"chainId":         1,
				"contractAddress": "0xtest",
				"tokenId":         "3",
			},
			Updater: bson.M{
				"folderId":        mockFolderId2,
				"chainId":         1,
				"contractAddress": "0xtest",
				"tokenId":         "3",
				"index":           0,
			},
		},
	})

	relations := []account.FolderNftRelationship{}
	err := s.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "", bson.M{"folderId": mockFolderId1}, &relations)
	s.Nil(err)
	s.Equal(2, len(relations))

	err = s.im.AddNftitemsToFolder(ctx, mockNftItemIds, mockFolderId2)
	s.Nil(err)

	expectedRelations := []account.FolderNftRelationship{
		{
			FolderId:        mockFolderId2,
			ChainId:         1,
			ContractAddress: "0xtest",
			TokenId:         "1",
			Index:           0,
		},
		{
			FolderId:        mockFolderId2,
			ChainId:         1,
			ContractAddress: "0xtest",
			TokenId:         "2",
			Index:           1,
		},
		{
			FolderId:        mockFolderId2,
			ChainId:         1,
			ContractAddress: "0xtest",
			TokenId:         "3",
			Index:           2,
		},
	}

	actualRelations := []account.FolderNftRelationship{}
	err = s.im.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "index", bson.M{"folderId": mockFolderId2}, &actualRelations)
	s.Nil(err)

	s.Equal(expectedRelations, actualRelations)
}

func (s *relationSuite) TestMoveNftitems() {
	ctx := ctx.Background()

	mockFolderId1 := "mockFolderId1"
	mockFolderId2 := "mockFolderId2"
	mockRelations := []account.FolderNftRelationship{
		{
			FolderId:        mockFolderId1,
			ChainId:         1,
			ContractAddress: "0x123",
			TokenId:         "1",
			Index:           1,
		},
		{
			FolderId:        mockFolderId1,
			ChainId:         1,
			ContractAddress: "0x123",
			TokenId:         "2",
			Index:           2,
		},
		{
			FolderId:        mockFolderId2,
			ChainId:         1,
			ContractAddress: "0x123",
			TokenId:         "3",
			Index:           1,
		},
	}
	mockItemIds := []nftitem.Id{
		{
			ChainId:         1,
			ContractAddress: "0x123",
			TokenId:         "1",
		},
		{
			ChainId:         1,
			ContractAddress: "0x123",
			TokenId:         "2",
		},
	}

	for _, r := range mockRelations {
		err := s.query.Insert(ctx, domain.TableFolderNftRelationships, &r)
		s.Nil(err)
	}

	err := s.im.MoveNftitems(ctx, mockItemIds, mockFolderId1, mockFolderId2)
	s.Nil(err)

	folderRelations1 := []account.FolderNftRelationship{}
	err = s.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "", bson.M{"folderId": mockFolderId1}, &folderRelations1)
	s.Nil(err)
	s.Equal(0, len(folderRelations1))

	folderRelations2 := []account.FolderNftRelationship{}
	err = s.query.Search(ctx, domain.TableFolderNftRelationships, 0, 0, "", bson.M{"folderId": mockFolderId2}, &folderRelations2)
	s.Nil(err)
	s.Equal(3, len(folderRelations2))
}
