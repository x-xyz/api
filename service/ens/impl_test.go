package ens

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/redisclient"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/redis"
)

type ensSuite struct {
	suite.Suite

	im *impl
}

func (s *ensSuite) SetupSuite() {
	redisCacheName := "cache"
	redisCacheURI := "localhost:6379"
	redisCachePwd := ""
	redisCachePool := redisclient.MustConnectRedis(redisCacheURI, redisCachePwd, redisclient.RedisParam{
		PoolMultiplier: 20,
		Retry:          true,
	})

	redisCache := redis.New(redisCacheName, metrics.New(redisCacheName), &redis.Pools{
		Src: redisCachePool,
	})

	s.im = New("rpc_url", redisCache).(*impl)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(ensSuite))
}

func (s *ensSuite) TestResolve() {
	name := "machibigbrother.eth"
	address := domain.Address("0x020cA66C30beC2c4Fe3861a94E4DB4A498A35872")

	res, err := s.im.Resolve(ctx.Background(), name)
	if s.NoError(err) {
		s.Equal(address.ToLowerStr(), res.ToLowerStr())
	}
}

func (s *ensSuite) TestReverseResolve() {
	name := "machibigbrother.eth"
	address := domain.Address("0x020cA66C30beC2c4Fe3861a94E4DB4A498A35872")

	res, err := s.im.ReverseResolve(ctx.Background(), address)
	if s.NoError(err) {
		s.Equal(name, res)
	}
}
