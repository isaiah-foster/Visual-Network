GO ?= go
APP_NAME ?= user-auth
BIN_DIR ?= bin
MAIN_PKG ?= .
SHELL := /bin/bash

BIN_PATH := $(BIN_DIR)/$(APP_NAME)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make deps              - Install Go and download modules"
	@echo "  make mysql-ubuntu      - Install and configure MySQL on Ubuntu"
	@echo "  make mysql-create-db   - Create database and table from db/db_init.sql"
	@echo "  make mysql-create-user - Create DB user (requires DB_USER, DB_PASS, DB_NAME)"
	@echo "  make .env-file         - Create .env file (requires DB_USER, DB_PASS, DB_NAME)"
	@echo "  make seed-admin        - Seed the first admin user"
	@echo "  make build             - Build binary to $(BIN_PATH)"
	@echo "  make run               - Run the auth service"
	@echo "  make clean             - Remove build artifacts"


.PHONY: deps
deps:
	sudo apt install golang
	$(GO) mod download
	$(GO) mod verify

.PHONY: mysql-ubuntu
mysql-ubuntu:
	@set -e; \
	if ! command -v mysql >/dev/null 2>&1; then \
		sudo apt-get update; \
		sudo DEBIAN_FRONTEND=noninteractive apt-get install -y mysql-server; \
	else \
		echo "MySQL is already installed."; \
	fi; \
	sudo systemctl enable --now mysql; \
	sudo mysql_secure_installation

.PHONY: .env-file
.env-file:
	@[ -n "$(DB_USER)" ] || (echo "DB_USER is required"; exit 1)
	@[ -n "$(DB_PASS)" ] || (echo "DB_PASS is required"; exit 1)
	@[ -n "$(DB_NAME)" ] || (echo "DB_NAME is required"; exit 1)
	@echo 'export MYSQL_DSN="$(DB_USER):$(DB_PASS)@tcp(127.0.0.1:3306)/$(DB_NAME)?parseTime=true"' > .env
	@echo ".env created"
	@echo "You can now run 'source .env' to load the environment variables for the auth service."

.PHONY: mysql-create-db
mysql-create-db:
	sudo mysql < db/db_init.sql

.PHONY: mysql-create-user
mysql-create-user:
	@[ -n "$(DB_USER)" ] || (echo "DB_USER is required"; exit 1)
	@[ -n "$(DB_PASS)" ] || (echo "DB_PASS is required"; exit 1)
	@[ -n "$(DB_NAME)" ] || (echo "DB_NAME is required"; exit 1)
	sudo mysql -e "CREATE USER IF NOT EXISTS '$(DB_USER)'@'localhost' IDENTIFIED BY '$(DB_PASS)'; \
				GRANT ALL PRIVILEGES ON $(DB_NAME).* TO '$(DB_USER)'@'localhost'; \
				FLUSH PRIVILEGES;"

.PHONY: seed-admin
seed-admin:
	. .env && $(GO) run ./cmd/seed

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_PATH) $(MAIN_PKG)

.PHONY: run
run:
	. .env && $(GO) run $(MAIN_PKG)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)