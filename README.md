# OpenClaw-Go

A Go language implementation of the OpenClaw personal AI assistant framework.

## 🎯 Status: Full Stack Implementation!

### 🌐 Web Interface
Access the chat interface at `http://localhost:18889`  
Supports Progressive Web App (PWA) for mobile installation

### 🤖 AI Model Support
- **Zhipu AI**: GLM-4 model support via API key
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
- **AI Integration**: Configurable model providers (Zhipu AI, others)
- **API Backend**: RESTful API for programmatic access

## Features Implemented

| Feature | Status |
|---------|--------|
| Web Interface (PWA) | ✅ Working |
| Mobile Installation | ✅ Working |
| Zhipu AI Integration | ✅ Configurable |
| Memory System | ✅ Working |
| Vector Storage & Search | ✅ Working |
| Short-term Memory | ✅ Working |
| Long-term Memory | ✅ Working |
| Working Memory | ✅ Working |
| Chat Sessions | ✅ Working |
| REST API | ✅ Working |
| Configuration System | ✅ Working |

## Quick Start

```bash
# Build
cd ~/projects/openclaw-go
./build.sh

# Configure (optional - for AI models)
cp config.example.json config.json
# Edit config.json with your API keys

# Run server
./bin/openclaw-server

# Access Web UI at http://localhost:18889
```

## API Endpoints (Port 18889)

- `GET /` - Web interface
- `GET /health` - Health check
- `POST /api/chat` - Chat with assistant
- `POST /api/memory/search` - Search memory
- `GET /api/memory/stats` - Memory statistics
- `GET /api/sessions` - List sessions

## Configuration

See [CONFIGURATION.md](CONFIGURATION.md) for detailed setup instructions.

### Zhipu AI Setup
```json
{
  "zhipu": {
    "apiKey": "your_zhipu_api_key",
    "model": "glm-4"
  }
}
```

## Project Structure

```
openclaw-go/
├── cmd/
│   ├── openclaw/          # CLI version
│   └── server/            # HTTP API + Web UI server
├── internal/
│   ├── chat/              # Chat session management
│   ├── config/            # Configuration system
│   ├── core/              # Core types
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