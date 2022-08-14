package config

import (
	"gopkg.in/redis.v5"
	"os"
)

var RedisClient *redis.Client

func SetupRedis() {
	redisUrl := os.Getenv("REDIS_URL")

	RedisClient = redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})
}
