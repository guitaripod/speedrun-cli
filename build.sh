#!/bin/bash

# Cross-platform build script for speedrun-cli

echo "🏗️  Building speedrun-cli for multiple platforms..."

# Create build directory
mkdir -p build

# Build for different platforms
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o build/speedrun-cli-linux-amd64 .

echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o build/speedrun-cli-macos-amd64 .

echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o build/speedrun-cli-macos-arm64 .

echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o build/speedrun-cli-windows-amd64.exe .

echo "✅ Build complete! Binaries available in build/ directory:"
ls -la build/

echo ""
echo "🚀 Installation instructions:"
echo "Copy the appropriate binary for your platform and add it to your PATH"
echo "Example for Linux: sudo cp build/speedrun-cli-linux-amd64 /usr/local/bin/speedrun-cli"