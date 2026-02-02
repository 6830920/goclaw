# Goclaw

A Go language implementation of the OpenClaw personal AI assistant framework.

## 🇺🇸 English Documentation (Default) | 🇨🇳 [中文文档](README-ZH.md)

## 🎯 Status: Full Stack Implementation!

### 🌐 Web Interface
Access the chat interface at `http://localhost:55789`  
Supports Progressive Web App (PWA) for mobile installation

### 🤖 AI Model Support
- **Minimax AI**: MiniMax-M2.1 model support
- **Qwen**: Qwen Coder model support
- **Configurable**: Easy setup for remote models
- **Fallback**: Built-in responses when no AI configured

### 🧠 Memory System
- **Short-term**: Conversation history management
- **Long-term**: Semantic search with vector embeddings
- **Working Memory**: Active task management

## Overview

This is a Go reimplementation of the original OpenClaw (https://github.com/openclaw/openclaw) project. It provides a personal AI assistant with:

- **Web Interface**: Modern UI accessible from any device
- **Mobile Ready**: PWA support for home screen installation
- **Memory System**: Short-term, long-term, and working memory with semantic search
- **AI Integration**: Configurable model providers (Minimax, Qwen, etc.)
- **API Backend**: RESTful API for programmatic access

## Features Implemented

| Feature | Status |
|---------|--------|
| Web Interface (PWA) | ✅ Working |
| Mobile Installation | ✅ Working |
| Minimax/Qwen Integration | ✅ Configurable |
| Memory System | ✅ Working |
| Vector Storage & Search | ✅ Working |
| Short-term Memory | ✅ Working |
| Long-term Memory | ✅ Working |
| Working Memory | ✅ Working |
| Chat Sessions | ✅ Working |
| REST API | ✅ Working |
| Configuration System | ✅ Working |
| Scheduled Tasks | ✅ Working |

## Quick Start

```bash
# Build
cd ~/projects/goclaw
./build.sh

# Configure (optional - for AI models)
cp config.example.json config.json
# Edit config.json with your API keys

# Run server
./bin/goclaw-server

# Access Web UI at http://localhost:55789
```

## API Endpoints (Port 55789)

- `GET /` - Web interface
- `GET /health` - Health check
- `POST /api/chat` - Chat with assistant
- `POST /api/memory/search` - Search memory
- `GET /api/memory/stats` - Memory statistics
- `GET /api/sessions` - List sessions
- `GET /api/cron/tasks` - List scheduled tasks
- `POST /api/cron/tasks` - Create new task
- `DELETE /api/cron/tasks/{id}` - Delete task
- `POST /api/cron/tasks/{id}/execute` - Execute task now

## Configuration

See [CONFIGURATION.md](CONFIGURATION.md) for detailed setup instructions.

### One-time Configuration Copy

To copy your existing OpenClaw configuration from `~/.openclaw/openclaw.json` to this project:

```bash
# Copy existing configuration (one-time operation)
cp ~/.openclaw/openclaw.json ~/projects/goclaw/config.json

# Or use the provided tool:
./bin/copy-config
```

### Supported AI Providers
The configuration supports multiple AI providers:
- **Minimax**: MiniMax-M2.1 model support
- **Qwen**: Qwen Coder and Vision models
- **Zhipu AI**: GLM-4 model support
- **Other providers**: Configurable via models.providers

Example configuration:
```json
{
  "models": {
    "providers": {
      "minimax": {
        "apiKey": "your_minimax_api_key",
        "baseUrl": "https://api.minimax.chat/v1"
      },
      "qwen-portal": {
        "apiKey": "your_qwen_api_key",
        "baseUrl": "https://portal.qwen.ai/v1"
      },
      "zhipu": {
        "apiKey": "your_zhipu_api_key",
        "model": "glm-4"
      }
    }
  }
}
```

## Project Structure

```
goclaw/
├── cmd/
│   └── server/            # HTTP API + Web UI server
├── internal/
│   ├── chat/              # Chat session management
│   ├── config/            # Configuration system
│   ├── core/              # Core types
│   ├── cron/              # Scheduled tasks system
│   ├── memory/            # Memory management
│   └── vector/            # Vector operations
├── pkg/
│   └── ai/                # AI model interfaces
├── static/                # Web UI files (generated)
├── bin/                   # Compiled binaries
├── config.example.json    # Example config
└── CONFIGURATION.md       # Setup guide
```

## Requirements

- Go 1.19+
- Web browser for UI access
- API keys for AI providers (optional)

## Documentation

- [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md) - Full development roadmap
- [CONFIGURATION.md](CONFIGURATION.md) - Setup guide
- [docs/project_structure.md](docs/project_structure.md) - Architecture details

## License

MIT License