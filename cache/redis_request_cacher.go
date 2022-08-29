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
	RedisRequestCacherOnce sync.Once
	RedisRequestCache      *RedisRequestCacher
)

func NewRedisCache(maxNumber int) *RedisRequestCacher {
	RedisRequestCacherOnce.Do(func() {
		redisUrl := os.Getenv("REDIS_URL")

		redisClient := redis.NewClient(&redis.Options{
			Addr: redisUrl,
		})

		RedisRequestCache = &RedisRequestCacher{maxNumber, redisClient}
	})

	return RedisRequestCache
}

func (library *RedisRequestCacher) Write(key string, value []byte) error {
	pushCmd := NewRedisCache(library.MaxNumber).RedisClient.LPush(key, value)

	if pushCmd.Err() != nil {
		return pushCmd.Err()
	}

	trimCmd := NewRedisCache(library.MaxNumber).RedisClient.LTrim(key, 0, int64(library.MaxNumber-1))

	if trimCmd.Err() != nil {
		return trimCmd.Err()
	}

	return nil
}

func (library *RedisRequestCacher) Read(key string) ([]string, error) {
	return NewRedisCache(library.MaxNumber).RedisClient.LRange(key, 0, int64(library.MaxNumber-1)).Result()
}
