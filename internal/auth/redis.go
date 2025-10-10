package database

import (
	"context"
	"os"
	"time"
	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client

func InitRedis(ctx context.Context) error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	Rdb = redis.NewClient(&redis.Options{
		Addr:        addr,
		DialTimeout: 5 * time.Second,
	})
	return Rdb.Ping(ctx).Err()
}

func CloseRedis() {
	if Rdb != nil {
		_ = Rdb.Close()
	}
}
