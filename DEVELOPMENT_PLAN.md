# Goclaw 开发计划

## 阶段一：GitHub代码管理和项目初始化

### 1.1 创建GitHub仓库
- 在GitHub上创建 `openclaw-go` 仓库
- 设置SSH密钥认证
- 创建 `.gitignore` 文件（Go项目标准配置）

### 1.2 初始化本地Git仓库
```bash
cd ~/projects/openclaw-go
git init
git add .
git commit -m "Initial commit: core structure"
git remote add origin git@github.com:yourusername/openclaw-go.git
git push -u origin main
```

### 1.3 Git工作流
- **main分支**: 生产代码
- **develop分支**: 开发中的功能
- **feature分支**: 各个功能模块
- 使用GitHub Issues跟踪任务
- 使用Pull Requests合并代码

---

## 阶段二：核心功能架构

### 2.1 向量化引擎 (Vectorization)
**目标**: 实现文本向量化，支持语义搜索

**实现方案**:
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

**可选模型**:
- 本地 Ollama（推荐，支持多种embedding模型）
- OpenAI API（兼容）
- Anthropic API（兼容）
- 本地sentence-transformers

### 2.2 记忆存储和检索 (Memory System)
**目标**: 长期和短期记忆存储，支持语义查询

**架构设计**:
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
    Index      vector.Index // 使用faiss或hnswlib
}
```

**存储方案**:
- 使用ChromaDB（Go客户端可用）
- 或使用Qdrant（原生Go，支持HTTP/gRPC）
- 或简单的内存向量存储（起步阶段）

### 2.3 对话系统 (Conversation)
**目标**: 实现多轮对话、上下文管理

**实现**:
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

## 阶段三：集成Claude Code

### 3.1 Claude Code集成方案
**目标**: 使用Claude Code作为AI后端

**配置**:
```json
{
  "models": {
    "anthropic/claude-opus-4-5": {
      "apiKey": "local-or-env-key",
      "endpoint": "http://localhost:8080" // 本地API
    }
  }
}
```

### 3.2 技能系统 (Skills)
参考OpenClaw的技能架构：
- 技能定义（SKILL.md）
- 技能注册表
- 技能安装和管理
- 可下载的技能包

**实现**:
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

## 阶段四：开发里程碑

### 里程碑 1: 基础架构（1周）✅
- [x] GitHub仓库初始化
- [x] 项目结构完善
- [x] 基础配置系统
- [ ] 单元测试覆盖 > 60%

### 里程碑 2: 向量化引擎（1周）✅
- [x] Embedding接口定义
- [x] 本地Ollama集成
- [x] 简单的向量存储
- [x] 基本的相似度搜索

### 里程碑 3: 记忆系统（1周）✅
- [x] 短期记忆（对话缓冲）
- [x] 长期记忆（向量存储）
- [x] 记忆检索API
- [x] 与对话系统集成

### 里程碑 4: 对话功能（1周）✅
- [x] ChatManager实现
- [x] 上下文管理
- [x] Claude Code集成（部分完成，有fallback）
- [x] CLI界面

### 里程碑 5: Web界面（1周）✅
- [x] HTTP API服务器
- [x] Web界面实现
- [x] PWA支持（添加到主屏幕）
- [x] 移动端适配

### 里程碑 6: AI模型集成（1周）✅
- [x] 智谱AI (Zhipu) 集成
- [x] 配置系统
- [x] API密钥支持
- [x] 模型切换机制

### 里程碑 7: 技能系统（1周）
- [ ] 技能注册表
- [ ] 技能安装/卸载
- [ ] 基础技能实现
- [ ] 技能市场集成

---

## 阶段五：可选功能

### 5.1 通信通道
- WebSocket网关
- Telegram Bot
- Discord Bot
- HTTP API

### 5.2 工具集成
- 文件操作
- 命令执行
- Web搜索
- 浏览器控制

### 5.3 语音支持
- 语音识别（TTS/STT）
- 语音对话

---

## 技术栈总结

| 模块 | 技术选择 |
|------|---------|
| 语言 | Go 1.19+ |
| 向量数据库 | Qdrant / ChromaDB |
| Embedding | Ollama（本地） |
| AI模型 | Claude Code（本地/云端） |
| 存储 | SQLite / 文件系统 |
| 配置 | JSON/YAML |
| 通信 | WebSocket / gRPC |

---

## 每日开发流程

1. **早上**: Pull最新代码，检查CI状态
2. **开发**: 创建feature分支，实现功能
3. **测试**: 编写单元测试，确保通过
4. **提交**: 提交代码，写清楚commit message
5. **Code Review**: 创建PR，自我审查
6. **合并**: 合并到develop分支

---

## 工具和脚本

- `build.sh`: 构建脚本
- `test.sh`: 运行测试
- `lint.sh`: 代码检查
- `run.sh`: 启动服务