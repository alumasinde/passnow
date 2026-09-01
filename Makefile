# PassNow development commands.
#
# Examples:
#   make migrate
#   make run
#   make test
#   make build

GO ?= go
BIN_DIR ?= bin
API_CMD ?= ./cmd/api
MIGRATE_CMD ?= ./cmd/migrate

.PHONY: help migrate run test build clean

help:
	@echo "PassNow commands:"
	@echo "  make migrate  - apply pending database migrations"
	@echo "  make run      - start the API server"
	@echo "  make test     - run the Go test suite"
	@echo "  make build    - build API and migration binaries"
	@echo "  make clean    - remove local build artifacts"

migrate:
	$(GO) run $(MIGRATE_CMD) -action up

run:
	$(GO) run $(API_CMD)

test:
	$(GO) test ./...

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/passnow-api $(API_CMD)
	$(GO) build -o $(BIN_DIR)/passnow-migrate $(MIGRATE_CMD)
	@echo "Build complete:"
	@echo "  $(BIN_DIR)/passnow-api"
	@echo "  $(BIN_DIR)/passnow-migrate"

clean:
	@rm -rf $(BIN_DIR)
	@echo "Build artifacts removed."
