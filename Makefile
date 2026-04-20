.PHONY: build build-all clean test

BINARY=crolab
VERSION?=0.2.0
BUILD_DIR=dist

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/crolab/

build-all: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-linux-amd64   ./cmd/crolab/
	GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-linux-arm64   ./cmd/crolab/
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64  ./cmd/crolab/
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64  ./cmd/crolab/
	@echo "✓ Binários em $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

test:
	go test -v -count=1 ./tests/unit/...
	go test -v -count=1 ./tests/load/...

test-e2e: build
	go test -v -count=1 ./tests/e2e/...

clean:
	rm -rf $(BUILD_DIR) $(BINARY)
