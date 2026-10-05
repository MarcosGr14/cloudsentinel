package main

import (
	"container/list"
	"sync"
)

// Record guarda el estado de una IP conectada
type Record struct {
	IP    string
	Count int
}

// LRUCache es una estructura thread-safe para rastrear IPs
type LRUCache struct {
	capacity  int
	mutex     sync.RWMutex
	items     map[string]*list.Element
	evictList *list.List
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity:  capacity,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

// AddOrUpdate incrementa el contador de la IP y la mueve al frente
func (c *LRUCache) AddOrUpdate(ip string) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Si la IP ya existe en el caché
	if elem, exists := c.items[ip]; exists {
		c.evictList.MoveToFront(elem)
		record := elem.Value.(*Record)
		record.Count++
		return record.Count
	}

	// Si el caché alcanzó su capacidad máxima, eliminamos el más antiguo
	if c.evictList.Len() >= c.capacity {
		c.removeOldest()
	}

	// Agregar nueva IP
	newRecord := &Record{IP: ip, Count: 1}
	elem := c.evictList.PushFront(newRecord)
	c.items[ip] = elem
	return 1
}

func (c *LRUCache) removeOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		c.evictList.Remove(elem)
		record := elem.Value.(*Record)
		delete(c.items, record.IP)
	}
}
