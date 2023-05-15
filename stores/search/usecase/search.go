package usecase

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/search"
	"github.com/x-xyz/goapi/service/query"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	maxAccounts    = 3
	maxCollections = 3
	maxTokens      = 10
)

type impl struct {
	q query.Mongo
}

func New(q query.Mongo) search.Usecase {
	return &impl{q: q}
}

func (im *impl) Search(c ctx.Ctx, keyword string, filter []string, collections []domain.Address) (*search.Result, error) {
	res := &search.Result{}
	for _, target := range filter {
		switch target {
		case search.Account:
			if accRes, err := im.SearchAccounts(c, keyword); err != nil {
				c.WithField("err", err).Error("SearchAccounts failed")
				return nil, err
			} else {
				res.Accounts = accRes.Accounts
			}
		case search.Collection:
			if colRes, err := im.SearchCollections(c, keyword, collections); err != nil {
				c.WithField("err", err).Error("SearchCollectinos failed")
				return nil, err
			} else {
				res.Collections = colRes.Collections
			}
		case search.Token:
			if tknRes, err := im.SearchTokens(c, keyword, collections); err != nil {
				c.WithField("err", err).Error("SearchTokens failed")
				return nil, err
			} else {
				res.Tokens = tknRes.Tokens
			}
		default:
			continue
		}
	}
	if len(filter) == 0 {
		if accRes, err := im.SearchAccounts(c, keyword); err != nil {
			c.WithField("err", err).Error("SearchAccounts failed")
			return nil, err
		} else {
			res.Accounts = accRes.Accounts
		}

		if colRes, err := im.SearchCollections(c, keyword, collections); err != nil {
			c.WithField("err", err).Error("SearchCollectinos failed")
			return nil, err
		} else {
			res.Collections = colRes.Collections
		}

		if tknRes, err := im.SearchTokens(c, keyword, collections); err != nil {
			c.WithField("err", err).Error("SearchTokens failed")
			return nil, err
		} else {
			res.Tokens = tknRes.Tokens
		}
	}

	return res, nil
}

func (im *impl) SearchAccounts(c ctx.Ctx, keyword string) (*search.Result, error) {
	res := &search.Result{}
	accs := []*account.Account{}

	searchStage := bson.D{{"$search", bson.M{
		"compound": bson.M{
			"must": bson.A{
				bson.M{"equals": bson.M{"path": "isAppropriate", "value": true}},
			},
			"should": bson.A{
				bson.M{"text": bson.M{
					"path":  []string{"address", "alias", "bio"},
					"query": keyword,
					"fuzzy": bson.M{"maxEdits": 1},
				}},
			},
		},
	}}}
	limitStage := bson.D{{"$limit", maxAccounts}}
	visibilityStage := bson.D{{"$project", bson.D{{"address", 1}, {"alias", 1}, {"imageHash", 1}, {"_id", 0}}}}
	iter, close, err := im.q.Pipe(c, domain.TableAccounts, mongo.Pipeline{searchStage, limitStage, visibilityStage})
	if err != nil {
		c.WithField("err", err).Error("q.Pipe failed")
		return nil, err
	}
	defer close()
	if err := iter.All(c, &accs); err != nil {
		c.WithField("err", err).Error("iter.Cursor.All failed")
		return nil, err
	}

	for _, acc := range accs {
		res.Accounts = append(res.Accounts, acc.ToSimpleAccount())
	}

	return res, nil
}

func (im *impl) SearchCollections(c ctx.Ctx, keyword string, collections []domain.Address) (*search.Result, error) {
	res := &search.Result{}
	var searchStage bson.D
	mustCond := bson.A{
		bson.M{"equals": bson.M{"path": "isAppropriate", "value": true}},
		bson.M{"equals": bson.M{"path": "status", "value": true}},
	}
	shouldCond := bson.A{
		bson.M{"text": bson.M{
			"path":  []string{"collectionName", "erc721Address", "description"},
			"query": keyword,
			"fuzzy": bson.M{"maxEdits": 1},
		}},
	}
	filterCond := bson.A{
		bson.M{"text": bson.M{
			"path":  []string{"erc721Address"},
			"query": collections,
		}},
	}
	if len(collections) == 0 {
		searchStage = bson.D{{"$search", bson.M{
			"compound": bson.M{
				"must":   mustCond,
				"should": shouldCond,
			},
		}}}
	} else {
		searchStage = bson.D{{"$search", bson.M{
			"compound": bson.M{
				"must":   mustCond,
				"should": shouldCond,
				"filter": filterCond,
			},
		}}}
	}
	limitStage := bson.D{{"$limit", maxCollections}}
	visibilityStage := bson.D{{"$project", bson.D{
		{"chainId", 1}, {"erc721Address", 1}, {"owner", 1},
		{"collectionName", 1}, {"description", 1}, {"categories", 1}, {"logoImageHash", 1}, {"logoImageUrl", 1},
		{"coverImageHash", 1}, {"coverImageUrl", 1}, {"siteUrl", 1}, {"discord", 1}, {"twitterHandle", 1}, {"instagramHandle", 1},
		{"mediumHandle", 1}, {"telegram", 1}, {"isVerified", 1}, {"royalty", 1}, {"feeRecipient", 1}, {"isRegistered", 1},
		{"supply", 1}, {"attributes", 1}, {"numOwners", 1}, {"totalVolume", 1}, {"floorPrice", 1}, {"usdFloorPrice", 1},
		{"hasFloorPrice", 1}, {"highestSale", 1}, {"highestSaleInUsd", 1}, {"lastSoldAt", 1}, {"hasBeenSold", 1},
		{"lastListedAt", 1}, {"hasBeenListed", 1}, {"viewCount", 1}, {"liked", 1},
		{"score", bson.M{"$meta": "searchScore"}}}}}
	iter, close, err := im.q.Pipe(c, domain.TableCollections, mongo.Pipeline{searchStage, limitStage, visibilityStage})
	if err != nil {
		c.WithField("err", err).Error("q.Pipe failed")
		return nil, err
	}
	defer close()
	if err := iter.All(c, &res.Collections); err != nil {
		c.WithField("err", err).Error("iter.Cursor.All failed")
		return nil, err
	}

	return res, nil
}

func (im *impl) SearchTokens(c ctx.Ctx, keyword string, collections []domain.Address) (*search.Result, error) {
	res := &search.Result{}
	tkns := []*nftitem.NftItem{}

	var searchStage bson.D
	const zeroAddress = "0x0000000000000000000000000000000000000000"

	mustCond := bson.M{"equals": bson.M{"path": "isAppropriate", "value": true}}
	mustNotCond := bson.M{"text": bson.M{"path": "owner", "query": zeroAddress}}
	shouldCond := bson.A{
		bson.M{"text": bson.M{
			"path":  []string{"name", "contractAddress"},
			"query": keyword,
			"fuzzy": bson.M{"maxEdits": 1},
		}},
	}

	filterCond := bson.A{
		bson.M{"text": bson.M{
			"path":  []string{"contractAddress"},
			"query": collections,
		}},
	}

	if len(collections) == 0 {
		searchStage = bson.D{{"$search", bson.M{
			"compound": bson.M{
				"must":    mustCond,
				"mustNot": mustNotCond,
				"should":  shouldCond,
			},
		}}}
	} else {
		searchStage = bson.D{{"$search", bson.M{
			"compound": bson.M{
				"must":    mustCond,
				"mustNot": mustNotCond,
				"should":  shouldCond,
				"filter":  filterCond,
			},
		}}}
	}

	limitStage := bson.D{{"$limit", maxTokens}}
	visibilityStage := bson.D{{"$project", bson.D{
		{"chainId", 1}, {"contractAddress", 1}, {"tokenID", 1}, {"name", 1}, {"tokenURI", 1}, {"thumbnailPath", 1},
		{"imagePath", 1}, {"imageURL", 1}, {"hostedImageURL", 1}, {"hostedTokenURI", 1}, {"_id", 0}, {"score", bson.M{"$meta": "searchScore"}},
	},
	}}
	iter, close, err := im.q.Pipe(c, domain.TableNFTItems, mongo.Pipeline{searchStage, limitStage, visibilityStage})
	if err != nil {
		c.WithField("err", err).Error("q.Pipe failed")
		return nil, err
	}

	defer close()

	if err := iter.All(c, &tkns); err != nil {
		c.WithField("err", err).Error("iter.Cursor.All failed")
		return nil, err
	}

	for _, tkn := range tkns {
		res.Tokens = append(res.Tokens, tkn.ToSimpleNftItem())
	}
	return res, nil
}
