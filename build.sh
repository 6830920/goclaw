#!/bin/bash

# Build script for OpenClaw-Go

set -e  # Exit on any error

echo "OpenClaw-Go Build Script"
echo "========================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

echo "Go version: $(go version)"

# Set the project directory
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

echo "Project directory: $PROJECT_DIR"

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Build the project
echo "Building OpenClaw-Go..."
go build -o bin/openclaw ./cmd/openclaw

echo "Build completed successfully!"
echo "Binary created at: $PROJECT_DIR/bin/openclaw"

# Option to run the built binary
if [ "$1" == "--run" ]; then
    echo "Running OpenClaw-Go..."
    $PROJECT_DIR/bin/openclaw
fi

echo "To run the application manually:"
echo "  $PROJECT_DIR/bin/openclaw"