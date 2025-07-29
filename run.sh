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
    "integration-test-keep")
        echo "Running integration tests (keeping files)..."
        go run cmd/main.go -test -keep-files
        ;;
    "clean")
        echo "Cleaning build artifacts..."
        rm -rf bin/
        ;;
    *)
        echo "Usage: $0 {build|test|integration-test|integration-test-keep|clean}"
        echo ""
        echo "Commands:"
        echo "  build                 Build the docgen MCP server"
        echo "  test                  Run unit tests"
        echo "  integration-test      Run integration tests (temp files deleted)"
        echo "  integration-test-keep Run integration tests (keep generated files)"
        echo "  clean                 Remove build artifacts"
        exit 1
        ;;
esac