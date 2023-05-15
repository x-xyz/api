package paging

import (
	"errors"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/redis"
)

var (
	ErrBadCursor    = errors.New("bad cursor")
	ErrGetLatestKey = errors.New("failed to get latestKey")
	// ErrBadContainer indicates container is not pointer of empty slice
	ErrBadContainer = errors.New("bad container")
)

/*
	Cursor structure:
	- cursor:
		"<createTs>:<totalCount>:<offset>" base64 encoded
	- latest key:
		"pagingService:la:<keyPfx>:<key>" return first cursor of page
	- lock key:
		"pagingService:lock:<keyPfx>:<key>"
	- cache key:
		"pagingService:<keyPfx>:<key>:<createTs>:<shardNum>"

	Timestamp all store in nanasecond
*/

// A Getter loads data for a key.
// `wholeList` should be slice of (object / object pointer).
//     i.e. []model.GameRow or []*model.GameRow
type Getter func(ctx ctx.Ctx, key string) (wholeList interface{}, err error)

type PagingConfig struct {
	RedisCache redis.Service
	KeyPfx     string
	Getter     Getter
	// renewDuration determines the time a page need to be rebuilt
	RenewDuration time.Duration
	// cacheDuration is the ttl of all snapshots
	CacheDuration time.Duration

	ShardSize int

	// ----- Optional Config -----
	// Default 10s.
	// ErrGetterTimeout returned if getter spend more than this time.
	GetterTimeout time.Duration
}

// Service provides manipulation interface of paging
type Service interface {
	// GetPage gets a page of the underlying slice. For the first request, cursor is an empty string.
	// `container`, which holds the result, should be a pointer to slice of object or object pointer.
	//     i.e. *[]model.GameRow or  *[]*model.GameRow
	Get(
		context ctx.Ctx, key string, cursor string, size int, container interface{},
	) (nextCursor string, totalCount int, err error)

	// Update triggers the paging to rebuild newest snapshot,
	Update(context ctx.Ctx) error
}
