package utils

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type CacheItem struct {
	Key       string
	Value     interface{}
	expiry    time.Time
	frequency int
}

type LRUCache struct {
	Capacity int
	Items    map[string]*list.Element
	Queue    *list.List
	Lock     sync.Mutex
	stopChan chan bool
	callback LRUCallback
}

func NewLRUCache(capacity int, callback LRUCallback) *LRUCache {
	cache := &LRUCache{
		Capacity: capacity,
		Items:    make(map[string]*list.Element),
		Queue:    list.New(),
		stopChan: make(chan bool),
		callback: callback,
	}
	go cache.ExpiryKeyScanner()
	return cache
}

func (c *LRUCache) Get(key string) interface{} {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if elem, ok := c.Items[key]; ok {
		c.Queue.MoveToFront(elem)
		elem.Value.(*CacheItem).frequency++
		return elem.Value.(*CacheItem).Value
	}
	return nil
}

func (c *LRUCache) Set(key string, value interface{}, timeout time.Duration) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if c.Queue.Len() >= c.Capacity {
		oldest := c.Queue.Back()
		c.Queue.Remove(oldest)
		delete(c.Items, oldest.Value.(*CacheItem).Key)
	}

	var expiry time.Time
	if timeout != 0 {
		expiry = time.Now().Add(timeout)
	}
	item := &CacheItem{Key: key, Value: value, expiry: expiry, frequency: 0}
	entry := c.Queue.PushFront(item)
	c.Items[key] = entry

	return nil
}

func (c *LRUCache) IsExist(key string) bool {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	_, exists := c.Items[key]
	return exists
}

func (c *LRUCache) Delete(key string) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if elem, ok := c.Items[key]; ok {
		c.Queue.Remove(elem)
		delete(c.Items, key)
		return nil
	}
	return errors.New("Key not found")
}

type LRUCallback func(key string, value interface{}, freq int)

func (c *LRUCache) ExpiryKeyScanner() {
	for {
		select {
		case <-c.stopChan:
			c.Lock.Lock()
			for key, item := range c.Items {
				if item.Value.(*CacheItem).frequency > 0 {
					c.callback(key, item.Value.(*CacheItem).Value, item.Value.(*CacheItem).frequency)
					item.Value.(*CacheItem).frequency = 0
				}

			}

			c.Lock.Unlock()
			return
		default:
			time.Sleep(1 * time.Minute)
			c.Lock.Lock()
			for key, item := range c.Items {
				if !item.Value.(*CacheItem).expiry.IsZero() && item.Value.(*CacheItem).expiry.Before(time.Now()) {
					c.callback(key, item.Value.(*CacheItem).Value, item.Value.(*CacheItem).frequency)
					c.Delete(key)
				}

				if item.Value.(*CacheItem).frequency > 0 {
					c.callback(key, item.Value.(*CacheItem).Value, item.Value.(*CacheItem).frequency)
					item.Value.(*CacheItem).frequency = 0
				}

			}

			c.Lock.Unlock()
		}
	}
}

func (c *LRUCache) StopExpiryKeyScanner() {
	c.stopChan <- true
}
