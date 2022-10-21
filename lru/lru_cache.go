package lru

import "sync"

const (
	// The default head node's key of LRUCache, it is forbidden to be used.
	lru_head = "__LRU_SAVED_HEAD__"

	// The default tail node's key of LRUCache, it is forbidden to be used.
	lru_tail = "__LRU_SAVED_TAIL__"
)

// LRUCache provides a thread-safe mem-cache of the lru mechanism.
type LRUCache struct {
	// The maximum capacity of the lru cache.
	// This variable is specified when the user initializes,
	// and the lru cache can be expanded through the expansion function.
	// If the current size exceeds this capacity,
	// the oldest unused key will be retired.
	capacity int

	// Record the current size of cached items.
	size int

	//
	cache map[string]*dLinkedNode
	mu    sync.Mutex
	head  *dLinkedNode
	tail  *dLinkedNode
}

type dLinkedNode struct {
	key   string
	value interface{}
	prev  *dLinkedNode
	next  *dLinkedNode
}

func initDLinkedNode(key string, value interface{}) *dLinkedNode {
	return &dLinkedNode{
		key:   key,
		value: value,
	}
}

func NewLRUCache(capacity int) *LRUCache {
	lru := &LRUCache{
		mu:       sync.Mutex{},
		size:     0,
		cache:    map[string]*dLinkedNode{},
		head:     initDLinkedNode(lru_head, nil),
		tail:     initDLinkedNode(lru_tail, nil),
		capacity: capacity,
	}
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	return lru
}

func (l *LRUCache) Get(key string) (interface{}, bool) {
	if key == lru_head || key == lru_tail {
		return nil, false
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.cache[key]; !ok {
		return nil, false
	}

	node := l.cache[key]
	l.moveToHead(node)
	return node.value, true
}

func (l *LRUCache) Put(key string, value interface{}) {
	if key == lru_head || key == lru_tail {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.cache[key]; !ok {
		node := initDLinkedNode(key, value)
		l.cache[key] = node
		l.addToHead(node)
		l.size++

		if l.size > l.capacity {
			removed := l.removeTail()
			delete(l.cache, removed.key)
			l.size--
		}
	} else {
		node := l.cache[key]
		node.value = value
		l.moveToHead(node)
	}
}

func (l *LRUCache) Expansion(cap int) {
	if cap < l.capacity {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.capacity = cap
}

func (l *LRUCache) addToHead(node *dLinkedNode) {
	node.prev = l.head
	node.next = l.head.next
	l.head.next.prev = node
	l.head.next = node
}

func (l *LRUCache) removeNode(node *dLinkedNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (l *LRUCache) moveToHead(node *dLinkedNode) {
	l.removeNode(node)
	l.addToHead(node)
}

func (l *LRUCache) removeTail() *dLinkedNode {
	node := l.tail.prev
	l.removeNode(node)
	return node
}
