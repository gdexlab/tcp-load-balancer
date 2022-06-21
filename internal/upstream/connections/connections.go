package connections

import "sync"

// Counter tracks the number of active connections to the host in a concurrency=safe manner.
type Counter struct {
	count int
	sync.Mutex
}

func (c *Counter) Increment() {
	c.Lock()
	c.count++
	c.Unlock()
}

func (c *Counter) Decrement() {
	c.Lock()
	c.count--
	c.Unlock()
}

func (c *Counter) Count() int {
	c.Lock()
	defer c.Unlock()
	return c.count
}
