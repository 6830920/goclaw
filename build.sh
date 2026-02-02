#!/bin/bash

# Build script for Goclaw

set -e  # Exit on any error

echo "Goclaw Build Script"
echo "========================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

echo "Go version: $(go version)"

# Set the project directory
PROJECT_DIR="$(pwd)"

echo "Project directory: $PROJECT_DIR"

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Create bin directory
mkdir -p bin

# Build the CLI application
echo "Building CLI application..."
go build -o bin/goclaw ./cmd/openclaw

# Build the server application
echo "Building server application..."
go build -o bin/goclaw-server ./cmd/server

echo "Build completed successfully!"
echo "Binaries created at: $PROJECT_DIR/bin/"
ls -la $PROJECT_DIR/bin/

# Option to run the built binary
if [ "$1" == "--run" ]; then
    echo "Running Goclaw CLI..."
    $PROJECT_DIR/bin/goclaw
fi

if [ "$1" == "--run-server" ]; then
    echo "Running Goclaw server..."
    $PROJECT_DIR/bin/goclaw-server
fi

echo "To run the application manually:"
echo "  CLI: $PROJECT_DIR/bin/goclaw"
echo "  Server: $PROJECT_DIR/bin/goclaw-server"