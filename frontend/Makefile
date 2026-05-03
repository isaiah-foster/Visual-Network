.PHONY: help setup run build test clean

NPM := npm

help:
	@echo "Available targets:"
	@echo "  make setup  - Install frontend dependencies"
	@echo "  make run    - Start development server"
	@echo "  make build  - Build production bundle"
	@echo "  make test   - Run tests"
	@echo "  make clean  - Remove node_modules and build output"

setup:
	sudo apt install nodejs npm
	$(NPM) install

run:
	$(NPM) start

build:
	$(NPM) run build

test:
	$(NPM) test

clean:
	rm -rf node_modules/ build/