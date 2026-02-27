GO ?= go
APP_NAME ?= user-auth
BIN_DIR ?= bin
MAIN_PKG ?= .

BIN_PATH := $(BIN_DIR)/$(APP_NAME)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make help    - Show this help message"
	@echo "  make fmt     - Format Go code"
	@echo "  make vet     - Run go vet"
	@echo "  make test    - Run unit tests"
	@echo "  make tidy    - Tidy go.mod and go.sum"
	@echo "  make check   - Run fmt, vet, and test"
	@echo "  make build   - Build binary to $(BIN_PATH)"
	@echo "  make run     - Run the application"
	@echo "  make clean   - Remove build artifacts"

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test ./...

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: check
check: fmt vet test

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_PATH) $(MAIN_PKG)

.PHONY: run
run:
	$(GO) run $(MAIN_PKG)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
