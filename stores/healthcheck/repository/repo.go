package repository

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	hcdomain "github.com/x-xyz/goapi/domain/healthcheck"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/service/redis"
)

type impl struct {
	mgoClient  *mongoclient.Client
	redisCache redis.Service
}

// New creates new healthCheckUsecase object representation of HealthCheckUsecase interface
func New(
	mgoClient *mongoclient.Client,
	redisCache redis.Service,
) hcdomain.HealthCheckRepo {
	return &impl{
		mgoClient:  mgoClient,
		redisCache: redisCache,
	}
}

func (im *impl) PingDB(context ctx.Ctx) error {
	ctx, cancel := ctx.WithTimeout(context, 2*time.Second)
	defer cancel()
	if err := im.mgoClient.Ping(ctx, readpref.Primary()); err != nil {
		context.WithField("err", err).Error("ping mongo error")
		return err
	}

	if err := im.redisCache.Set(ctx, keys.RedisKey(keys.PfxHealthCheck, "testset"), []byte("1"), 30*time.Second); err != nil {
		context.WithField("err", err).Error("test redis set failed")
		return err
	}
	return nil
}
