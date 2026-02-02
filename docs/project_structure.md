# OpenClaw-Go Project Structure

This document describes the structure and components of the OpenClaw-Go project, which is a Go reimplementation of the original OpenClaw personal AI assistant framework.

## Directory Structure

```
openclaw-go/
├── cmd/
│   └── openclaw/
│       └── main.go              # Main application entry point
├── internal/
│   ├── core/
│   │   └── types.go            # Core types and interfaces
│   └── config/
│       └── config.go           # Configuration management
├── pkg/
│   ├── tools/
│   │   └── tools.go            # Utility tools (file operations, execution, etc.)
│   ├── messages/
│   │   └── messages.go         # Message and session management
│   └── [additional packages]   # Other reusable packages
├── docs/                       # Documentation files
├── config.example.json         # Example configuration file
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
└── README.md                   # Project overview
```

## Component Descriptions

### cmd/openclaw/main.go
The main entry point for the OpenClaw-Go application. This file sets up the WebSocket server that serves as the control plane, similar to the original OpenClaw gateway. It handles incoming connections and routes them appropriately.

### internal/core/types.go
Defines the core types used throughout the application:
- Session: Represents an AI assistant session
- Message: Represents a message within a session
- Agent: Manages AI interactions
- Channel: Represents communication channels (WhatsApp, Telegram, etc.)
- Gateway: The main control plane that coordinates sessions, channels, and agents

### internal/config/config.go
Handles configuration loading, saving, and validation. Includes structures for:
- Agent configuration
- Channel configurations
- Gateway settings
- Model configurations
- Authentication settings

### pkg/tools/tools.go
Provides essential utility functions that mirror the tooling in the original OpenClaw:
- FileReader: Reads files from the filesystem
- FileWriter: Writes content to files
- Executor: Executes shell commands securely
- FileEditor: Performs text replacements in files
- FileSystem: General filesystem operations

### pkg/messages/messages.go
Manages message and session handling:
- Message: Structure representing individual messages
- Session: Structure representing conversation sessions
- Manager: Handles creation, retrieval, and management of sessions and messages

## Planned Components

Additional packages that will be developed:

### pkg/channels/
Will contain implementations for various communication channels:
- Telegram integration
- WhatsApp integration (via Baileys or similar)
- Discord integration
- Slack integration
- Signal integration
- Other messaging platforms

### pkg/agents/
Will contain the AI agent logic:
- Model interaction
- Prompt management
- Response generation
- Tool calling capabilities

### pkg/websocket/
Will handle WebSocket communication protocols similar to the original OpenClaw gateway.

### pkg/cli/
Will provide command-line interface tools similar to the original OpenClaw CLI.

## Design Philosophy

The Go implementation aims to preserve the core concepts and functionality of the original OpenClaw while taking advantage of Go's strengths:

1. **Performance**: Go's efficient concurrency model for handling multiple sessions
2. **Reliability**: Strong typing and error handling to prevent runtime crashes
3. **Maintainability**: Clear separation of concerns and modular design
4. **Compatibility**: Maintaining API compatibility with existing OpenClaw concepts

## Building and Running

To build the project:
```bash
cd ~/projects/openclaw-go
go mod tidy
go build ./cmd/openclaw
```

To run the project:
```bash
./openclaw
```

Or directly with Go:
```bash
go run ./cmd/openclaw
```

## Next Steps

1. Implement channel integrations (starting with a simple test channel)
2. Develop the agent system with AI model connectivity
3. Add more sophisticated tool implementations
4. Implement proper session management and persistence
5. Add authentication and security features
6. Create CLI tools for management tasks