BINARY_NAME := proj-audit
BIN_DIR := bin

.PHONY: all build run test clean

all: build

build:
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/proj-audit

run: build
	./$(BIN_DIR)/$(BINARY_NAME)

run-dev: build
	./$(BIN_DIR)/$(BINARY_NAME) --root ~/dev

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
