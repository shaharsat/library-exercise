package config

type LibraryCache interface {
	Write(key string, value []byte) bool
	Read(n int) []string
}
