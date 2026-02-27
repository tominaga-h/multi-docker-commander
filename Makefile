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

lint:
	go vet ./...
	golangci-lint run

check:
	make lint && make test-all

install-hooks:
	chmod +x githooks/pre-push.sh
	cp githooks/pre-push.sh .git/hooks/pre-push
