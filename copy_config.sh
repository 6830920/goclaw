#!/bin/bash

# One-time configuration copy script for OpenClaw-Go

echo "OpenClaw-Go One-Time Configuration Copy"
echo "======================================"

# Build the config copy tool
echo "Building configuration copy tool..."
go build -o bin/copy-config ./cmd/copy-config

if [ $? -ne 0 ]; then
    echo "Error: Failed to build configuration copy tool"
    exit 1
fi

echo "Running configuration copy tool..."
./bin/copy-config

if [ $? -eq 0 ]; then
    echo ""
    echo "Configuration copied successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Review config.json to ensure settings are correct"
    echo "2. Start the server with: ./bin/openclaw-server"
    echo ""
else
    echo "Configuration copy failed."
    exit 1
fi