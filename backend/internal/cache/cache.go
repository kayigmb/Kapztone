// Package cache is a small Redis-backed cache. A cache failure is never an
// error for the caller: reads report a miss and writes are dropped, so the
// database stays the source of truth.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kayigmb/chaintrack/backend/internal/logx"
)

type Cache struct {
	rdb *redis.Client
	ttl time.Duration
}

// New connects lazily; use Ping to check the server is reachable.
func New(url string, ttl time.Duration) (*Cache, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	// Short timeouts: when Redis is slow or down, fall back to the database quickly.
	opt.DialTimeout = 500 * time.Millisecond
	opt.ReadTimeout = 250 * time.Millisecond
	opt.WriteTimeout = 250 * time.Millisecond
	return &Cache{rdb: redis.NewClient(opt), ttl: ttl}, nil
}

func (c *Cache) Ping(ctx context.Context) error { return c.rdb.Ping(ctx).Err() }
func (c *Cache) Close() error                   { return c.rdb.Close() }

// Get fills dst and reports true on a hit.
func (c *Cache) Get(ctx context.Context, key string, dst any) bool {
	b, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logx.From(ctx).Warn("cache get failed", "key", key, logx.KeyError, err)
		}
		return false
	}
	return json.Unmarshal(b, dst) == nil
}

func (c *Cache) Set(ctx context.Context, key string, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	if err := c.rdb.Set(ctx, key, b, c.ttl).Err(); err != nil {
		logx.From(ctx).Warn("cache set failed", "key", key, logx.KeyError, err)
	}
}

func (c *Cache) Del(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		logx.From(ctx).Warn("cache delete failed", logx.KeyError, err)
	}
}

// DelPattern deletes every key matching a glob pattern.
// SCAN walks the whole keyspace, fine while invalidations are rare
// (a role or organisation change); switch to tag sets if they become frequent.
func (c *Cache) DelPattern(ctx context.Context, pattern string) {
	iter := c.rdb.Scan(ctx, 0, pattern, 500).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		logx.From(ctx).Warn("cache scan failed", "pattern", pattern, logx.KeyError, err)
	}
	c.Del(ctx, keys...)
}

// Incr adds one to a counter that expires window after its first increment,
// and returns the new count. On a cache failure it returns 0.
func (c *Cache) Incr(ctx context.Context, key string, window time.Duration) int64 {
	n, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		logx.From(ctx).Warn("cache incr failed", "key", key, logx.KeyError, err)
		return 0
	}
	if n == 1 {
		c.rdb.Expire(ctx, key, window)
	}
	return n
}

// Count reads a counter without changing it; 0 if absent or on failure.
func (c *Cache) Count(ctx context.Context, key string) int64 {
	n, err := c.rdb.Get(ctx, key).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.From(ctx).Warn("cache count failed", "key", key, logx.KeyError, err)
	}
	return n
}
