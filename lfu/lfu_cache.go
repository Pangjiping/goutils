package lfu

import (
	"container/list"
	"sync"
)

type node struct {
	key   string
	value interface{}
	freq  int
}

type LFUCache struct {
	kv      map[string]*list.Element
	fk      map[int]*list.List
	minFreq int
	cap     int
	mu      sync.Mutex
}

func NewLFUCache(cap int) *LFUCache {
	if cap < 0 || cap > 1e4 {
		return nil
	}

	lfu := &LFUCache{
		cap:     cap,
		kv:      make(map[string]*list.Element),
		fk:      make(map[int]*list.List),
		minFreq: 0,
		mu:      sync.Mutex{},
	}
	return lfu
}

func (l *LFUCache) Get(key string) (interface{}, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var elem *list.Element
	var ok bool
	if elem, ok = l.kv[key]; !ok {
		return nil, false
	}

	node := elem.Value.(node)
	oldFreq := node.freq
	node.freq++

	okList := l.fk[oldFreq]
	okList.Remove(elem)

	var exist bool
	var keyList *list.List

	if keyList, exist = l.fk[node.freq]; !exist {
		keyList = list.New()
	}

	l.kv[key] = keyList.PushFront(node)
	l.fk[node.freq] = keyList
	if okList.Len() == 0 {
		if l.minFreq == oldFreq {
			l.minFreq++
		}
	}
	return node.value, true
}

func (l *LFUCache) Put(key string, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.cap == 0 {
		return
	}

	if elem, exist := l.kv[key]; exist {
		node := elem.Value.(node)
		node.value = value
		oldFreq := node.freq
		node.freq++

		kList := l.fk[oldFreq]
		kList.Remove(elem)

		var t bool
		var nkList *list.List
		if nkList, t = l.fk[oldFreq]; !t {
			nkList = list.New()
			l.fk[oldFreq+1] = nkList
		}
		l.kv[key] = nkList.PushFront(node)

		if kList.Len() == 0 {
			if oldFreq == l.minFreq {
				l.minFreq++
			}
		}
		return
	}

	if l.cap == len(l.kv) {
		minList := l.fk[l.minFreq]

		b := minList.Back()
		minList.Remove(b)
		node := b.Value.(node)

		delete(l.kv, node.key)
	}

	nN := node{
		key:   key,
		value: value,
		freq:  1,
	}
	l.minFreq = nN.freq

	var keyList *list.List
	var exist bool
	if keyList, exist = l.fk[l.minFreq]; !exist {
		keyList = list.New()
	}
	l.kv[key] = keyList.PushFront(nN)
	l.fk[l.minFreq] = keyList
}
