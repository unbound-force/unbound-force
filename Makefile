.PHONY: check build test lint

check: lint test build

build:
	go build ./...

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
	if command -v golangci-lint > /dev/null; then golangci-lint run; else echo "golangci-lint not installed, skipping"; fi
