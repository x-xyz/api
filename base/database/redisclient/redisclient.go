package redisclient

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/x-xyz/goapi/base/log"
)

// The constant
const (
	dialTimeout  = 2 * time.Second
	readTimeout  = 1500 * time.Millisecond
	writeTimeout = 1500 * time.Millisecond
)

// RedisParam is the optional param for redis connection
type RedisParam struct {
	PoolMultiplier float64
	Retry          bool
}

// MustConnectRedis connects to one redis uri
// NOTE This function panics if the connection fails.
func MustConnectRedis(uri, password string, param ...RedisParam) *redis.Pool {
	p, err := ConnectRedis(uri, password, param...)
	if err != nil {
		log.Log().WithFields(log.Fields{"redisURI": uri, "err": err}).Panic("fail to dial Redis")
	}
	return p
}

// ConnectRedis connects to one redis uri
func ConnectRedis(uri, password string, param ...RedisParam) (*redis.Pool, error) {
	maxIdle := 200
	maxActive := 1024
	retry := false
	if len(param) > 0 {
		cpu := float64(runtime.NumCPU())
		// allowing 25% idle connection
		maxIdle = int(cpu * param[0].PoolMultiplier / 4)
		maxActive = int(cpu * param[0].PoolMultiplier)
		retry = param[0].Retry
	}

	opts := []redis.DialOption{
		redis.DialConnectTimeout(dialTimeout),
		redis.DialReadTimeout(readTimeout),
		redis.DialWriteTimeout(writeTimeout),
	}
	if password != "" {
		opts = append(opts, redis.DialPassword(password))
	}
	p := &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", uri, opts...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			// No need to test if it's been recycled less than 1 sec.
			if time.Since(t) < time.Second {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	// NOTE: In k8s, a small amount of containers will fail to connect redis
	// caused by network issue, so we retry 3 times here.
	retryCount := 3

	var c redis.Conn
	var dialErr error

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	// at least 1 second
	sleepOffset := 1 * time.Second

	for i := retryCount; i >= 0; i-- {
		if i < retryCount {
			// NOTE retry is currently true for all environments. It's only false
			// when running unit tests.
			if retry == false {
				break
			}
			randSleep := r1.Float32() * 1000
			time.Sleep(time.Duration(randSleep)*time.Millisecond + sleepOffset)
		}
		// Do a TestOnBorrow to make sure connecion is okay
		c, dialErr = p.Dial()
		if dialErr != nil {
			log.Log().WithFields(log.Fields{
				"redisURI": uri,
				"err":      dialErr,
				"retry":    i,
			}).Error("fail to dial Redis")
			continue
		}
		defer c.Close()
		if dialErr = p.TestOnBorrow(c, time.Now()); dialErr != nil {
			log.Log().WithFields(log.Fields{
				"redisURI": uri,
				"err":      dialErr,
				"retry":    i,
			}).Error("fail to TestOnBorrow Redis")
			continue
		}
		if dialErr == nil {
			break
		}
	}
	if dialErr != nil {
		log.Log().WithFields(log.Fields{
			"redisURI": uri,
			"err":      dialErr,
		}).Error("fail to dial Redis")
		return nil, dialErr
	}

	log.Log().WithField("redisURI", uri).Info("redis connected")

	return p, nil
}
