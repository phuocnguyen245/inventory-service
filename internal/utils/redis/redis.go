package redisUtils

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

func AcquireLock(ctx context.Context, client *redis.Client, lockKey string, expiration time.Duration) (bool, error) {
	return client.SetNX(ctx, lockKey, "locked", expiration).Result()
}

func ReleaseLock(ctx context.Context, client *redis.Client, lockKey string) error {
	return client.Del(ctx, lockKey).Err()
}

func InvalidateCache(ctx context.Context, client *redis.Client, key string) error {
	return client.Del(ctx, key).Err()
}
