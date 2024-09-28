.PHONY: build test build-docker run test-one

build:
	CGO_ENABLED=0 GOOS=linux go build -o ./playground-mono

format:
	go fmt ./...

run:
	go run main.go

# TODO: Uncomment once integration-tests are added
#test:
#	go test -v -count=1 -p=1 ./integration_tests/...
#
#test-one:
#	go test -v -race -count=1 --run $(TEST_FUNC) ./integration_tests/...

unittest:
	go test -race -count=1 ./internal/...

unittest-one:
	go test -v -race -count=1 --run $(TEST_FUNC) ./internal/...

GOLANGCI_LINT_VERSION := 1.60.1

lint: check-golangci-lint
	golangci-lint run ./...

check-golangci-lint:
	@if ! command -v golangci-lint > /dev/null || ! golangci-lint version | grep -q "$(GOLANGCI_LINT_VERSION)"; then \
		echo "Required golangci-lint version $(GOLANGCI_LINT_VERSION) not found."; \
		echo "Please install golangci-lint version $(GOLANGCI_LINT_VERSION) with the following command:"; \
		echo "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.60.1"; \
		exit 1; \
	fi
