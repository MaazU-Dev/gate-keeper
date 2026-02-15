.PHONY: build run test clean lint

APP_NAME := gate-keeper
BUILD_DIR := out

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/gate-keeper

run: build
	./$(BUILD_DIR)/$(APP_NAME)

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)

lint:
	golangci-lint run ./...
