package memcache

type Message struct {
	Pk        int
	Timestamp int    `json:"ts"`
	Content   string `json:"msg"`
}
