package memcache

type Cache interface {
	Add(Message)
	GetLast(int) []interface{}
}
