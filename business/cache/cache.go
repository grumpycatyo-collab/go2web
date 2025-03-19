package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CacheItem struct {
	Data       []byte    `json:"data"`
	Expiration time.Time `json:"expiration"`
}

type Cache struct {
	mutex    sync.RWMutex
	cacheDir string
}

func NewCache() *Cache {
	cacheDir := "./cache"

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Printf("Failed to create cache directory: %v\n", err)
		cacheDir = "."
	}

	cache := &Cache{
		cacheDir: cacheDir,
	}

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			cache.cleanup()
		}
	}()

	return cache
}

func (c *Cache) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	hashedKey := c.hashKey(key)
	filePath := filepath.Join(c.cacheDir, hashedKey)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil
	}

	var item CacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		fmt.Printf("Error unmarshaling cache item: %v\n", err)
		return nil
	}

	if time.Now().After(item.Expiration) {
		os.Remove(filePath)
		return nil
	}

	return item.Data
}

func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var data []byte

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	case io.Reader:
		var err error
		data, err = ioutil.ReadAll(v)
		if err != nil {
			fmt.Printf("Error reading from io.Reader: %v\n", err)
			return
		}
	default:
		var err error
		data, err = json.Marshal(v)
		if err != nil {
			fmt.Printf("Error marshaling value to JSON: %v\n", err)
			return
		}
	}

	item := CacheItem{
		Data:       data,
		Expiration: time.Now().Add(5 * time.Minute),
	}

	itemData, err := json.Marshal(item)
	if err != nil {
		fmt.Printf("Error marshaling cache item: %v\n", err)
		return
	}

	hashedKey := c.hashKey(key)
	filePath := filepath.Join(c.cacheDir, hashedKey)

	if err := ioutil.WriteFile(filePath, itemData, 0644); err != nil {
		fmt.Printf("Error writing cache file: %v\n", err)
	}
}

func (c *Cache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	files, err := ioutil.ReadDir(c.cacheDir)
	if err != nil {
		fmt.Printf("Error reading cache directory: %v\n", err)
		return
	}

	now := time.Now()
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(c.cacheDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue
		}

		var item CacheItem
		if err := json.Unmarshal(data, &item); err != nil {
			os.Remove(filePath)
			continue
		}

		if now.After(item.Expiration) {
			os.Remove(filePath)
		}
	}
}

func (c *Cache) hashKey(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
