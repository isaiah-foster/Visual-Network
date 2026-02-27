.PHONY: all clean generate build run install-deps setup

# Default target
all: install-deps clean generate build

# Complete setup (first time only)
setup:
	@echo "Running initial setup..."
	./setup.sh

# Install Go dependencies for backend and optionally auth_service

AUTH_DIR ?= auth_service

install-deps:
	@echo "Downloading Go modules (backend)..."
	go mod download
	@echo "Updating go.sum (backend)..."
	go mod tidy
	@echo "Installing bpf2go..."
	go install github.com/cilium/ebpf/cmd/bpf2go@latest

	# If auth_service has its own go.mod, download its modules too
	@if [ -f "$(AUTH_DIR)/go.mod" ]; then \
		echo "Downloading Go modules (auth_service)..."; \
		(cd $(AUTH_DIR) && go mod download && go mod verify) ; \
	fi

# Generate Go code from eBPF C code
generate: install-deps
	@echo "Generating eBPF Go bindings..."
	cd traffic-aggregation/ebpf && go generate

# Build the collector binary
BUILD_DIR ?= bin
COLLECTOR_BIN := $(BUILD_DIR)/collector

	# Auth service build vars
AUTH_APP_NAME ?= user-auth
AUTH_BIN_DIR := $(BUILD_DIR)
AUTH_BIN := $(AUTH_BIN_DIR)/$(AUTH_APP_NAME)

.PHONY: build build-collector build-auth
build: build-collector build-auth

build-collector:
	@mkdir -p $(BUILD_DIR)
	go build -o $(COLLECTOR_BIN) ./traffic-aggregation/main.go

build-auth:
	@mkdir -p $(AUTH_BIN_DIR)
	# Build auth service binary into backend bin/
	go build -o $(AUTH_BIN) ./auth_service

# Run the collector (requires sudo and interface name)
# Usage: make run IFACE=<network interface>
.PHONY: run run-collector run-auth
run: run-collector

run-collector:
	@if [ -z "$(IFACE)" ]; then \
		echo "Error: IFACE not specified. Usage: make run IFACE=<your network interface>"; \
		exit 1; \
	fi
	sudo $(COLLECTOR_BIN) $(IFACE)

# Run auth service (run built binary if present, else `go run`)
run-auth:
	@if [ -x "$(AUTH_BIN)" ]; then \
		echo "Running auth service binary: $(AUTH_BIN)"; \
		$(AUTH_BIN); \
	else \
		echo "Running auth service with 'go run' from $(AUTH_DIR)"; \
		cd $(AUTH_DIR) && go run .; \
	fi


# Formatting, vet, test, tidy (developer helpers)
.PHONY: fmt vet test tidy
fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

tidy:
	$(GO) mod tidy

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)/
	rm -rf $(AUTH_BIN_DIR)/
	rm -f traffic-aggregation/ebpf/tc_monitor_bpfel.go
	rm -f traffic-aggregation/ebpf/tc_monitor_bpfel.o
	rm -f traffic-aggregation/ebpf/tc_monitor_bpfeb.go
	rm -f traffic-aggregation/ebpf/tc_monitor_bpfeb.o

# Development: watch for changes and rebuild (collector only)
watch:
	while true; do \
		$(MAKE) build-collector; \
		inotifywait -qre close_write .; \
	done

# Check for required tools
check-tools:
	@command -v clang >/dev/null 2>&1 || { echo "Error: clang not found. Install with: apt-get install clang"; exit 1; }
	@command -v llvm-strip >/dev/null 2>&1 || { echo "Error: llvm-strip not found. Install with: apt-get install llvm"; exit 1; }
	@echo "All required tools are installed"

## Additional targets integrated from auth_service/Makefile
# Formatting, vet, test, tidy, auth-specific setup
GO ?= go

.PHONY: help
help:
	@echo "Visual Network - backend Makefile (collector + auth_service)"
	@echo ""
	@echo "Targets:"
	@echo "  all                 - Install deps, generate eBPF code and build both binaries"
	@echo "  install-deps        - Install Go deps (backend and auth_service if present)"
	@echo "  generate            - Generate eBPF Go bindings"
	@echo "  build               - Build collector and auth service binaries"
	@echo "  build-collector     - Build collector binary"
	@echo "  build-auth          - Build auth service binary"
	@echo "  run-collector       - Run collector (requires IFACE=<interface>)"
	@echo "  run-auth            - Run auth service (built binary or go run)"
	@echo "  clean               - Remove build artifacts"
	@echo "  check-tools         - Verify required native tools (clang, llvm-strip)"
	@echo "  fmt                 - go fmt ./..."
	@echo "  vet                 - go vet ./..."
	@echo "  test                - go test ./..."
	@echo "  tidy                - go mod tidy"
	@echo "  mysql-ubuntu        - Install/configure MySQL on Ubuntu for auth service"
	@echo "  mysql-schema-template - Print SQL template for user_auth table"
	@echo ""
	@echo "Examples:"
	@echo "  make all"
	@echo "  make run-collector IFACE=eth0"
	@echo "  make run-auth"
	@echo "  make build-auth"

# MySQL helpers (copied from auth_service/Makefile)
.PHONY: mysql-ubuntu mysql-schema-template
mysql-ubuntu:
	@set -e; \
		echo "Checking for MySQL on Ubuntu..."; \
		if ! command -v mysql >/dev/null 2>&1; then \
			echo "MySQL client not found. Installing mysql-server..."; \
			sudo apt-get update; \
			sudo DEBIAN_FRONTEND=noninteractive apt-get install -y mysql-server; \
		else \
			echo "MySQL is already installed."; \
		fi; \
		echo "Ensuring the MySQL service is enabled and running..."; \
		sudo systemctl enable --now mysql; \
		echo ""; \
		echo "Launching mysql_secure_installation."; \
		echo "Follow the prompts to set your local security options."; \
		sudo mysql_secure_installation; \
		echo ""; \
		echo "Open a MySQL shell after this target completes with: sudo mysql"; \
		echo "Use the statements below to create your application database and user:"; \
		echo "  CREATE DATABASE your_database CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"; \
		echo "  CREATE USER 'your_user'@'localhost' IDENTIFIED BY 'your_password';"; \
		echo "  GRANT ALL PRIVILEGES ON your_database.* TO 'your_user'@'localhost';"; \
		echo "  FLUSH PRIVILEGES;"; \
		echo ""; \
		$(MAKE) mysql-schema-template

mysql-schema-template:
	@echo "Table template for this app (edit names/options as needed before running it):"
	@echo "  USE your_database;"
	@echo "  CREATE TABLE user_auth ("
	@echo "      user_id VARCHAR(64) NOT NULL,"
	@echo "      email VARCHAR(255) NOT NULL,"
	@echo "      pass_hash VARCHAR(255) NOT NULL,"
	@echo "      role VARCHAR(32) NOT NULL DEFAULT 'user',"
	@echo "      session_token_hash BINARY(32) NULL,"
	@echo "      session_expiration DATETIME NULL,"
	@echo "      is_active TINYINT(1) NOT NULL DEFAULT 1,"
	@echo "      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,"
	@echo "      updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,"
	@echo "      PRIMARY KEY (user_id),"
	@echo "      UNIQUE KEY uq_user_auth_email (email),"
	@echo "      KEY idx_user_auth_session_token_hash (session_token_hash),"
	@echo "      KEY idx_user_auth_session_expiration (session_expiration)"
	@echo "  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;"
	@echo ""
	@echo "Notes:"
	@echo "  - user_id is used as the login identifier by the current handlers."
	@echo "  - pass_hash should store the full Argon2id encoded hash string."
	@echo "  - session_token_hash stores a 32-byte SHA-256 hash."
	@echo "  - Adjust lengths, indexes, and extra columns before applying in your environment."

