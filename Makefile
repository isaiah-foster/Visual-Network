.PHONY: help setup build run clean run-frontend run-backend

# Default target
help:
	@echo "Visual Network - Root Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make setup        - Install dependencies for both frontend and backend"
	@echo "  make build        - Build both frontend and backend"
	@echo "  make run          - Start both frontend and backend servers"
	@echo "  make run-frontend - Start only the frontend dev server"
	@echo "  make run-backend  - Start only the backend collector (requires IFACE=<interface>)"
	@echo "  make clean        - Clean build artifacts from both frontend and backend"
	@echo ""
	@echo "Setup from scratch after cloning:"
	@echo "  make setup && make build && make run-frontend"
	@echo ""
	@echo "Examples:"
	@echo "  make setup         # Install all dependencies"
	@echo "  make build         # Build both services"
	@echo "  make run-backend IFACE=eth0  # Run backend collector on eth0"

# Setup: Install dependencies for both frontend and backend
setup:
	@echo "Setting up Visual Network (frontend + backend)..."
	@echo ""
	@echo "=== Frontend Setup ==="
	cd frontend && $(MAKE) setup
	@echo ""
	@echo "=== Backend Setup ==="
	cd backend && $(MAKE) setup
	@echo ""
	@echo "✓ Setup complete! Run 'make build' to build both services."

# Build: Build both frontend and backend
build:
	@echo "Building Visual Network (frontend + backend)..."
	@echo ""
	@echo "=== Frontend Build ==="
	cd frontend && $(MAKE) build
	@echo ""
	@echo "=== Backend Build ==="
	cd backend && $(MAKE) generate && $(MAKE) build
	@echo ""
	@echo "✓ Build complete!"

# Run frontend and backend (instructions for running in separate terminals)
run:
	@echo "Visual Network - Start Services"
	@echo ""
	@echo "To run both services, open two terminals and run:"
	@echo ""
	@echo "  Terminal 1 (Frontend) :"
	@echo "    make run-frontend"
	@echo ""
	@echo "  Terminal 2 (Backend)  :"
	@echo "    make run-backend IFACE=<your-network-interface>"
	@echo ""
	@echo "Example network interfaces: eth0, wlan0, en0 (macOS), etc."
	@echo ""
	@echo "To find your interface, run:"
	@echo "    ip link show   (Linux)"
	@echo "    ifconfig       (macOS/older systems)"

# Run frontend only
run-frontend:
	@echo "Starting frontend development server..."
	cd frontend && $(MAKE) run

# Run backend only (requires IFACE parameter)
run-backend:
	@if [ -z "$(IFACE)" ]; then \
		echo "Error: IFACE not specified."; \
		echo "Usage: make run-backend IFACE=<network-interface>"; \
		echo ""; \
		echo "Example: make run-backend IFACE=eth0"; \
		exit 1; \
	fi
	@echo "Starting backend collector on interface: $(IFACE)..."
	cd backend && $(MAKE) run-collector IFACE=$(IFACE)

# Clean: Remove build artifacts from both
clean:
	@echo "Cleaning build artifacts..."
	cd frontend && $(MAKE) clean
	cd backend && $(MAKE) clean
	@echo "✓ Clean complete!"
