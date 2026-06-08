.PHONY: test coverage

# Get all packages that have test files
TEST_PACKAGES := $(shell go list ./... | grep -v '/mocks/' | while read pkg; do \
	if [ -n "$$(go list -f '{{.TestGoFiles}}' $$pkg | grep -v '\[\]')" ]; then \
		echo $$pkg; \
	fi \
done)

test:
	go test -v -race -count=1 ./...

coverage:
	@if [ -z "$(TEST_PACKAGES)" ]; then \
		echo "No packages with tests found"; \
		exit 1; \
	fi
	go test -v -race -count=1 -coverprofile=coverage.out -covermode=atomic $(TEST_PACKAGES)
	go tool cover -html=coverage.out -o coverage.html

coverage-ci:
	@if [ -z "$(TEST_PACKAGES)" ]; then \
		echo "No packages with tests found"; \
		exit 1; \
	fi
	go test -v -race -count=1 -coverprofile=coverage.out -covermode=atomic $(TEST_PACKAGES)