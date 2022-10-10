package zmap

import (
	"bytes"
	"encoding/json"
	"sync"
)

type ZMap struct {
	kv map[interface{}]*Element
	ll list
	mu sync.RWMutex
}

func NewZMap() *ZMap {
	return &ZMap{
		kv: make(map[interface{}]*Element),
		mu: sync.RWMutex{},
	}
}

// Get returns the value for a key. If the key does not exist, the second return
// parameter will be false and the value will be nil.
func (m *ZMap) Get(key interface{}) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	element, ok := m.kv[key]
	if ok {
		return element.Value, true
	}
	return nil, false
}

// Set will set (or replace) a value for a key. If the key was new, then true
// will be returned. The returned value will be false if the value was replaced
// (even if the value was the same).
func (m *ZMap) Set(key, value interface{}) bool {
	m.mu.RLock()
	_, alreadyExist := m.kv[key]
	if alreadyExist {
		m.kv[key].Value = value
		return false
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	element := m.ll.PushBack(key, value)
	m.kv[key] = element
	return true
}

// GetOrDefault returns the value for a key. If the key does not exist, returns
// the default value instead.
func (m *ZMap) GetOrDefault(key, defaultValue interface{}) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if element, ok := m.kv[key]; ok {
		return element.Value
	}
	return defaultValue
}

// GetElement returns the element for a key. If the key does not exist, the
// pointer will be nil.
func (m *ZMap) GetElement(key interface{}) *Element {
	m.mu.RLock()
	defer m.mu.RUnlock()

	element, ok := m.kv[key]
	if ok {
		return element
	}
	return nil
}

// Len returns the number of elements in the map.
// It's for inner.
func (m *ZMap) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.kv)
}

// Keys returns all keys in the order they were inserted. If a key was
// replaced it will retain the same position. To ensure most recently set keys
// are always at the end you must always Delete before Set.
func (m *ZMap) Keys() (keys []interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys = make([]interface{}, 0, len(m.kv))
	for el := m.Front(); el != nil; el = el.Next() {
		keys = append(keys, el.Key)
	}
	return keys
}

// Delete will remove a key from the map. It will return true if the key was
// removed (the key did exist).
func (m *ZMap) Delete(key interface{}) (didDelete bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	element, ok := m.kv[key]
	if ok {
		m.ll.Remove(element)
		delete(m.kv, key)
	}

	return ok
}

// Front will return the element that is the first (oldest Set element). If
// there are no elements this will return nil.
func (m *ZMap) Front() *Element {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.ll.Front()
}

// Back will return the element that is the last (most recent Set element). If
// there are no elements this will return nil.
func (m *ZMap) Back() *Element {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.ll.Back()
}

// Copy returns a new OrderedMap with the same elements.
// Using Copy while there are concurrent writes may mangle the result.
func (m *ZMap) Copy() *ZMap {
	m2 := NewZMap()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for el := m.Front(); el != nil; el = el.Next() {
		m2.Set(el.Key, el.Value)
	}
	return m2
}

func (m *ZMap) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.kv) == 0 {
		return []byte("{}"), nil
	}

	var buf bytes.Buffer
	buf.WriteByte('{')
	encoder := json.NewEncoder(&buf)
	for elem := m.Front(); elem != nil; elem = elem.Next() {
		if err := encoder.Encode(elem.Key); err != nil {
			return nil, err
		}
		buf.WriteByte(':')
		if err := encoder.Encode(elem.Value); err != nil {
			return nil, err
		}
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
