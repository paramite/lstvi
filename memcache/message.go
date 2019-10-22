package memcache

import (
	"sync"
)

type MessageCache struct {
	lastPk  int
	indexTs map[int][]int
	store   []Message
	lock    sync.RWMutex
}

func NewMessageCache(initialCapacity int) *MessageCache {
	var output MessageCache
	output.lastPk = 0
	output.indexTs = make(map[int][]int)
	output.store = make([]Message, 0, initialCapacity)
	return &output
}

func (self *MessageCache) Add(msg Message) {
	self.lock.Lock()
	if _, ok := self.indexTs[msg.Timestamp]; ok {
		self.indexTs[msg.Timestamp] = make([]int, 0)
	}
	self.indexTs[msg.Timestamp] = append(self.indexTs[msg.Timestamp], self.lastPk)
	msg.Pk = self.lastPk
	self.store = append(self.store, msg)
	self.lastPk++
	self.lock.Unlock()
}

func (self *MessageCache) GetLast(count int) []interface{} {
	index := len(self.store) - count
	if index < 0 {
		index = 0
	}
	output := make([]interface{}, len(self.store)-index)
	for idx, val := range self.store[index:] {
		output[idx] = val
	}
	return output
}

func (self *MessageCache) GetByTimestamp(ts int) []Message {
	var result []Message
	if val, ok := self.indexTs[ts]; ok {
		result = make([]Message, 0, len(val))
		for i := range val {
			result = append(result, self.store[i])
		}
	}
	return result
}
