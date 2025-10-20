.PHONY: test-unit test-e2e

test-unit:
	@echo "Running unit tests..."
	@export SRC_DIR=$(CURDIR)/cmd/edge-cd; \
	for f in test/edge-cd/lib/test_*.sh; do \
		echo "Running $$f"; \
		bash $$f; \
	done
	@echo "Unit tests finished."

test-e2e:
	@echo "Running end-to-end tests..."
	@bash test/edge-cd/e2e/test.sh
	@echo "End-to-end tests finished."
