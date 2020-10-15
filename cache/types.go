package cache

// Cache (Redis)
type RDB interface{}

type Cache interface {
	RDB() RDB
}
