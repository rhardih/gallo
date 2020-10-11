package lib

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type CacheProvider interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
}

// Simple decorator to hide the *redis.StringCmd / *redis.StatusCmd return
// values, allowing callers to omit the full redis import
type RedisClientDecorator struct {
	Client *redis.Client
}

func (c RedisClientDecorator) Get(key string) (string, error) {
	val, err := c.Client.Get(key).Result()
	if err == nil {
		return val, nil
	}

	return "", errors.New(fmt.Sprintf("Key not found in cache: %q", key))
}

func (c RedisClientDecorator) Set(key, value string, expiration time.Duration) error {
	err := c.Client.Set(key, value, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}
