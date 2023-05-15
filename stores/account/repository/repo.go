package repository

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/service/cache"
	"github.com/x-xyz/goapi/service/cache/provider"
	"github.com/x-xyz/goapi/service/cache/provider/compound"
	"github.com/x-xyz/goapi/service/cache/provider/primitive"
	redisCache "github.com/x-xyz/goapi/service/cache/provider/redis"
	"github.com/x-xyz/goapi/service/query"
	"github.com/x-xyz/goapi/service/redis"
)

type impl struct {
	query        query.Mongo
	accountCache cache.Service
}

// New creates new account repo
func New(query query.Mongo, redis redis.Service) account.Repo {
	cacheProviders := []provider.Provider{
		primitive.NewPrimitive("account", 128),
	}

	if redis != nil {
		cacheProviders = append(cacheProviders, redisCache.NewRedis(redis))
	}

	return &impl{
		query: query,
		accountCache: cache.New(cache.ServiceConfig{
			Ttl:   time.Hour,
			Pfx:   "account",
			Cache: compound.NewCompound(cacheProviders),
		}),
	}
}

func (im *impl) Get(c ctx.Ctx, address domain.Address) (*account.Account, error) {
	res := &account.Account{}

	if err := im.accountCache.GetByFunc(c, address.ToLowerStr(), res, func() (interface{}, error) {
		return im.get(c, address)
	}); err != nil {
		c.WithFields(log.Fields{
			"err":     err,
			"address": address,
		}).Error("accountCache.GetByFunc failed")
		return nil, err
	}

	return res, nil
}

func (im *impl) get(c ctx.Ctx, address domain.Address) (*account.Account, error) {
	a := &account.Account{}
	id := strings.ToLower(string(address))
	err := im.query.FindOne(c, domain.TableAccounts, bson.M{"address": id}, a)
	if err != nil && err != query.ErrNotFound {
		c.WithFields(log.Fields{
			"address": address,
			"err":     err,
		}).Error("find account failed")
	} else if err == query.ErrNotFound {
		err = domain.ErrNotFound
	}
	return a, err
}

func (im *impl) GetAccounts(c ctx.Ctx, addresses []domain.Address) ([]*account.Account, error) {
	inAddresses := make([]domain.Address, len(addresses))
	for i, addr := range addresses {
		inAddresses[i] = addr.ToLower()
	}
	accounts := []*account.Account{}
	if err := im.query.Search(c, domain.TableAccounts, 0, len(inAddresses),
		"", bson.M{"address": bson.M{"$in": inAddresses}}, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (im *impl) Insert(c ctx.Ctx, a *account.Account) error {
	a.Address = domain.Address(strings.ToLower(string(a.Address)))
	if err := im.query.Insert(c, domain.TableAccounts, a); err != nil {
		c.WithFields(log.Fields{
			"address": a.Address,
			"err":     err,
		}).Error("insert account failed")
		return err
	}
	return nil
}

func (im *impl) Update(c ctx.Ctx, address domain.Address, updater *account.Updater) error {
	updaterBson, err := mongoclient.MakeBsonM(updater)
	if err != nil {
		c.WithFields(log.Fields{
			"address": address,
			"err":     err,
		}).Error("make bsonM failed")
		return err
	}
	id := strings.ToLower(string(address))
	if err := im.query.Patch(c, domain.TableAccounts, bson.M{"address": id}, updaterBson); err != nil {
		c.WithFields(log.Fields{
			"address": address,
			"err":     err,
		}).Error("patch account failed")
		return err
	}
	if err := im.accountCache.Del(c, address.ToLowerStr()); err != nil {
		c.WithFields(log.Fields{
			"address": address,
			"err":     err,
		}).Error("accountCache.Del failed")
		return nil
	}
	return nil
}
