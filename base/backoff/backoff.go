package backoff

import (
	"context"
	"math"
	"time"
)

type BackoffStrategy interface {
	GetBackoffDuration(int, time.Duration, time.Duration) time.Duration
}

type Backoff struct {
	LastDuration time.Duration
	NextDuration time.Duration
	start        time.Duration
	limit        time.Duration
	count        int
	strategy     BackoffStrategy
}

func NewBackoff(strategy BackoffStrategy, start time.Duration, limit time.Duration) *Backoff {
	backoff := Backoff{strategy: strategy, start: start, limit: limit}
	backoff.Reset()
	return &backoff
}

func (b *Backoff) Reset() {
	b.count = 0
	b.LastDuration = 0
	b.NextDuration = b.getNextDuration()
}

func (b *Backoff) Backoff(ctx context.Context) (err error) {
	sleepCtx, cancelSleep := context.WithTimeout(ctx, b.NextDuration)
	<-sleepCtx.Done()
	cancelSleep()
	if sleepCtx.Err() == context.DeadlineExceeded {
		b.count++
		b.LastDuration = b.NextDuration
		b.NextDuration = b.getNextDuration()
		return nil
	}
	return sleepCtx.Err()
}

func (b *Backoff) getNextDuration() time.Duration {
	backoff := b.strategy.GetBackoffDuration(b.count, b.start, b.LastDuration)
	if b.limit > 0 && backoff > b.limit {
		backoff = b.limit
	}
	return backoff
}

type exponential struct{}

func (exponential) GetBackoffDuration(backoffCount int, start time.Duration, lastBackoff time.Duration) time.Duration {
	period := int64(math.Pow(2, float64(backoffCount)))
	return time.Duration(period) * start
}

func NewExponential(start time.Duration, limit time.Duration) *Backoff {
	return NewBackoff(exponential{}, start, limit)
}

type linear struct{}

func (linear) GetBackoffDuration(backoffCount int, start time.Duration, lastBackoff time.Duration) time.Duration {
	return time.Duration(backoffCount) * start
}

func NewLinear(start time.Duration, limit time.Duration) *Backoff {
	return NewBackoff(linear{}, start, limit)
}
