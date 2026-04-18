package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/sskotezhov/maturin/config"
)

func Connect(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return rdb, nil
}
