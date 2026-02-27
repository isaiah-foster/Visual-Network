.PHONY: all clean generate build run install-deps setup

# Default target
all: install-deps clean generate build

# Complete setup (first time only)
setup:
	@echo "Running initial setup..."
	./setup.sh

# Install Go dependencies
install-deps:
	@echo "Downloading Go modules..."
	go mod download
	@echo "Updating go.sum..."
	go mod tidy
	@echo "Installing bpf2go..."
	go install github.com/cilium/ebpf/cmd/bpf2go@latest

# Generate Go code from eBPF C code
generate: install-deps
	@echo "Generating eBPF Go bindings..."
	cd traffic-aggregation/ebpf && go generate

# Build the collector binary
build:
	go build -o bin/collector ./traffic-aggregation/main.go

# Run the collector (requires sudo and interface name)
# Usage: make run IFACE=<network interface>
run:
	@if [ -z "$(IFACE)" ]; then \
		echo "Error: IFACE not specified. Usage: make run IFACE=<your network interface>"; \
		exit 1; \
	fi
	sudo ./bin/collector $(IFACE)

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f traffic-aggregation/ebpf/tc_monitor_bpfel.go
	rm -f traffic-aggregation/ebpf/tc_monitor_bpfel.o
	rm -f traffic-aggregation/ebpf/tc_monitor_bpfeb.go
	rm -f traffic-aggregation/ebpf/tc_monitor_bpfeb.o

# Development: watch for changes and rebuild
watch:
	while true; do \
		make build; \
		inotifywait -qre close_write .; \
	done

# Check for required tools
check-tools:
	@command -v clang >/dev/null 2>&1 || { echo "Error: clang not found. Install with: apt-get install clang"; exit 1; }
	@command -v llvm-strip >/dev/null 2>&1 || { echo "Error: llvm-strip not found. Install with: apt-get install llvm"; exit 1; }
	@echo "All required tools are installed"

# Show help
help:
	@echo "Visual Network - eBPF Network Monitor"
	@echo ""
	@echo "Targets:"
	@echo "  all           - Generate eBPF code and build"
	@echo "  install-deps  - Install Go dependencies"
	@echo "  generate      - Generate Go code from eBPF C"
	@echo "  build         - Build the collector binary"
	@echo "  run           - Run the collector (requires IFACE=<interface>)"
	@echo "  clean         - Remove build artifacts"
	@echo "  check-tools   - Check if required tools are installed"
	@echo ""
	@echo "Example usage:"
	@echo "  make all"
	@echo "  make run IFACE=<your network interface>"
