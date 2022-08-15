package cache

import "gin/config"

type LibraryRedisCache struct {
	MaxNumber int
}

func CreateRedisCache(maxNumber int) LibraryRedisCache {
	return LibraryRedisCache{maxNumber}
}

func (library *LibraryRedisCache) Write(key string, value []byte) error {
	pushCmd := config.RedisClient.LPush(key, value)

	if pushCmd.Err() != nil {
		return pushCmd.Err()
	}

	trimCmd := config.RedisClient.LTrim(key, 0, int64(library.MaxNumber-1))

	if trimCmd.Err() != nil {
		return trimCmd.Err()
	}

	return nil
}

func (library *LibraryRedisCache) Read(key string) ([]string, error) {
	return config.RedisClient.LRange(key, 0, int64(library.MaxNumber-1)).Result()
}
