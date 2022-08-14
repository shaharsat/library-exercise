package models

import "gin/config"

type Cache interface {
	Write(key string, value []byte) bool
	Read(n int) []string
}

type LibraryCache struct {
	MaxNumber int
}

func CreateRedisCache(maxNumber int) LibraryCache {
	return LibraryCache{maxNumber}
}

func (library *LibraryCache) Write(key string, value []byte) error {
	config.RedisClient.LPush(key, value)
	config.RedisClient.LTrim(key, 0, int64(library.MaxNumber-1))
	return nil
}

func (library *LibraryCache) Read(key string) ([]string, error) {
	return config.RedisClient.LRange(key, 0, int64(library.MaxNumber-1)).Result()
}
