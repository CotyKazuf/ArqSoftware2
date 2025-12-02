package cache

import (
	"log"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/karlseguin/ccache/v3"
)

// Cache abstracts a basic key-value cache with TTL support.
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
	Delete(key string)
	Flush() error
}

// CCacheLayer implements Cache backed by an in-memory CCache instance.
type CCacheLayer struct {
	cache *ccache.Cache[[]byte]
}

// NewCCacheLayer builds a CCache layer with a maximum number of entries.
func NewCCacheLayer(maxEntries int64) *CCacheLayer {
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	return &CCacheLayer{
		cache: ccache.New(ccache.Configure[[]byte]().MaxSize(maxEntries)),
	}
}

func (c *CCacheLayer) Get(key string) ([]byte, bool) {
	if c == nil || c.cache == nil {
		return nil, false
	}
	item := c.cache.Get(key)
	if item == nil || item.Expired() {
		return nil, false
	}
	return item.Value(), true
}

func (c *CCacheLayer) Set(key string, value []byte, ttl time.Duration) {
	if c == nil || c.cache == nil {
		return
	}
	c.cache.Set(key, value, ttl)
}

func (c *CCacheLayer) Delete(key string) {
	if c == nil || c.cache == nil {
		return
	}
	c.cache.Delete(key)
}

func (c *CCacheLayer) Flush() error {
	if c == nil || c.cache == nil {
		return nil
	}
	c.cache.Clear()
	return nil
}

// MemcachedLayer implements Cache using a Memcached backend.
type MemcachedLayer struct {
	client *memcache.Client
}

// NewMemcachedLayer builds a Memcached cache pointing at the given address.
func NewMemcachedLayer(addr string) *MemcachedLayer {
	if addr == "" {
		return nil
	}
	return &MemcachedLayer{client: memcache.New(addr)}
}

func (m *MemcachedLayer) Get(key string) ([]byte, bool) {
	if m == nil || m.client == nil {
		return nil, false
	}
	item, err := m.client.Get(key)
	if err != nil {
		if err != memcache.ErrCacheMiss {
			log.Printf("memcached get %s: %v", key, err)
		}
		return nil, false
	}
	return item.Value, true
}

func (m *MemcachedLayer) Set(key string, value []byte, ttl time.Duration) {
	if m == nil || m.client == nil {
		return
	}
	expiration := int32(ttl.Seconds())
	if ttl <= 0 {
		expiration = 0
	}
	if err := m.client.Set(&memcache.Item{Key: key, Value: value, Expiration: expiration}); err != nil {
		log.Printf("memcached set %s: %v", key, err)
	}
}

func (m *MemcachedLayer) Delete(key string) {
	if m == nil || m.client == nil {
		return
	}
	if err := m.client.Delete(key); err != nil && err != memcache.ErrCacheMiss {
		log.Printf("memcached delete %s: %v", key, err)
	}
}

func (m *MemcachedLayer) Flush() error {
	if m == nil || m.client == nil {
		return nil
	}
	if err := m.client.FlushAll(); err != nil {
		log.Printf("memcached flush: %v", err)
		return err
	}
	return nil
}

// LayeredCache combines a fast in-memory cache with an optional distributed cache.
type LayeredCache struct {
	memory      Cache
	distributed Cache
	warmTTL     time.Duration
}

// NewLayeredCache wires the cache layers together.
func NewLayeredCache(memory Cache, distributed Cache, warmTTL time.Duration) *LayeredCache {
	return &LayeredCache{
		memory:      memory,
		distributed: distributed,
		warmTTL:     warmTTL,
	}
}

func (c *LayeredCache) Get(key string) ([]byte, bool) {
	if c == nil {
		return nil, false
	}
	if c.memory != nil {
		if val, ok := c.memory.Get(key); ok {
			return val, true
		}
	}

	if c.distributed != nil {
		if val, ok := c.distributed.Get(key); ok {
			if c.memory != nil && c.warmTTL > 0 {
				c.memory.Set(key, val, c.warmTTL)
			}
			return val, true
		}
	}
	return nil, false
}

func (c *LayeredCache) Set(key string, value []byte, ttl time.Duration) {
	if c == nil {
		return
	}
	if c.distributed != nil {
		c.distributed.Set(key, value, ttl)
	}
	if c.memory != nil {
		effectiveTTL := ttl
		if effectiveTTL <= 0 {
			effectiveTTL = c.warmTTL
		}
		c.memory.Set(key, value, effectiveTTL)
	}
}

func (c *LayeredCache) Delete(key string) {
	if c == nil {
		return
	}
	if c.memory != nil {
		c.memory.Delete(key)
	}
	if c.distributed != nil {
		c.distributed.Delete(key)
	}
}

func (c *LayeredCache) Flush() error {
	if c == nil {
		return nil
	}
	if c.memory != nil {
		_ = c.memory.Flush()
	}
	if c.distributed != nil {
		return c.distributed.Flush()
	}
	return nil
}
