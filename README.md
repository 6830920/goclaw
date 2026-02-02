# OpenClaw-Go

A Go language implementation of the OpenClaw personal AI assistant framework.

## 🎯 Status: Core Features Working!

```
OpenClaw-Go v0.1.0
==================

You: /help
Assistant: 
Commands:
  /new           - Start new session
  /quit          - Exit
  /remember <x>  - Save to memory
  /recall <x>    - Search memory
  /stats         - Show memory stats
  /help          - Show this help
```

## Overview

This is a Go reimplementation of the original OpenClaw (https://github.com/openclaw/openclaw) project. It provides a personal AI assistant with:

- **Memory System**: Short-term, long-term, and working memory with semantic search
- **Vector Embeddings**: Text embedding via local Ollama server
- **Chat Sessions**: Multi-session conversation management
- **CLI Interface**: Interactive command-line interface

## Features Implemented

| Feature | Status |
|---------|--------|
| Vector Embedding (Ollama) | ✅ Working |
| Vector Storage & Search | ✅ Working |
| Short-term Memory | ✅ Working |
| Long-term Memory | ✅ Working |
| Working Memory | ✅ Working |
| Chat Sessions | ✅ Working |
| CLI Interface | ✅ Working |
| Claude Code Integration | ⚠️ Fallback mode |

## Quick Start

```bash
# Build
cd ~/projects/openclaw-go
./build.sh

# Run
./bin/openclaw

# In another terminal, test memory
# (requires Ollama running with nomic-embed-text model)
```

## Project Structure

```
openclaw-go/
├── cmd/openclaw/           # Main application & CLI
├── internal/
│   ├── chat/              # Chat session management
│   ├── config/            # Configuration system
│   ├── core/              # Core types
│   ├── memory/            # Memory management
│   │   ├── buffer.go      # Short-term memory
│   │   ├── memory.go      # Main memory interface
│   │   ├── vector_memory.go # Long-term memory
│   │   └── working_memory.go # Working memory
│   └── vector/            # Vector operations
│       ├── embedding.go   # Ollama embedding client
│       └── store.go       # Vector storage
├── pkg/                   # Reusable packages
├── docs/                  # Documentation
├── bin/                   # Compiled binaries
└── config.example.json   # Example config
```

## CLI Commands

- `/new` - Start new session
- `/quit` - Exit
- `/remember <text>` - Save to memory
- `/recall <query>` - Search memory
- `/stats` - Show memory statistics
- `/help` - Show help

## Configuration

Copy `config.example.json` to `config.json` and configure:

```json
{
  "agent": {
    "model": "anthropic/claude-opus-4-5"
  }
}
```

## Requirements

- Go 1.19+
- Ollama (optional, for embeddings)
- Claude Code CLI (optional, for AI responses)

## Documentation

- [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md) - Full development roadmap
- [docs/project_structure.md](docs/project_structure.md) - Architecture details

## License

MIT License