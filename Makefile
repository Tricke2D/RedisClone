# Makefile untuk Redis Clone project
# Automation untuk build, test, run, clean

.PHONY: help setup build test test-race run clean lint benchmark

help:
	@echo "Redis Clone - Development Commands"
	@echo "make setup       - Setup development environment"
	@echo "make build       - Build server binary"
	@echo "make test        - Run all tests"
	@echo "make test-race   - Run tests dengan race condition detector"
	@echo "make run         - Run server locally (port 6379)"
	@echo "make clean       - Remove artifacts"
	@echo "make lint        - Run linter"
	@echo "make benchmark   - Run benchmark tests"

setup:
	go mod download
	go mod tidy
	@echo "✓ Dependencies installed"

build:
	go build -o bin/redis-clone cmd/server/main.go
	@echo "✓ Binary built: ./bin/redis-clone"

test:
	go test -v -count=1 ./...
	@echo "✓ All tests passed"

test-race:
	go test -v -race ./...
	@echo "✓ No race conditions detected"

run: build
	@echo "Starting Redis Clone server on localhost:6379..."
	./bin/redis-clone

clean:
	rm -rf bin/
	go clean
	@echo "✓ Cleanup complete"

lint:
	golangci-lint run ./... || true

benchmark:
	go test -bench=. -benchmem ./...