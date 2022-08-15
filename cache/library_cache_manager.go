package cache

type LibraryCacheManager interface {
	Write(key string, value []byte) bool
	Read(n int) []string
}
