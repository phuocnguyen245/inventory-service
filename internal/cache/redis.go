package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func InitRedis(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
