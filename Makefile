.PHONY: build build-cli build-server run-server run-cli dev migrate clean build-all test

# Strip debug symbols for smaller binaries
LDFLAGS := -ldflags="-s -w"

# Build both binaries
build: build-cli build-server

build-cli:
	go build $(LDFLAGS) -o bin/pidrive ./cmd/pidrive

build-server:
	go build $(LDFLAGS) -o bin/pidrive-server ./cmd/pidrive-server

# Run server locally
run-server: build-server
	source .env 2>/dev/null; ./bin/pidrive-server

# Dev: start infra + server
dev:
	docker compose up -d
	@echo "Waiting for postgres..."
	@sleep 3
	@make run-server

# Stop infra
down:
	docker compose down

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/pidrive-linux-amd64 ./cmd/pidrive
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/pidrive-linux-arm64 ./cmd/pidrive
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/pidrive-darwin-amd64 ./cmd/pidrive
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/pidrive-darwin-arm64 ./cmd/pidrive
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/pidrive-server-linux-amd64 ./cmd/pidrive-server
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/pidrive-server-linux-arm64 ./cmd/pidrive-server

# Run tests
test:
	go test ./...

clean:
	rm -rf bin/

# Tidy deps
tidy:
	go mod tidy
