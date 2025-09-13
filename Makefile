.PHONY: all

build:
	go build -ldflags="-s -w" -o bin/dolme cmd/main.go
