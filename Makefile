.PHONY: check build test lint install

check: lint test build

build:
	go build ./...

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
	if command -v golangci-lint > /dev/null; then golangci-lint run; else echo "golangci-lint not installed, skipping"; fi

install:
	go build -o $(shell go env GOPATH)/bin/unbound-force ./cmd/unbound-force/
	ln -sf $(shell go env GOPATH)/bin/unbound-force $(shell go env GOPATH)/bin/uf
