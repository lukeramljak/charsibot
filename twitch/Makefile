.PHONY: build run dev

vet:
	go vet ./...

build: vet
	go build -o bin/charsibot ./cmd/charsibot

run: build
	./bin/charsibot

dev:
	air

test:
	go test ./...

lint:
	golangci-lint run
