package main

import (
	"net"
	"sync"
)

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus // vector clock
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

type Gossiper struct {
	address *net.UDPAddr
	conn    *net.UDPConn
	Name    string
}

type Message struct {
	Type, Text string
}

type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

type SearchResult struct {
	FileName     string
	ChunkCount   uint64
	MetafileHash []byte
	ChunkMap     []uint64
}

// The packet eventually transmitted between nodes
type GossipPacket struct {
	Simple         *SimpleMessage
	Rumor          *RumorMessage
	Status         *StatusPacket
	Private        *PrivateMessage
	DataRequest    *DataRequest
	DataReply      *DataReply
	SearchRequest  *SearchRequest
	SearchReply    *SearchReply
	TxPublish      *TxPublish
	BlockPublish   *BlockPublish
	WeightPacket   *WeightPacket
	GradientPacket *GradientPacket
}

type ConcurrentSlice struct {
	sync.RWMutex
	items []interface{}
}

type ConcurrentSliceString struct {
	sync.RWMutex
	items []string
}

func (cs *ConcurrentSlice) Append(item interface{}) {
	cs.Lock()
	defer cs.Unlock()

	cs.items = append(cs.items, item)
}

func (cs *ConcurrentSlice) Exist(item interface{}) bool {
	cs.Lock()
	defer cs.Unlock()

	for val := range cs.Iter() {
		if val == item {
			return true
		}
	}

	return false
}

func (cs *ConcurrentSlice) Iter() <-chan interface{} {
	c := make(chan interface{})

	f := func() {
		cs.Lock()
		defer cs.Unlock()
		for _, value := range cs.items {
			c <- value
		}
		close(c)
	}

	go f()

	return c
}

func (cs *ConcurrentSlice) Get(index int) interface{} {
	cs.Lock()
	defer cs.Unlock()

	return cs.items[index]
}

func (cs *ConcurrentSlice) GetSlice() []interface{} {
	cs.Lock()
	defer cs.Unlock()

	return cs.items
}

func (cs *ConcurrentSlice) Len() int {
	cs.Lock()
	defer cs.Unlock()

	return len(cs.items)
}

// func (cs *ConcurrentSlice) DeleteAll() {
// 	cs.Lock()
// 	defer cs.Unlock()

// 	return len(cs.items)
// }

func (cs *ConcurrentSliceString) Append(item string) {
	cs.Lock()
	defer cs.Unlock()

	cs.items = append(cs.items, item)
}

func (cs *ConcurrentSliceString) Iter() <-chan string {
	c := make(chan string)

	f := func() {
		cs.Lock()
		defer cs.Unlock()
		for _, value := range cs.items {
			c <- value
		}
		close(c)
	}

	go f()

	return c
}

func (cs *ConcurrentSliceString) Get(index int) string {
	cs.Lock()
	defer cs.Unlock()

	return cs.items[index]
}

func (cs *ConcurrentSliceString) GetSlice() []string {
	cs.Lock()
	defer cs.Unlock()

	return cs.items
}

func (cs *ConcurrentSliceString) Len() int {
	cs.Lock()
	defer cs.Unlock()

	return len(cs.items)
}

// type ConcurrentMap struct {
// 	sync.RWMutex
// 	items map[string]bool
// }

// func (cm *ConcurrentMap) Set(key string) {
// 	cm.Lock()
// 	defer cm.Unlock()

// 	cm.items[key] = true
// }

// func (cm *ConcurrentMap) Exist(key string) bool {
// 	cm.Lock()
// 	defer cm.Unlock()

// 	_, ok := cm.items[key]

// 	return ok
// }

// func (cm *ConcurrentMap) Delete(key string) {
// 	cm.Lock()
// 	defer cm.Unlock()

// 	delete(cm.items, key)
// }

// type ConcurrentMap struct {
// 	sync.RWMutex
// 	items map[string]interface{}
// }

// func (cm *ConcurrentMap) Set(key string, value interface{}) {
// 	cm.Lock()
// 	defer cm.Unlock()

// 	cm.items[key] = value
// }

// func (cm *ConcurrentMap) Exist(key string) bool {
// 	cm.Lock()
// 	defer cm.Unlock()

// 	_, ok := cm.items[key]

// 	return ok
// }

// func (cm *ConcurrentMap) Delete(key string) {
// 	cm.Lock()
// 	defer cm.Unlock()

// 	delete(cm.items, key)
// }

type ConcurrentMap struct {
	sync.RWMutex
	items map[interface{}]interface{}
}

func (cm *ConcurrentMap) Set(key interface{}, value interface{}) {
	cm.Lock()
	defer cm.Unlock()

	cm.items[key] = value
}

func (cm *ConcurrentMap) Exist(key interface{}) bool {
	cm.Lock()
	defer cm.Unlock()

	_, ok := cm.items[key]

	return ok
}

// func (cm *ConcurrentMap) Get(key interface{}) map[interface{}]interface{} {
func (cm *ConcurrentMap) Get(key interface{}) interface{} {

	cm.Lock()
	defer cm.Unlock()

	data, _ := cm.items[key]

	return data
}

func (cm *ConcurrentMap) Len() int {
	cm.Lock()
	defer cm.Unlock()

	return len(cm.items)
}

func (cm *ConcurrentMap) Delete(key interface{}) {
	cm.Lock()
	defer cm.Unlock()

	delete(cm.items, key)
}

func (cm *ConcurrentMap) DeleteAll() {
	cm.Lock()
	defer cm.Unlock()

	for key := range cm.IterKey() {
		delete(cm.items, key)
	}
}

func (cm *ConcurrentMap) Iter() <-chan interface{} {
	c := make(chan interface{})

	f := func() {
		cm.Lock()
		defer cm.Unlock()
		for _, value := range cm.items {
			c <- value
		}
		close(c)
	}

	go f()

	return c
}

func (cm *ConcurrentMap) IterKey() <-chan interface{} {
	c := make(chan interface{})

	f := func() {
		cm.Lock()
		defer cm.Unlock()
		for key, _ := range cm.items {
			c <- key
		}
		close(c)
	}

	go f()

	return c
}
