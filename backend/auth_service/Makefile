GO ?= go
APP_NAME ?= user-auth
BIN_DIR ?= bin
MAIN_PKG ?= .
MYSQL_PKG ?= mysql-server
MYSQL_SERVICE ?= mysql

BIN_PATH := $(BIN_DIR)/$(APP_NAME)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make help    - Show this help message"
	@echo "  make setup   - Download Go modules and show MySQL setup guidance"
	@echo "  make deps    - Download Go modules required by this project"
	@echo "  make mysql-ubuntu - Install/configure MySQL on Ubuntu and print schema template"
	@echo "  make mysql-schema-template - Print the SQL template for the user_auth table"
	@echo "  make mysql-create-db - Create the database and user using db/db_init.sql"
	@echo "  make fmt     - Format Go code"
	@echo "  make vet     - Run go vet"
	@echo "  make test    - Run unit tests"
	@echo "  make tidy    - Tidy go.mod and go.sum"
	@echo "  make check   - Run fmt, vet, and test"
	@echo "  make build   - Build binary to $(BIN_PATH)"
	@echo "  make run     - Run the application"
	@echo "  make clean   - Remove build artifacts"

.PHONY: setup
setup: deps mysql-ubuntu

.PHONY: deps
deps:
	sudo apt install golang
	$(GO) mod download
	$(GO) mod verify

.PHONY: mysql-ubuntu
mysql-ubuntu:
	@set -e; \
	echo "Checking for MySQL on Ubuntu..."; \
	if ! command -v mysql >/dev/null 2>&1; then \
		echo "MySQL client not found. Installing $(MYSQL_PKG)..."; \
		sudo apt-get update; \
		sudo DEBIAN_FRONTEND=noninteractive apt-get install -y $(MYSQL_PKG); \
	else \
		echo "MySQL is already installed."; \
	fi; \
	echo "Ensuring the MySQL service is enabled and running..."; \
	sudo systemctl enable --now $(MYSQL_SERVICE); \
	echo ""; \
	echo "Launching mysql_secure_installation."; \
	echo "Follow the prompts to set your local security options."; \
	sudo mysql_secure_installation; \
	echo ""; \a
	$(MAKE) mysql-schema-template

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
	$(GO) run ./cmd/seed


.PHONY: mysql-schema-template
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
