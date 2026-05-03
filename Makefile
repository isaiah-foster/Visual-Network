.PHONY: help setup build run clean run-frontend run-backend run-proxy run-full setup-reverse-proxy build-reverse-proxy run-collector-remote run-proxy-remote

# Default target
help:
	@echo "Visual Network - Root Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make setup              - Install dependencies for frontend, backend, and reverse proxy"
	@echo "  make build              - Build frontend, backend, and reverse proxy"
	@echo "  make run                - Show instructions for running all services"
	@echo "  make run-frontend       - Start only the frontend dev server"
	@echo "  make run-backend        - Start only the backend collector (requires IFACE=<interface>)"
	@echo "  make run-proxy          - Start only the reverse proxy"
	@echo "  make run-full           - Summary for running all services in separate terminals"
	@echo "  make setup-reverse-proxy - Setup reverse proxy (generate certs, install deps)"
	@echo "  make build-reverse-proxy - Build only the reverse proxy"
	@echo "  make clean              - Clean build artifacts from all services"
	@echo ""
	@echo "Full Stack Setup (after cloning):"
	@echo "  make setup && make build && make run-full"
	@echo ""
	@echo "Examples:"
	@echo "  make setup              # Install all dependencies"
	@echo "  make build              # Build all services"
	@echo "  make run-backend IFACE=eth0  # Run backend collector on eth0"
	@echo "  make setup-reverse-proxy    # Setup reverse proxy with certs"

# Setup: Install dependencies for frontend, backend, and reverse proxy
setup:
	@echo "Setting up Visual Network (frontend + backend + reverse proxy)..."
	@echo ""
	@echo "=== Frontend Setup ==="
	cd frontend && $(MAKE) setup
	@echo ""
	@echo "=== Backend Setup ==="
	cd backend && $(MAKE) setup
	@echo ""
	@echo "=== Reverse Proxy Setup ==="
	cd rev-proxy && $(MAKE) setup
	@echo ""
	@echo "✓ Setup complete! Run 'make build' to build all services."

# Build: Build frontend, backend, and reverse proxy
build:
	@echo "Building Visual Network (frontend + backend + reverse proxy)..."
	@echo ""
	@echo "=== Frontend Build ==="
	cd frontend && $(MAKE) build
	@echo ""
	@echo "=== Backend Build ==="
	cd backend && $(MAKE) generate && $(MAKE) build
	@echo ""
	@echo "=== Reverse Proxy Build ==="
	cd rev-proxy && $(MAKE) build
	@echo ""
	@echo "✓ Build complete!"

# Run frontend and backend (instructions for running in separate terminals)
run:
	@echo "Visual Network - Start Services"
	@echo ""
	@echo "To run all services, open three terminals and run:"
	@echo ""
	@echo "  Terminal 1 (Frontend Dev Server):"
	@echo "    make run-frontend"
	@echo ""
	@echo "  Terminal 2 (Backend Collector):"
	@echo "    make run-backend IFACE=<your-network-interface>"
	@echo ""
	@echo "  Terminal 3 (Reverse Proxy & HTTPS):"
	@echo "    make run-proxy"
	@echo ""
	@echo "Example network interfaces: eth0, wlan0, en0 (macOS), etc."
	@echo ""
	@echo "To find your interface, run:"
	@echo "    ip link show   (Linux)"
	@echo "    ifconfig       (macOS/older systems)"
	@echo ""
	@echo "Then access:"
	@echo "  Frontend:  http://localhost:8080"
	@echo "  API Auth:  http://localhost:8080/api/auth"
	@echo "  Metrics WS: ws://localhost:8080/api/metrics/ws"

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

# Clean: Remove build artifacts from all services
clean:
	@echo "Cleaning build artifacts..."
	cd frontend && $(MAKE) clean
	cd backend && $(MAKE) clean
	cd rev-proxy && $(MAKE) clean
	@echo "✓ Clean complete!"


run-auth_service:
	@echo "Starting auth service..."
	cd backend/auth_service && $(MAKE) run

# Run reverse proxy only
run-proxy:
	@echo "Starting reverse proxy (HTTP on :8080, HTTPS on :8443)..."
	cd rev-proxy && $(MAKE) run-no-auth

# Setup reverse proxy with certificates
setup-reverse-proxy:
	@echo "Setting up reverse proxy with self-signed certificates..."
	cd rev-proxy && $(MAKE) setup
	cd rev-proxy && $(MAKE) generate-certs
	@echo "✓ Reverse proxy setup complete!"

# Build reverse proxy only
build-reverse-proxy:
	@echo "Building reverse proxy..."
	cd rev-proxy && $(MAKE) build
	@echo "✓ Reverse proxy build complete!"

# Full stack summary
run-full:
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                  Visual Network - Full Stack Setup                             ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "To run the complete Visual Network stack, open FOUR separate terminals:"
	@echo "NOTE: It is important to start the auth service before the frontend for proper connection to the database server."
	@echo ""
	@echo "┌─ Terminal 1: Frontend Dev Server ──────────────────────────────────────────────┐"
	@echo "│                                                                                │"
	@echo "│  $$ make run-frontend                                                          │"
	@echo "│                                                                                │"
	@echo "│  Access at: http://localhost:3000 (development)                               │"
	@echo "└────────────────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Terminal 2: Backend Collector ────────────────────────────────────────────────┐"
	@echo "│                                                                                │"
	@echo "│  $$ make run-backend IFACE=eth0    (replace eth0 with your interface)          │"
	@echo "│                                                                                │"
	@echo "│  Find your interface:                                                         │"
	@echo "│    Linux:  $$ ip link show                                                     │"
	@echo "│    macOS:  $$ ifconfig                                                         │"
	@echo "└────────────────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Terminal 3: Auth Service ─────────────────────────────────────────────────────┐"
	@echo "│                                                                                │"
	@echo "│  $$ make run-auth_service                                                      │"
	@echo "│                                                                                │"
	@echo "│  Requires MYSQL_DSN env var set                                                │"
	@echo "└────────────────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Terminal 4: Reverse Proxy ────────────────────────────────────────────────────┐"
	@echo "│                                                                                │"
	@echo "│  $$ make run-proxy                                                             │"
	@echo "│                                                                                │"
	@echo "│  Access at: http://localhost:8080                                             │"
	@echo "└────────────────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "Services will be available at:"
	@echo "  • Frontend:         http://localhost:8080"
	@echo "  • Auth API:         http://localhost:8080/api/auth"
	@echo "  • Backend API:      http://localhost:8080/api/backend"
	@echo "  • Metrics WebSocket: ws://localhost:8080/api/metrics/ws"
	@echo ""
	@echo "For HTTPS with the reverse proxy:"
	@echo "  $$ make setup-reverse-proxy"
	@echo "  Then access via: https://localhost:8443"
	@echo ""

# ---------------------------------------------------------------------------
# Split-host deployment (Issue #13)
#
# Host A (network monitor, requires root for eBPF):
#   make run-collector-remote IFACE=eth0 COLLECTOR_API_KEY=secret
#
# Host B (management host, no special privileges):
#   make run-proxy-remote COLLECTOR_HOST=<host-a-ip> COLLECTOR_API_KEY=secret
#
# The reverse proxy on Host B connects to the aggregator on Host A using the
# shared API key and forwards data to browser clients via WebSocket.
# ---------------------------------------------------------------------------

# Run collector in remote-accessible mode with optional API key protection.
run-collector-remote:
	@if [ -z "$(IFACE)" ]; then \
		echo "Error: IFACE not specified."; \
		echo "Usage: make run-collector-remote IFACE=eth0 [COLLECTOR_API_KEY=secret]"; \
		exit 1; \
	fi
	@echo "Starting collector (remote mode) on $(IFACE), API addr 0.0.0.0:9090..."
	BACKEND_API_ADDR=0.0.0.0:9090 COLLECTOR_API_KEY=$(COLLECTOR_API_KEY) \
		cd backend && $(MAKE) run-collector IFACE=$(IFACE)

# Run the reverse proxy pointing to a remote collector host.
COLLECTOR_HOST ?= localhost
run-proxy-remote:
	@echo "Starting reverse proxy -> collector at http://$(COLLECTOR_HOST):9090 ..."
	cd rev-proxy && $(MAKE) run-no-auth \
		EXTRA_FLAGS="-backend http://$(COLLECTOR_HOST):9090 -backend-api-key $(COLLECTOR_API_KEY)"
