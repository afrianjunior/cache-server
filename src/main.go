package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

var (
	cache = make(map[string]string)
	mutex = &sync.RWMutex{}
)

func main() {
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/set", setHandler)

	log.Println("Cache server is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	mutex.RLock()
	value, exists := cache[key]
	mutex.RUnlock()

	if !exists {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"key": key, "value": value})
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	key, keyExists := body["key"]
	value, valueExists := body["value"]

	if !keyExists || !valueExists {
		http.Error(w, "Key and value are required", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	cache[key] = value
	mutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
