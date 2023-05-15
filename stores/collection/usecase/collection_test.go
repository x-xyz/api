package usecase

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/erc1155"
	erc721contract "github.com/x-xyz/goapi/domain/erc721/contract"
	"github.com/x-xyz/goapi/service/query"
	collectionRepository "github.com/x-xyz/goapi/stores/collection/repository"
	erc1155Repository "github.com/x-xyz/goapi/stores/erc1155/repository"
	tokenRepository "github.com/x-xyz/goapi/stores/token/repository"
)

type collectionTestSuite struct {
	suite.Suite
	db     *mongoclient.Client
	dbName string
	q      query.Mongo

	address721  domain.Address
	address1155 domain.Address

	collectionRepo collection.Repo
	erc721Repo     erc721contract.Repo
	erc1155Repo    erc1155.Repo
	collectionUC   collection.Usecase
}

func Test(t *testing.T) {
	suite.Run(t, new(collectionTestSuite))
}

func (s *collectionTestSuite) SetupSuite() {
	// to setup mongodb replica set at local quickly, run the following command
	// npm install run-rs -g
	// run-rs -v 4.4.0
	// uri := "mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=rs"
	uri := "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"
	authDBName := "admin"
	s.dbName = "test-collection-usecase"
	s.db = mongoclient.MustConnectMongoClient(uri, authDBName, s.dbName, false, true, 2)
	q := query.New(s.db, false)
	s.q = q

	collectionRepo := collectionRepository.NewCollection(q)
	registrationRepo := collectionRepository.NewRegistration(q)

	erc721Repo := collectionRepository.NewErc721Contract(q)
	erc1155Repo := erc1155Repository.NewContractRepo(q)
	erc1155HoldingRepo := erc1155Repository.NewHoldingRepo(q)

	s.collectionRepo = collectionRepo
	s.erc721Repo = erc721Repo
	s.erc1155Repo = erc1155Repo

	// bayc
	s.address721 = "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"
	// serum
	s.address1155 = "0x22c36bfdcef207f9c0cc941936eff94d4246d14a"
	// mock chain service
	erc721ChainService := newErc721(s.address721)
	erc1155ChainService := newErc1155(s.address1155)

	nftitemRepo := tokenRepository.NewNftItem(q, nil)

	s.collectionUC = NewCollection(&CollectionUseCaseCfg{
		CollectionRepo:      collectionRepo,
		RegistrationRepo:    registrationRepo,
		Erc721contractRepo:  erc721Repo,
		Erc1155contractRepo: erc1155Repo,
		Erc1155holdingRepo:  erc1155HoldingRepo,
		Erc721ChainService:  erc721ChainService,
		Erc1155ChainService: erc1155ChainService,
		NftitemRepo:         nftitemRepo,
	})
}

func (s *collectionTestSuite) TearDownSuite() {
	s.Require().NoError(s.db.Database(s.dbName).Drop(bCtx.Background()))
}

func (s *collectionTestSuite) TestRegisterTokenType() {
	testcases := []struct {
		name          string
		address       domain.Address
		tokenType     domain.TokenType
		wantTokenType domain.TokenType
		err           error
	}{
		{
			name:          "test erc721 register default token type is erc721",
			address:       s.address721,
			wantTokenType: domain.TokenType721,
			err:           nil,
		},
		{
			name:          "test erc721 correct token type",
			address:       s.address721,
			tokenType:     domain.TokenType721,
			wantTokenType: domain.TokenType721,
			err:           nil,
		},
		{
			name:      "test erc721 incorrect token type",
			address:   s.address721,
			tokenType: domain.TokenType1155,
			err:       domain.ErrErc1155InterfaceUnsupported,
		},
		{
			name:    "test erc1155 register default token type is erc721",
			address: s.address1155,
			err:     domain.ErrErc721InterfaceUnsupported,
		},
		{
			name:          "test erc1155 correct token type",
			address:       s.address1155,
			tokenType:     domain.TokenType1155,
			wantTokenType: domain.TokenType1155,
			err:           nil,
		},
		{
			name:      "test erc1155 incorrect token type",
			address:   s.address1155,
			tokenType: domain.TokenType721,
			err:       domain.ErrErc721InterfaceUnsupported,
		},
	}

	for _, tc := range testcases {
		s.T().Run(tc.name, func(t *testing.T) {
			registration := collection.Registration{
				ChainId:       domain.ChainId(1),
				Erc721Address: tc.address,
				TokenType:     tc.tokenType,
			}
			r, err := s.collectionUC.Register(bCtx.Background(), registration)
			s.Require().Equal(tc.err, err)
			if err == nil {
				s.Require().Equal(registration.ChainId, r.ChainId)
				s.Require().Equal(registration.Erc721Address, r.Erc721Address)
				s.Require().Equal(tc.wantTokenType, r.TokenType)
			}

			// drop database
			s.Require().NoError(s.db.Database(s.dbName).Drop(bCtx.Background()))
		})
	}
}

func (s *collectionTestSuite) TestAccept() {
	testcases := []struct {
		name         string
		registration collection.Registration
	}{
		{
			name: "accept erc721",
			registration: collection.Registration{
				ChainId:       domain.ChainId(1),
				Erc721Address: s.address721,
				TokenType:     domain.TokenType721,
				FeeRecipient:  "0x70e105cE24A8D0099eeA39F9d6C4B23Ec561d833",
			},
		},
		{
			name: "accept erc1155",
			registration: collection.Registration{
				ChainId:       domain.ChainId(1),
				Erc721Address: s.address1155,
				TokenType:     domain.TokenType1155,
				FeeRecipient:  "0x70e105cE24A8D0099eeA39F9d6C4B23Ec561d833",
			},
		},
	}

	for _, tc := range testcases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.collectionUC.Register(bCtx.Background(), tc.registration)
			s.Require().NoError(err)

			cid := collection.CollectionId{
				ChainId: tc.registration.ChainId,
				Address: tc.registration.Erc721Address,
			}
			fmt.Println(cid.ChainId)
			fmt.Println(cid.Address)

			_, err = s.collectionUC.Accept(bCtx.Background(), cid)
			s.Require().NoError(err)

			c, err := s.collectionRepo.FindOne(bCtx.Background(), cid)
			s.Require().NoError(err)
			s.Require().Equal(tc.registration.TokenType, c.TokenType)

			// test related contract inserted correctly
			if c.TokenType == domain.TokenType721 {
				erc721, err := s.erc721Repo.FindOne(bCtx.Background(), erc721contract.WithChainId(c.ChainId), erc721contract.WithAddress(c.Erc721Address))
				s.Require().NoError(err)
				s.Require().Equal(c.ChainId, erc721.ChainId)
				s.Require().Equal(c.Erc721Address, erc721.Address)
			} else if c.TokenType == domain.TokenType1155 {
				erc1155, err := s.erc1155Repo.FindOne(bCtx.Background(), erc1155.WithChainId(c.ChainId), erc1155.WithAddress(c.Erc721Address))
				s.Require().NoError(err)
				s.Require().Equal(c.ChainId, erc1155.ChainId)
				s.Require().Equal(c.Erc721Address, erc1155.Address)
			}
		})
	}

	s.Require().NoError(s.db.Database(s.dbName).Drop(bCtx.Background()))
}

func (s *collectionTestSuite) TestFindAllIncludingUnregistered() {
	err := s.erc721Repo.Create(bCtx.Background(), erc721contract.Contract{
		ChainId: domain.ChainId(1),
		Address: s.address721,
	})
	s.Require().NoError(err)

	err = s.erc1155Repo.Create(bCtx.Background(), erc1155.Contract{
		ChainId: domain.ChainId(3),
		Address: s.address1155,
	})
	s.Require().NoError(err)

	cs, err := s.collectionUC.FindAllIncludingUnregistered(bCtx.Background(), collection.WithChainId(1))
	s.Require().NoError(err)
	s.Require().Len(cs, 1)
	s.Require().Equal(domain.ChainId(1), cs[0].ChainId)
	s.Require().Equal(s.address721, cs[0].Erc721Address)
	s.Require().Equal(domain.TokenType721, cs[0].TokenType)

	cs, err = s.collectionUC.FindAllIncludingUnregistered(bCtx.Background(), collection.WithChainId(3))
	s.Require().NoError(err)
	s.Require().Len(cs, 1)
	s.Require().Equal(domain.ChainId(3), cs[0].ChainId)
	s.Require().Equal(s.address1155, cs[0].Erc721Address)
	s.Require().Equal(domain.TokenType1155, cs[0].TokenType)

	cs, err = s.collectionUC.FindAllIncludingUnregistered(bCtx.Background(), collection.WithChainId(10))
	s.Require().NoError(err)
	s.Require().Len(cs, 0)

	s.Require().NoError(s.db.Database(s.dbName).Drop(bCtx.Background()))
}

func (s *collectionTestSuite) TestFindOne() {
	err := s.erc721Repo.Create(bCtx.Background(), erc721contract.Contract{
		ChainId: domain.ChainId(1),
		Address: s.address721,
	})
	s.Require().NoError(err)

	err = s.erc1155Repo.Create(bCtx.Background(), erc1155.Contract{
		ChainId: domain.ChainId(3),
		Address: s.address1155,
	})
	s.Require().NoError(err)

	c, err := s.collectionUC.FindOne(bCtx.Background(), collection.CollectionId{ChainId: 1, Address: s.address721})
	s.Require().NoError(err)
	s.Require().Equal(domain.ChainId(1), c.ChainId)
	s.Require().Equal(s.address721, c.Erc721Address)
	s.Require().Equal(domain.TokenType721, c.TokenType)

	c, err = s.collectionUC.FindOne(bCtx.Background(), collection.CollectionId{ChainId: 3, Address: s.address1155})
	s.Require().NoError(err)
	s.Require().Equal(domain.ChainId(3), c.ChainId)
	s.Require().Equal(s.address1155, c.Erc721Address)
	s.Require().Equal(domain.TokenType1155, c.TokenType)

	s.Require().NoError(s.db.Database(s.dbName).Drop(bCtx.Background()))
}

func (s *collectionTestSuite) TestUpdateTraitFloorPrice() {
	ctx := bCtx.Background()
	collectionId := collection.CollectionId{
		ChainId: 1,
		Address: "0xabc123",
	}

	s.q.Insert(ctx, domain.TableCollections, collection.Collection{
		ChainId:        collectionId.ChainId,
		Erc721Address:  collectionId.Address,
		CollectionName: "test collection",
	})

	err := s.collectionUC.UpdateTraitFloorPrice(ctx, collectionId, "test_trait_name", "test_trait_value", 123.123)
	s.Nil(err)

	output, err := s.collectionRepo.FindOne(ctx, collectionId)
	s.Nil(err)
	s.Equal(output.TraitFloorPrice, map[string]map[string]float64{
		"test_trait_name": {
			"test_trait_value": 123.123,
		},
	})

	err = s.collectionUC.UpdateTraitFloorPrice(ctx, collectionId, "test_trait_name2", "test_trait_value2", 456.456)
	s.Nil(err)

	output, err = s.collectionRepo.FindOne(ctx, collectionId)
	s.Nil(err)
	s.Equal(output.TraitFloorPrice, map[string]map[string]float64{
		"test_trait_name": {
			"test_trait_value": 123.123,
		},
		"test_trait_name2": {
			"test_trait_value2": 456.456,
		},
	})
}

type mockChainService struct {
	erc721  domain.Address
	erc1155 domain.Address
}

func newErc721(address domain.Address) *mockChainService {
	return &mockChainService{
		erc721: address,
	}
}

func newErc1155(address domain.Address) *mockChainService {
	return &mockChainService{
		erc1155: address,
	}
}

func (m *mockChainService) Supports721Interface(ctx bCtx.Ctx, chainId int32, addr string) (bool, error) {
	return addr == m.erc721.ToLowerStr(), nil
}

func (m *mockChainService) OwnerOf(ctx bCtx.Ctx, chainId int32, addr string, tokenId *big.Int) (string, error) {
	return "", nil
}

func (m *mockChainService) Supports1155Interface(ctx bCtx.Ctx, chainId int32, addr string) (bool, error) {
	return addr == m.erc1155.ToLowerStr(), nil
}
