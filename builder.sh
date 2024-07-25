#!/bin/bash

# Change into the app folder
cd app

# Build the Go-based cache server
go build -o cache-server main.go

# Run the cache server
./cache-server
