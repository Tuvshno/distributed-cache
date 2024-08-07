package cache

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Cache struct {
	lock sync.RWMutex
	data map[string][]byte
}

func New() *Cache {
	return &Cache{
		data: make(map[string][]byte),
	}
}

func (c *Cache) Delete(key []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.data, string(key))

	return nil
}

func (c *Cache) Has(key []byte) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	_, ok := c.data[string(key)]
	return ok
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	keyStr := string(key)
	val, ok := c.data[keyStr]
	if !ok {
		return nil, fmt.Errorf("key %s not found", keyStr)
	}

	log.Printf("GET %s = %s", string(key), string(val))

	return val, nil
}

func (c *Cache) Set(key, value []byte, ttl time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.data[string(key)] = value
	log.Printf("SET %s to %s \n", string(key), string(value))

	go func() {
		<-time.After(ttl * time.Second)
		delete(c.data, string(key))
	}()

	return nil
}

func (c *Cache) Print() {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if len(c.data) == 0 {
		fmt.Println("Cache is empty")
		return
	}

	fmt.Println("Cache contents:")
	for key, value := range c.data {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
}
