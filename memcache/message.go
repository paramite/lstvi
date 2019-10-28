package memcache

import (
	"sort"
	"sync"
)

/************************* MessageCache implementation ***********************/

type MessageCache struct {
	Queue         chan Message
	lastPk        int
	indexTs       map[int][]int
	store         map[int]Message
	pkLock        sync.RWMutex
	storeLock     sync.RWMutex
	orderLock     sync.RWMutex
	indexLock     sync.RWMutex
	tsOrderedList []int
	tsNeedRefresh bool
}

func NewMessageCache(initialCapacity int) *MessageCache {
	var output MessageCache
	output.lastPk = 0
	output.indexTs = make(map[int][]int)
	output.store = make(map[int]Message)
	output.Queue = make(chan Message, initialCapacity)
	output.tsNeedRefresh = false
	output.tsOrderedList = make([]int, 0)
	return &output
}

func (self *MessageCache) Add(msg Message) {
	self.storeLock.Lock()
	self.store[msg.Pk] = msg
	self.storeLock.Unlock()
}

func (self *MessageCache) Index(msg Message) {
	self.indexLock.Lock()
	if _, ok := self.indexTs[msg.Timestamp]; !ok {
		self.indexTs[msg.Timestamp] = make([]int, 0)
		self.tsNeedRefresh = true
		self.tsOrderedList = append(self.tsOrderedList, msg.Timestamp)
	}
	self.indexTs[msg.Timestamp] = append(self.indexTs[msg.Timestamp], msg.Pk)
	self.indexLock.Unlock()
}

func (self *MessageCache) GetLast(count int) []Message {
	output := make([]Message, 0, count)

	if self.tsNeedRefresh {
		self.orderLock.Lock()
		sort.Sort(sort.Reverse(sort.IntSlice(self.tsOrderedList)))
		self.orderLock.Unlock()
		self.tsNeedRefresh = false
	}

	outLen := 0
	self.orderLock.RLock()
	for outLen < count {
		for _, ts := range self.tsOrderedList {
			resList := self.GetByTimestamp(ts)
			resLen := len(resList)
			if outLen+resLen <= count {
				output = append(output, resList...)
				outLen += resLen
			} else {
				for _, val := range resList[:count-outLen] {
					output = append(output, val)
				}
			}
		}
	}
	self.orderLock.RUnlock()
	return output
}

func (self *MessageCache) GetByTimestamp(ts int) []Message {
	var result []Message
	self.indexLock.RLock()
	if val, ok := self.indexTs[ts]; ok {
		self.indexLock.RUnlock()
		result = make([]Message, 0, len(val))
		for _, i := range val {
			self.storeLock.RLock()
			result = append(result, self.store[i])
			self.storeLock.RUnlock()
		}
	} else {
		result = make([]Message, 0)
	}
	return result
}

func (self *MessageCache) Process() {
	for msg := range self.Queue {
		self.pkLock.Lock()
		msg.Pk = self.lastPk
		self.lastPk++
		self.pkLock.Unlock()
		go self.Add(msg)
		go self.Index(msg)
	}
}
