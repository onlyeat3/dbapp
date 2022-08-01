package virtdb

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type GenericRedisClient struct {
	redisClient        *redis.Client
	redisClusterClient *redis.ClusterClient
}

func (h GenericRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	if h.redisClient != nil {
		return h.redisClient.Get(ctx, key)
	} else {
		return h.redisClusterClient.Get(ctx, key)
	}
}

func (h GenericRedisClient) Set(ctx context.Context, query string, encoded []byte, duration time.Duration) *redis.StatusCmd {
	if h.redisClient != nil {
		return h.redisClient.Set(ctx, query, encoded, duration)
	} else {
		return h.redisClusterClient.Set(ctx, query, encoded, duration)
	}
}

func NewGenericRedisClient(config *VirtdbConfig) *GenericRedisClient {
	if len(config.RedisAddresses) > 0 {
		options := &redis.ClusterOptions{
			Addrs:    config.RedisAddresses,
			PoolSize: config.RedisPoolSize,
			Password: config.RedisPassword,
		}
		redisClusterClient := redis.NewClusterClient(options)
		return &GenericRedisClient{
			redisClient:        nil,
			redisClusterClient: redisClusterClient,
		}
	} else {
		options := &redis.Options{Addr: config.RedisAddress, PoolSize: config.RedisPoolSize, Password: config.RedisPassword}
		redisClient := redis.NewClient(options)
		return &GenericRedisClient{
			redisClient:        redisClient,
			redisClusterClient: nil,
		}
	}
}

func NewGenericRedisClientWithConfig(address string, poolSize int, password string) *GenericRedisClient {
	config := &VirtdbConfig{
		RedisAddress:  address,
		RedisPoolSize: poolSize,
		RedisPassword: password,
	}
	return NewGenericRedisClient(config)
}
