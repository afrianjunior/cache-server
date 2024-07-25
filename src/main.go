package main

import (
    "bytes"
    "encoding/json"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"

    "golang.org/x/crypto/openpgp"
)

type Config struct {
    Port     string `json:"port"`
    CacheDir string `json:"cacheDir"`
}

var (
    config     Config
    cacheDir   string
    publicKeys = "/etc/nix/public-keys/public.key" // Path to your public key file
)

func loadConfig() {
    file, err := os.Open("config.json")
    if err != nil {
        log.Fatalf("Error opening config file: %v", err)
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&config); err != nil {
        log.Fatalf("Error decoding config file: %v", err)
    }
    cacheDir = config.CacheDir
}

func verifySignature(filePath string) error {
    // Load the public key
    keyFile, err := os.Open(publicKeys)
    if err != nil {
        return err
    }
    defer keyFile.Close()

    keyRing, err := openpgp.ReadArmoredKeyRing(keyFile)
    if err != nil {
        return err
    }

    // Open the file to be verified
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Read the signature and payload
    sigFilePath := filePath + ".sig" // Assuming the signature is stored separately with a .sig extension
    sigFile, err := os.Open(sigFilePath)
    if err != nil {
        return err
    }
    defer sigFile.Close()

    sigData, err := ioutil.ReadAll(sigFile)
    if err != nil {
        return err
    }

    fileData, err := ioutil.ReadAll(file)
    if err != nil {
        return err
    }

    // Verify the signature
    packetReader := openpgp.EntityList{}
    for _, key := range keyRing {
        packetReader = append(packetReader, key)
    }

    _, err = openpgp.CheckDetachedSignature(packetReader, bytes.NewReader(fileData), bytes.NewReader(sigData))
    if err != nil {
        return err
    }

    return nil
}

func main() {
    loadConfig()

    http.HandleFunc("/cache/", cacheHandler)

    log.Printf("Nix cache server is running on port %s...", config.Port)
    if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}

func cacheHandler(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path[len("/cache/"):]

    switch r.Method {
    case http.MethodGet:
        getHandler(w, path)
    case http.MethodPut:
        putHandler(w, path, r.Body)
    default:
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    }
}

func getHandler(w http.ResponseWriter, path string) {
    fullPath := filepath.Join(cacheDir, path)

    // Verify the signature
    if err := verifySignature(fullPath); err != nil {
        http.Error(w, "Signature verification failed: "+err.Error(), http.StatusUnauthorized)
        return
    }

    if _, err := os.Stat(fullPath); os.IsNotExist(err) {
        http.Error(w, "Cache miss", http.StatusNotFound)
        return
    }

    http.ServeFile(w, nil, fullPath)
}

func putHandler(w http.ResponseWriter, path string, body io.Reader) {
    fullPath := filepath.Join(cacheDir, path)

    file, err := os.Create(fullPath)
    if err != nil {
        http.Error(w, "Failed to store cache", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    if _, err := io.Copy(file, body); err != nil {
        http.Error(w, "Failed to store cache", http.StatusInternalServerError)
        return
    }

    // Optionally, sign the artifact here

    w.WriteHeader(http.StatusNoContent)
}
