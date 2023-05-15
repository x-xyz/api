package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
)

type folderSuite struct {
	suite.Suite

	query query.Mongo
	im    *folderReopImpl
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(folderSuite))
}

func (s *folderSuite) SetupSuite() {
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	dbName := "myFirstDatabase"
	enableSSL := false
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, false)

	s.query = q
	s.im = NewFolderRepo(q).(*folderReopImpl)
}

func (s *folderSuite) SetupTest() {
	s.query.RemoveAll(ctx.Background(), domain.TableFolders, bson.M{})
}

func (s *folderSuite) TestFolderRepo() {
	ctx := ctx.Background()

	// insert
	ID := "testid"
	owner := domain.Address("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	folder := account.Folder{
		Id:                    ID,
		Name:                  "testname",
		IsPrivate:             false,
		IsBuiltIn:             false,
		FloorPriceInUsd:       123,
		TotalValueInUsd:       456,
		InstantLiquidityInUsd: 789,
		Owner:                 owner,
		CreatedAt:             time.Unix(123, 0).UTC(),
	}

	err := s.im.Insert(ctx, &folder)
	s.Nil(err, "folder insert failed")

	folder.Owner = folder.Owner.ToLower()

	// Get
	folderFromGet, err := s.im.Get(ctx, folder.Id)
	s.Nil(err)
	s.Equal(folder, *folderFromGet)

	// GetFolders
	folders, err := s.im.GetFolders(ctx, account.WithOwner(owner))
	s.Nil(err, "failed to GetFolders")
	s.Equal(folder, *folders[0], "folder not exists")

	// Update folder
	newName := "testnewname"
	updater := &account.FolderUpdater{
		Name: ptr.String(newName),
	}
	err = s.im.Update(ctx, ID, updater)
	s.Nil(err, "failed to update folder")

	folder.Name = newName
	folders, err = s.im.GetFolders(ctx, account.WithOwner(owner))
	s.Nil(err, "failed to GetFolders")
	s.Equal(folder, *folders[0])

	// Delete folder
	err = s.im.Delete(ctx, ID)
	s.Nil(err, "failed to Delete folder")

	folders, err = s.im.GetFolders(ctx, account.WithOwner(owner))
	s.Nil(err, "failed to GetFolders")
	s.Equal(0, len(folders))

	builtInFolder := account.Folder{
		Id:                    "testBuiltInFolder",
		Name:                  "builtinFolder",
		IsPrivate:             false,
		IsBuiltIn:             true,
		FloorPriceInUsd:       123,
		TotalValueInUsd:       456,
		InstantLiquidityInUsd: 789,
		Owner:                 owner,
		CreatedAt:             time.Unix(123, 0).UTC(),
	}

	// insert builtIn folder
	err = s.im.Insert(ctx, &builtInFolder)
	s.Nil(err)

	// GetFolders withBuiltIn
	folders, err = s.im.GetFolders(ctx, account.WithBuiltIn(true), account.WithOwner(owner))
	s.Nil(err)
	s.Equal(1, len(folders))
	s.Equal(builtInFolder, *folders[0])

	// GetFolders withPrivate
	folders, err = s.im.GetFolders(ctx, account.WithBuiltIn(true), account.WithOwner(owner), account.WithPrivate(true))
	s.Nil(err)
	s.Equal(0, len(folders))

	folders, err = s.im.GetFolders(ctx, account.WithBuiltIn(true), account.WithOwner(owner), account.WithPrivate(false))
	s.Nil(err)
	s.Equal(1, len(folders))
	s.Equal(builtInFolder, *folders[0])
}
