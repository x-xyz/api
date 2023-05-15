package ctx

import (
	"context"
	"time"

	log "github.com/x-xyz/goapi/base/log"
)

type Ctx struct {
	context.Context
	log.Logger
}

func Background() Ctx {
	return Ctx{
		Context: context.Background(),
		Logger:  log.Log(),
	}
}

func Todo() Ctx {
	return Ctx{
		Context: context.TODO(),
		Logger:  log.Log(),
	}
}

func WithValue(parent Ctx, key string, val interface{}) Ctx {
	return Ctx{
		Context: context.WithValue(parent, key, val),
		Logger:  parent.Logger.WithField(key, val),
	}
}

func WithValues(parent Ctx, kvs map[string]interface{}) Ctx {
	c := parent
	for k, v := range kvs {
		c = WithValue(c, k, v)
	}
	return c
}

func WithCancel(parent Ctx) (Ctx, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	return Ctx{
		Context: ctx,
		Logger:  parent.Logger,
	}, cancel
}

func WithTimeout(parent Ctx, timeout time.Duration) (Ctx, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	return Ctx{
		Context: ctx,
		Logger:  parent.Logger,
	}, cancel
}
