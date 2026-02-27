#!/bin/bash

# Setup script for Visual Network on Ubuntu
# Fixes header path issues

set -e

echo "=== Visual Network Setup ==="
echo ""

# Detect architecture
ARCH=$(uname -m)
echo "Detected architecture: $ARCH"

# Map architecture to kernel arch name
case $ARCH in
    x86_64)
        KERNEL_ARCH="x86"
        HEADER_ARCH="x86_64-linux-gnu"
        ;;
    aarch64)
        KERNEL_ARCH="arm64"
        HEADER_ARCH="aarch64-linux-gnu"
        ;;
    armv7l)
        KERNEL_ARCH="arm"
        HEADER_ARCH="arm-linux-gnueabihf"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "Kernel arch: $KERNEL_ARCH"
echo "Header path: /usr/include/$HEADER_ARCH"
echo ""

# Check for required packages
echo "Checking dependencies..."
MISSING_PKGS=""

if ! command -v clang &> /dev/null; then
    MISSING_PKGS="$MISSING_PKGS clang"1
fi

if ! command -v llvm-strip &> /dev/null; then
    MISSING_PKGS="$MISSING_PKGS llvm"
fi

if ! dpkg -l | grep -q libbpf-dev; then
    MISSING_PKGS="$MISSING_PKGS libbpf-dev"
fi

if ! dpkg -l | grep -q linux-headers-$(uname -r); then
    MISSING_PKGS="$MISSING_PKGS linux-headers-$(uname -r)"
fi

if [ -n "$MISSING_PKGS" ]; then
    echo "Missing packages:$MISSING_PKGS"
    echo ""
    echo "Install with:"
    echo "  sudo apt-get update"
    echo "  sudo apt-get install -y$MISSING_PKGS"
    echo ""
    read -p "Install now? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo apt-get update
        sudo apt-get install -y $MISSING_PKGS
    else
        echo "Please install missing packages manually."
        exit 1
    fi
fi

echo "✓ All dependencies installed"
echo ""

# Create asm symlink if it doesn't exist
if [ ! -e /usr/include/asm ]; then
    echo "Creating /usr/include/asm symlink..."
    if [ -d "/usr/include/$HEADER_ARCH/asm" ]; then
        sudo ln -sf "$HEADER_ARCH/asm" /usr/include/asm
        echo "✓ Created symlink: /usr/include/asm -> $HEADER_ARCH/asm"
    else
        echo "Warning: $HEADER_ARCH/asm not found"
    fi
else
    echo "✓ /usr/include/asm already exists"
fi

# Create asm-generic symlink if needed
if [ ! -e /usr/include/asm-generic ]; then
    echo "Creating /usr/include/asm-generic symlink..."
    if [ -d "/usr/include/$HEADER_ARCH/asm-generic" ]; then
        sudo ln -sf "$HEADER_ARCH/asm-generic" /usr/include/asm-generic
        echo "✓ Created symlink: /usr/include/asm-generic -> $HEADER_ARCH/asm-generic"
    fi
else
    echo "✓ /usr/include/asm-generic already exists"
fi

echo ""
echo "=== Setup Complete ==="
echo ""
echo "You can now build the project:"
echo "  make clean"
echo "  make generate"
echo "  make build"
echo ""
echo "Then run with:"
echo "  sudo make run IFACE=<your network interface>"
echo ""
