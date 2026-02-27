APP_DIR := Visual-Network
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
	cd $(APP_DIR) && $(NPM) install

run:
	cd $(APP_DIR) && $(NPM) start

build:
	cd $(APP_DIR) && $(NPM) run build

test:
	cd $(APP_DIR) && $(NPM) test

clean:
	rm -rf $(APP_DIR)/node_modules $(APP_DIR)/build