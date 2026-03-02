APP_DIR := Visual-Network/frontend
NPM := npm

.PHONY: help setup run build test clean

help:
	@echo "Available targets:"
	@echo "  make setup  - Install frontend dependencies"
	@echo "  make run    - Start development server"
	@echo "  make build  - Build production bundle"
	@echo "  make test   - Run tests"
	@echo "  make clean  - Remove node_modules and build output"

setup:
	$(NPM) install

run:
	$(NPM) start

build:
	$(NPM) run build

test:
	$(NPM) test

clean:
	rm -rf node_modules/ build/