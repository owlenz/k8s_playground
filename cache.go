package main

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type usersCache struct {
	mu      sync.RWMutex
	data    []byte
	expires time.Time
	group   singleflight.Group
}

func (c *usersCache) get() ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.data != nil && time.Now().Before(c.expires) {
		return c.data, true
	}
	return nil, false
}

func (c *usersCache) set(b []byte, ttl time.Duration) {
	c.mu.Lock()
	c.data, c.expires = b, time.Now().Add(ttl)
	c.mu.Unlock()
}

func (c *usersCache) invalidate() {
	c.mu.Lock()
	c.data = nil
	c.mu.Unlock()
}
