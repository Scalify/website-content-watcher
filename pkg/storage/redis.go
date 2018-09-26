package storage

import (
	"fmt"

	"github.com/go-redis/redis"
)

// RedisStorage is a KV store
type RedisStorage struct {
	client redisClient
}

// NewRedis returns a new RedisStorage
func NewRedis(client redisClient) *RedisStorage {
	return &RedisStorage{
		client: client,
	}
}

// Set set's a key to a given value
func (r *RedisStorage) Set(key, value string) error {
	if err := r.client.Set(key, value, 0).Err(); err != nil {
		return fmt.Errorf("failed to set value: %v", err)
	}

	return nil
}

// Get fetches a value from the store
func (r *RedisStorage) Get(key string) (string, error) {
	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to fetch value: %v", err)
	}

	return val, nil
}

//Del deletes a key from the store
func (r *RedisStorage) Del(key string) error {
	if err := r.client.Del(key).Err(); err != nil {
		return fmt.Errorf("failed to delete key: %v", err)
	}

	return nil
}
