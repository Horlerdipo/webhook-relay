package datastore

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type Store interface {
	Name() string
	Ping(ctx context.Context) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
}

type RedisStore struct {
	client *redis.Client
}

func (rs *RedisStore) Ping(ctx context.Context) error {
	err := rs.client.Ping(ctx).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rs *RedisStore) Get(ctx context.Context, key string) (string, error) {
	val, err := rs.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (rs *RedisStore) Set(ctx context.Context, key string, value any, ttl time.Duration) (string, error) {
	val, err := rs.client.Set(ctx, key, value, ttl).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (rs *RedisStore) Name() string {
	return "redis"
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
	}
}
