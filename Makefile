.PHONY: run build test cover vet fmt lint check

run:
	go run ./cmd/server

build:
	go build -o bin/wheater-api ./cmd/server

test:
	go test ./... -count=1

cover:
	go test ./... -count=1 -cover

vet:
	go vet ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

check: vet lint test