package cache

import (
	"gopkg.in/redis.v5"
	"os"
	"sync"
)

type RedisRequestCacher struct {
	MaxNumber   int
	RedisClient *redis.Client
}

var (
	Once              sync.Once
	RedisRequestCache *RedisRequestCacher
)

func NewRedisCache(maxNumber int) *RedisRequestCacher {
	Once.Do(func() {
		redisUrl := os.Getenv("REDIS_URL")

		redisClient := redis.NewClient(&redis.Options{
			Addr: redisUrl,
		})

		RedisRequestCache = &RedisRequestCacher{maxNumber, redisClient}
	})

	return RedisRequestCache
}

func (r *RedisRequestCacher) Write(key string, value []byte) error {
	pushCmd := NewRedisCache(r.MaxNumber).RedisClient.LPush(key, value)

	if pushCmd.Err() != nil {
		return pushCmd.Err()
	}

	trimCmd := NewRedisCache(r.MaxNumber).RedisClient.LTrim(key, 0, int64(r.MaxNumber-1))

	if trimCmd.Err() != nil {
		return trimCmd.Err()
	}

	return nil
}

func (r *RedisRequestCacher) Read(key string) ([]string, error) {
	return NewRedisCache(r.MaxNumber).RedisClient.LRange(key, 0, int64(r.MaxNumber-1)).Result()
}
