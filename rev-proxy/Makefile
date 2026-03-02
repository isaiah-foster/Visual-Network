.PHONY: help setup build run clean generate-certs

help:
	@echo "Available targets:"
	@echo "  make setup  - Install dependencies"
	@echo "  make build  - Build reverse proxy binary"
	@echo "  make run    - Start reverse proxy server"
	@echo "  make clean  - Clean build artifacts"
	@echo "  make generate-certs - Generate self-signed TLS certificates (dev only)"

setup:
	go mod download
	go mod tidy

build:
	@mkdir -p bin
	go build -o bin/reverse-proxy .

run:
	@if [ ! -f "bin/reverse-proxy" ]; then \
		$(MAKE) build; \
	fi
	./bin/reverse-proxy \
		-listen ":8443" \
		-frontend "../frontend/build" \
		-auth "http://localhost:8080" \
		-backend "http://localhost:9090" \
		-mysql-dsn "$(MYSQL_DSN)"

run-no-auth:
	@if [ ! -f "bin/reverse-proxy" ]; then \
		$(MAKE) build; \
	fi
	./bin/reverse-proxy \
		-listen ":8443" \
		-frontend "../frontend/build" \
		-auth "http://localhost:8080" \
		-backend "http://localhost:9090"

clean:
	rm -f bin/reverse-proxy
	rm -f certs/*.pem

generate-certs:
	@mkdir -p certs
	openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem \
		-days 365 -nodes -subj "/CN=localhost"
	@echo "Self-signed certs generated in certs/"
