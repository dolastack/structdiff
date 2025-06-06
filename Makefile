BINARY=structdiff
GOFILES=$(wildcard *.go)

default: help

help:
	@echo "  Usage:"
	@echo "  make build        Build the binary"
	@echo "  make test         Run unit tests"
	@echo "  make clean        Remove binary and test cache"
	@echo "  make fmt          Format Go code"
	@echo "  make vet          Run go vet"
	@echo "  make lint         Run golangci-lint"
	@echo "  make tidy         Run go mod tidy"
	@echo "  make all          Clean, fmt, vet, tidy, build, test"

build:
	go build -o ${BINARY}

test:
	go test -v ./...

clean:
	rm -f ${BINARY}
	rm -rf ./test-results/

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run --deadline=3m

tidy:
	go mod tidy

all: clean fmt vet tidy build test

.PHONY: help build test clean fmt vet lint tidy all
