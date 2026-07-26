CONFIG_PATH ?= ./config/local.yaml
BIN := ./bin/url-shortener

.PHONY: run build test clean

run:
	CONFIG_PATH=$(CONFIG_PATH) go run ./cmd/url-shortener

build:
	go build -o $(BIN) ./cmd/url-shortener

test:
	go test ./...

clean:
	rm -rf ./bin