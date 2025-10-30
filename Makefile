.PHONY: build build-edge-cd-go test-unit test-e2e test-e2e-shell test-e2e-go clean

build: build-edge-cd-go

build-edge-cd-go:
	@echo "Building edge-cd-go..."
	@mkdir -p bin
	@go build -o bin/edge-cd-go ./cmd/edge-cd-go
	@echo "Built bin/edge-cd-go"

test-unit:
	@echo "Running unit tests..."
	@export SRC_DIR=$(CURDIR)/cmd/edge-cd; \
	for f in test/edge-cd/lib/*_test.sh; do \
		echo "Running $$f"; \
		bash $$f; \
	done
	@go test -v ./cmd/...
	@go test -v ./pkg/...
	@echo "Unit tests finished."

test-e2e: test-e2e-shell test-e2e-go
	@echo "All end-to-end tests finished."

test-e2e-shell:
	@echo "Running end-to-end tests with shell implementation..."
	@bash ./test/edge-cd/e2e/test.sh shell
	@go run ./cmd/edgectl-e2e test
	@echo "Shell E2E tests finished."

test-e2e-go:
	@echo "Running end-to-end tests with Go implementation..."
	@bash ./test/edge-cd/e2e/test.sh go
	@go run ./cmd/edgectl-e2e test
	@echo "Go E2E tests finished."

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "Clean complete."
