#!/bin/bash
# Add GOPATH/bin to PATH so wails can be found
export PATH=$PATH:$(go env GOPATH)/bin

# Check if wails is installed
if ! command -v wails &> /dev/null; then
    echo "Error: wails binary not found in $(go env GOPATH)/bin"
    exit 1
fi

echo "Starting Wails dev..."
wails dev
