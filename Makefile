build:
	go build -o mdc .

test:
	go test ./internal/... -v

test-integration:
	go test ./test/... -v

test-all:
	go test ./... -v

test-cover:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -func=coverage.out
