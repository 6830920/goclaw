# Goclaw Development Plan

## 🇨🇳 [中文文档](DEVELOPMENT_PLAN-ZH.md) | 🇺🇸 English Documentation

## Phase 1: GitHub Code Management and Project Initialization

### 1.1 Create GitHub Repository
- Create `goclaw` repository on GitHub
- Set up SSH key authentication
- Create `.gitignore` file (standard Go project configuration)

### 1.2 Initialize Local Git Repository
```bash
cd ~/projects/goclaw
git init
git add .
git commit -m "Initial commit: core structure"
git remote add origin git@github.com:6830920/goclaw.git
git push -u origin main
```

### 1.3 Git Workflow
- **main branch**: Production code
- **develop branch**: Features under development
- **feature branches**: Individual feature modules
- Use GitHub Issues to track tasks
- Use Pull Requests to merge code

---

## Phase 2: Core Functional Architecture

### 2.1 Vectorization Engine
**Goal**: Implement text vectorization, support semantic search

**Implementation**:
```go
// internal/vector/embedding.go
type Embedding struct {
    Vector []float32
    Model  string
}

type Embedder interface {
    Embed(text string) ([]float32, error)
    EmbedBatch(texts []string) ([][]float32, error)
}
```

**Model Options**:
- Local Ollama (recommended, supports multiple embedding models)
- OpenAI API (compatible)
- Anthropic API (compatible)
- Local sentence-transformers

### 2.2 Memory Storage and Retrieval
**Goal**: Long-term and short-term memory storage, support semantic queries

**Architecture**:
```go
// internal/memory/memory.go
type MemoryStore struct {
    ShortTerm  *ConversationBuffer
    LongTerm   *VectorStore
    WorkingSet *WorkingMemory
}

type VectorStore struct {
    Embeddings []Embedding
    Metadata   []MemoryMetadata
    Index      vector.Index // Use faiss or hnswlib
}
```

**Storage Solutions**:
- Use ChromaDB (Go client available)
- Or use Qdrant (native Go, supports HTTP/gRPC)
- Or simple in-memory vector storage (initial phase)

### 2.3 Conversation System
**Goal**: Implement multi-turn conversations, context management

**Implementation**:
```go
// internal/chat/chat.go
type ChatManager struct {
    Sessions map[string]*ChatSession
    Config   ChatConfig
}

type ChatSession struct {
    ID          string
    Messages    []Message
    Memory      *MemoryStore
    SystemPrompt string
}
```

---

## Phase 3: Claude Code Integration

### 3.1 Claude Code Integration Plan
**Goal**: Use Claude Code as AI backend

**Configuration**:
```json
{
  "models": {
    "anthropic/claude-opus-4-5": {
      "apiKey": "local-or-env-key",
      "endpoint": "http://localhost:8080" // local API
    }
  }
}
```

### 3.2 Skill System
Based on OpenClaw's skill architecture:
- Skill definition (SKILL.md)
- Skill registry
- Skill installation and management
- Downloadable skill packages

**Implementation**:
```go
// pkg/skills/registry.go
type SkillRegistry struct {
    Skills map[string]Skill
}

type Skill struct {
    Name        string
    Description string
    Execute     func(ctx context.Context, input string) string
}
```

---

## Phase 4: Development Milestones

### Milestone 1: Basic Architecture (1 week) ✅
- [x] GitHub repository initialization
- [x] Project structure completion
- [x] Basic configuration system
- [x] Unit test coverage

### Milestone 2: Vectorization Engine (1 week) ✅
- [x] Embedding interface definition
- [x] Local Ollama integration
- [x] Simple vector storage
- [x] Basic similarity search

### Milestone 3: Memory System (1 week) ✅
- [x] Short-term memory (conversation buffer)
- [x] Long-term memory (vector storage)
- [x] Memory retrieval API
- [x] Integration with conversation system

### Milestone 4: Conversation Features (1 week) ✅
- [x] ChatManager implementation
- [x] Context management
- [x] Claude Code integration (partially completed, with fallback)
- [x] CLI interface

### Milestone 5: Web Interface (1 week) ✅
- [x] HTTP API server
- [x] Web interface implementation
- [x] PWA support (add to home screen)
- [x] Mobile adaptation

### Milestone 6: AI Model Integration (1 week) ✅
- [x] Minimax AI integration
- [x] Qwen integration
- [x] Zhipu AI integration
- [x] Configuration system
- [x] API key support
- [x] Model switching mechanism

### Milestone 7: Skill System (1 week)
- [ ] Skill registry
- [ ] Skill install/uninstall
- [ ] Basic skill implementations
- [ ] Skill marketplace integration

---

## Phase 5: Optional Features

### 5.1 Communication Channels
- WebSocket gateway
- Telegram Bot
- Discord Bot
- HTTP API

### 5.2 Tool Integration
- File operations
- Command execution
- Web search
- Browser control

### 5.3 Voice Support
- Voice recognition (TTS/STT)
- Voice conversations

---

## Technology Stack Summary

| Module | Technology Choice |
|--------|-------------------|
| Language | Go 1.19+ |
| Vector Database | In-memory / JSON persistence |
| Embedding | Ollama (local) |
| AI Models | Minimax, Qwen, Zhipu (cloud APIs) |
| Storage | File system |
| Configuration | JSON |
| Communication | HTTP API / PWA |

---

## Daily Development Process

1. **Morning**: Pull latest code, check CI status
2. **Development**: Create feature branch, implement functionality
3. **Testing**: Write unit tests, ensure passes
4. **Commit**: Commit code, write clear commit messages
5. **Code Review**: Create PR, self-review
6. **Merge**: Merge to develop branch

---

## Tools and Scripts

- `build.sh`: Build script
- `test.sh`: Run tests
- `copy_config.sh`: Configuration copy tool
- `run.sh`: Start service