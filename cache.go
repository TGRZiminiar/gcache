package gcache

import (
	"runtime"
	"sync"
	"time"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type Item struct {
	Object     interface{}
	Expiration int64
}

type Keeper struct {
	stopCh         chan struct{}
	intervalExpire time.Duration
}

type GCache struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
	Keeper            *Keeper
}

func NewGCache(de time.Duration) *GCache {
	return &GCache{
		defaultExpiration: de,
		items:             make(map[string]Item),
	}
}

func NewCache(defaultExpiration time.Duration, clearInterval time.Duration) *GCache {
	c := NewGCache(defaultExpiration)
	if clearInterval > 0 {
		c.runKeeper(defaultExpiration)
		// When k is no longer reachable and is about to be collected by the garbage collector,
		// stopKeeper will be called with k as its argument,
		// allowing the program to clean up resources or perform any necessary finalization.
		runtime.SetFinalizer(c, stopKeeper)
	}
	return c
}

// Set store the key and value into a cache expiration time must be provided
func (c *GCache) Set(k string, v interface{}, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.mu.Lock()
	c.items[k] = Item{
		Object:     v,
		Expiration: e,
	}
	defer c.mu.Unlock()
}

// SetDefault store the key and value into a cache using default expiration
func (c *GCache) SetDefault(k string, v interface{}) {
	c.Set(k, v, DefaultExpiration)
}

// Get use the key to recieve a value from a cache it will return a value and boolean
func (c *GCache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[k]
	if !ok {
		return nil, ok
	}
	if v.Expiration > 0 {
		if time.Now().UnixNano() > v.Expiration {
			return nil, false
		}
	}
	return v.Object, ok
}

// Get use the key to recieve a value from a cache it will return a value, expiration time and boolean
func (c *GCache) GetWithExpiration(k string) (interface{}, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[k]
	if !ok {
		return nil, time.Time{}, ok
	}
	if v.Expiration > 0 {
		if time.Now().UnixNano() > v.Expiration {
			return nil, time.Time{}, false
		}

		return v.Object, time.Unix(0, v.Expiration), ok
	}

	// case of no expiration time (-1)
	return v.Object, time.Time{}, ok
}

// runKeeper is to call a run function that will trigger the cache clear interval expiration
func (c *GCache) runKeeper(e time.Duration) {
	keeper := &Keeper{
		stopCh:         make(chan struct{}),
		intervalExpire: e,
	}
	c.Keeper = keeper
	go c.Keeper.run(c)
}

// run used to call a function to delete a expired cache when the expiration time has come
// and will stop the routine when the stopCh got trigger
func (k *Keeper) run(c *GCache) {
	ticker := time.NewTicker(k.intervalExpire)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-k.stopCh:
			ticker.Stop()
		}
	}
}

// for stoping keeper from watching
func stopKeeper(c *GCache) {
	c.Keeper.stopCh <- struct{}{}
}

// Delete all expired items from the cache.
func (c *GCache) DeleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			c.delete(k)
		}
	}
}

// delete use with a function of DeleteExpired since the mutex already lock on the parent function
// so we don't have to lock in here
func (c *GCache) delete(k string) {
	_, ok := c.items[k]
	if ok {
		delete(c.items, k)
	}
}

// Delete remove a cache  using a key
func (c *GCache) Delete(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.items[k]
	if ok {
		delete(c.items, k)
	}
}

// Items return all the items in the cache
func (c *GCache) Items() map[string]Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now().UnixNano()
	m := make(map[string]Item, len(c.items))
	for k, v := range c.items {
		if v.Expiration <= 0 || now <= v.Expiration {
			m[k] = v
		}
	}

	return m
}
