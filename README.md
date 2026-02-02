# OpenClaw-Go

A Go language implementation of the OpenClaw personal AI assistant framework.

## Overview

This is a Go reimplementation of the original OpenClaw (https://github.com/openclaw/openclaw) project, which is written in Node.js. OpenClaw is a personal AI assistant that runs on your own devices and integrates with various communication channels like WhatsApp, Telegram, Slack, Discord, etc.

## Goals

- Recreate the core functionality of OpenClaw in Go
- Maintain compatibility with the original concepts and architecture
- Provide similar multi-channel support (WhatsApp, Telegram, Slack, Discord, etc.)
- Implement agent session management
- Support for tools and automation
- Cross-platform compatibility
- Leverage Go's performance and concurrency features

## Architecture

The Go version implements:

1. **Core Gateway** - WebSocket-based control plane for sessions, channels, tools, and events
2. **Channel Integration** - Support for various messaging platforms (WhatsApp, Telegram, Discord, etc.)
3. **Agent System** - Session-based AI assistant with tool access
4. **Tool Framework** - Built-in tools like browser control, file operations, etc.
5. **Configuration Management** - Flexible configuration system

## Current Progress

- ✅ Basic project structure created
- ✅ Core types and interfaces defined
- ✅ Configuration management system implemented
- ✅ Tool implementations (file operations, command execution)
- ✅ Message and session management system
- ✅ WebSocket server foundation
- ⏳ Dependency resolution and building (in progress)
- ⏳ Channel integrations (planned)
- ⏳ AI agent system (planned)

## Project Structure

- `cmd/` - Main application entry points
- `internal/` - Internal packages not meant for external use
- `pkg/` - Public packages that can be imported by other projects
- `docs/` - Documentation files

## Installation

First, ensure you have Go 1.19 or later installed:

```bash
go version
```

Then clone and build the project:

```bash
cd ~/projects/openclaw-go
go mod tidy
go build -o bin/openclaw ./cmd/openclaw
```

Or use the provided build script:

```bash
./build.sh
```

## Usage

To run the gateway server:

```bash
./bin/openclaw
```

The server will start on port 18789 by default (following OpenClaw convention).

## Configuration

An example configuration file is provided at `config.example.json`. Copy this to `config.json` and customize it with your API keys and settings.

## Contributing

Contributions are welcome! Please see the project structure documentation in `docs/project_structure.md` for details on how to extend the codebase.

## License

MIT License