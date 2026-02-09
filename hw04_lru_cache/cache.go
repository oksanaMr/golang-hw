package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheEntry struct {
	key   Key
	value interface{}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (lru *lruCache) Set(key Key, value interface{}) bool {
	if item, ok := lru.items[key]; ok {
		item.Value = cacheEntry{
			key:   key,
			value: value,
		}
		lru.queue.MoveToFront(item)
		return ok
	}

	// Если достигли емкости, удаляем самый старый элемент
	if lru.queue.Len() >= lru.capacity {
		lastNode := lru.queue.Back()
		if lastNode != nil {
			delete(lru.items, lastNode.Value.(cacheEntry).key)
			lru.queue.Remove(lastNode)
		}
	}

	newEntry := cacheEntry{
		key:   key,
		value: value,
	}
	newItem := lru.queue.PushFront(newEntry)
	lru.items[key] = newItem

	return false
}

func (lru *lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := lru.items[key]; ok {
		lru.queue.MoveToFront(item)

		entry := item.Value.(cacheEntry)
		return entry.value, true
	}
	return nil, false
}

func (lru *lruCache) Clear() {
	lru.queue = NewList()
	lru.items = make(map[Key]*ListItem, lru.capacity)
}
