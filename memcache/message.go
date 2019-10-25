package memcache

import (
	"sync"
)

/************************* MessageCache implementation ***********************/

type MessageCache struct {
	Queue     chan Message
	lastPk    int
	indexTs   map[int][]int
	store     map[int]Message
	storeLock sync.RWMutex
	indexLock sync.RWMutex
}

func NewMessageCache(initialCapacity int) *MessageCache {
	var output MessageCache
	output.lastPk = 0
	output.indexTs = make(map[int][]int)
	output.store = make(map[int]Message)
	output.Queue = make(chan Message, initialCapacity)
	return &output
}

func (self *MessageCache) Add(msg Message) {
	self.storeLock.Lock()
	msg.Pk = self.lastPk
	self.store[msg.Pk] = msg
	self.lastPk++
	self.storeLock.Unlock()
}

func (self *MessageCache) Index(msg Message) {
	self.indexLock.Lock()
	if _, ok := self.indexTs[msg.Timestamp]; ok {
		self.indexTs[msg.Timestamp] = make([]int, 0)
	}
	self.indexTs[msg.Timestamp] = append(self.indexTs[msg.Timestamp], self.lastPk)
	self.indexLock.Unlock()
}

func (self *MessageCache) GetLast(count int) []Message {
	index := self.lastPk - count
	if index < 0 {
		index = 0
	}
	output := make([]Message, 0, self.lastPk-index)
	self.storeLock.RLock()
	for idx := index; idx < self.lastPk; idx++ {
		output = append(output, self.store[idx])
	}
	self.storeLock.RUnlock()
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

func (self *MessageCache) Process() {
	for msg := range self.Queue {
		go self.Add(msg)
		go self.Index(msg)
	}
}
