package cache

import "gin/config"

type RedisRequestCacher struct {
	MaxNumber int
}

func CreateRedisCache(maxNumber int) RedisRequestCacher {
	return RedisRequestCacher{maxNumber}
}

func (library *RedisRequestCacher) Write(key string, value []byte) error {
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

func (library *RedisRequestCacher) Read(key string) ([]string, error) {
	return config.RedisClient.LRange(key, 0, int64(library.MaxNumber-1)).Result()
}
