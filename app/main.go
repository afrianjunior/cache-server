package main

import (
	"fmt"
	"net/http"
	"sync"
)

type Cache struct {
	mu    sync.Mutex
	cache map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if val, ok := c.cache[key]; ok {
		return val, true
	}
	return "", false
}

func (c *Cache) Set(key string, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = val
}

func main() {
	cache := &Cache{
		cache: make(map[string]string),
	}

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val, ok := cache.Get(key)
		if ok {
			fmt.Fprint(w, val)
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("val")
		cache.Set(key, val)
	})

	fmt.Println("Server start at port 8000")
	http.ListenAndServe(":8080", nil)
}
