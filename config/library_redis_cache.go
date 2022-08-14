package config

type LibraryRedisCache struct {
	MaxNumber int
}

func CreateRedisCache(maxNumber int) LibraryRedisCache {
	return LibraryRedisCache{maxNumber}
}

func (library *LibraryRedisCache) Write(key string, value []byte) error {
	RedisClient.LPush(key, value)
	RedisClient.LTrim(key, 0, int64(library.MaxNumber-1))
	return nil
}

func (library *LibraryRedisCache) Read(key string) ([]string, error) {
	return RedisClient.LRange(key, 0, int64(library.MaxNumber-1)).Result()
}
