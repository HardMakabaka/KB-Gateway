.PHONY: test build run

test:
	go test ./...

build:
	go build ./cmd/kb-gateway

run:
	go run ./cmd/kb-gateway
