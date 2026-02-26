VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -o mdc .

build-v:
	go build -ldflags "-X mdc/internal/version.Version=$(VERSION)" -o mdc .

test:
	go test ./internal/... -v

test-integration:
	go test ./test/... -v

test-all:
	go test ./... -v

test-cover:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -func=coverage.out
