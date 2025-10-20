.PHONY: test-unit

test-unit:
	@echo "Running unit tests..."
	@export SRC_DIR=$(CURDIR)/cmd/edge-cd; \
	for f in test/edge-cd/lib/test_*.sh; do \
		echo "Running $$f"; \
		bash $$f; \
	done
	@echo "Unit tests finished."