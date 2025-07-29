#!/bin/bash

case "$1" in
    "build")
        echo "Building docgen MCP server..."
        go build -o bin/docgen cmd/main.go
        ;;
    "test")
        echo "Running unit tests..."
        go test ./pkg/...
        ;;
    "integration-test")
        echo "Running integration tests..."
        go run cmd/main.go -test
        ;;
    "clean")
        echo "Cleaning build artifacts..."
        rm -rf bin/
        ;;
    *)
        echo "Usage: $0 {build|test|integration-test|clean}"
        exit 1
        ;;
esac