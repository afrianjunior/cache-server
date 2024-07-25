// src/main.go
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
)

type Config struct {
    Port     string `json:"port"`
    CacheDir string `json:"cache_dir"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, NixOS Cache Server!")
}

func main() {
    file, err := os.Open("config.json")
    if err != nil {
        log.Fatalf("Error opening config file: %v", err)
    }
    defer file.Close()

    byteValue, _ := ioutil.ReadAll(file)
    var config Config
    json.Unmarshal(byteValue, &config)

    if config.Port == "" {
        config.Port = "8080"
    }

    fmt.Printf("Starting server on port %s with cache directory %s\n", config.Port, config.CacheDir)
    http.HandleFunc("/", handler)
    http.ListenAndServe("127.0.0.1:"+config.Port, nil)
}
