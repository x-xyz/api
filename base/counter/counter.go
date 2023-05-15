package counter

import "sync"

type Counter struct {
	count int
	mu    sync.RWMutex
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Add(val int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count += val
}

func (c *Counter) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}
