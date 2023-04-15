.PHONY: all clean

BINARY_NAME=arbitrage

all: clean build

clean:
	go clean
	rm -f bin/*

build:
	go build -o bin/$(BINARY_NAME) cmd/main.go