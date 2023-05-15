package provider

import (
	"errors"
	"time"

	"github.com/x-xyz/goapi/base/ctx"
)

var (
	ErrNotFound = errors.New("Cache not found")
)

// raw cache implementation
type Provider interface {
	Get(c ctx.Ctx, key string) ([]byte, time.Duration, error)
	Set(c ctx.Ctx, key string, value []byte, ttl time.Duration) error
	Incr(c ctx.Ctx, key string, val int) (int64, time.Duration, error)
	Del(c ctx.Ctx, key string) error
}
