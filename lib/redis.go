package lib

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	gorediscache "github.com/go-redis/cache/v8"
)

type RedisClientProvider interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
}

type RedisCacheProvider interface {
  Once(item *gorediscache.Item) error
}
