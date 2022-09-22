package cache

type RequestCacher interface {
	Write(key string, value []byte) bool
	Read(n int) []string
}
