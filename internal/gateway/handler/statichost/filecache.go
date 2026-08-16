package statichost

import (
	"sync"
	"time"
)

const (
	defaultFileCacheTTL     = 5 * time.Second
	defaultFileCacheMax     = 1024
	defaultFileCacheBodyMax = 64 << 10
)

// fileCacheEntry 打开文件缓存项，对应 nginx open_file_cache 的 stat / 小文件内容。
type fileCacheEntry struct {
	deadline        time.Time
	missing         bool
	isDir           bool
	size            int64
	modTime         time.Time
	body            []byte
	realPath        string
	contentEncoding string
}

// fileCache 进程内短 TTL 文件缓存，降低重复 Lstat/Open。
// 热更新通过 TTL 过期可见；不跨代际共享文件描述符。
type fileCache struct {
	mu      sync.Mutex
	items   map[string]*fileCacheEntry
	ttl     time.Duration
	max     int
	bodyMax int64
}

func newFileCache(ttl time.Duration, max int, bodyMax int64) *fileCache {
	if max <= 0 {
		max = defaultFileCacheMax
	}
	if bodyMax <= 0 {
		bodyMax = defaultFileCacheBodyMax
	}
	return &fileCache{
		items:   make(map[string]*fileCacheEntry, max),
		ttl:     ttl,
		max:     max,
		bodyMax: bodyMax,
	}
}

func defaultFileCache() *fileCache {
	return newFileCache(defaultFileCacheTTL, defaultFileCacheMax, defaultFileCacheBodyMax)
}

func (c *fileCache) enabled() bool {
	return c != nil && c.ttl > 0
}

func (c *fileCache) get(key string) *fileCacheEntry {
	if !c.enabled() || key == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.items[key]
	if !ok {
		return nil
	}
	if time.Now().After(entry.deadline) {
		delete(c.items, key)
		return nil
	}
	return entry
}

func (c *fileCache) put(key string, entry *fileCacheEntry) {
	if !c.enabled() || key == "" || entry == nil {
		return
	}
	entry.deadline = time.Now().Add(c.ttl)
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) >= c.max {
		c.evictExpiredLocked()
		if len(c.items) >= c.max {
			c.evictHalfLocked()
		}
	}
	c.items[key] = entry
}

func (c *fileCache) evictExpiredLocked() {
	now := time.Now()
	for key, entry := range c.items {
		if now.After(entry.deadline) {
			delete(c.items, key)
		}
	}
}

func (c *fileCache) evictHalfLocked() {
	limit := len(c.items) / 2
	removed := 0
	for key := range c.items {
		delete(c.items, key)
		removed++
		if removed >= limit {
			return
		}
	}
}

func (c *fileCache) close() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.items = make(map[string]*fileCacheEntry)
	c.mu.Unlock()
}
