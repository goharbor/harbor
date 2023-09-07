.PHONY: all test testrace int

all: test

test:
	go test ./...

testrace:
	go test -race ./...

int:
	INT=1 go test ./...
