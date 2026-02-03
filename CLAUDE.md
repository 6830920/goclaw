# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此代码仓库中工作时提供指导。

## 项目概述

Goclaw 是一个基于 Go 语言的个人 AI 助手框架，是原始 OpenClaw 项目的重新实现。它提供完整的 AI 助手系统，具有 Web 界面、多层记忆管理、多提供商 AI 集成和定时任务功能。

**默认端口**: 55789（从 OpenClaw 原始端口 18888 映射而来）

## 构建和开发命令

### 构建
```bash
# 构建所有二进制文件（服务器和CLI）
./build.sh

# 构建特定组件
go build -o bin/goclaw-server ./cmd/server
go build -o bin/goclaw ./cmd/openclaw
```

### 测试
```bash
# 运行所有单元测试
go test ./... -v

# 运行特定包的测试
go test ./internal/vector -v
go test ./internal/cron -v

# 完整集成测试（构建 + 启动 + 测试）
./test_server.sh full-test

# 交互式测试菜单
./test_server.sh
```

### 开发工作流
```bash
# 快速开发循环
./test_server.sh build    # 仅构建
./test_server.sh start    # 启动测试服务器
./test_server.sh test-api # 测试 API 端点
./test_server.sh reload   # 构建 + 重启 + 测试
```

### 配置
```bash
# 复制现有 OpenClaw 配置（一次性设置）
cp ~/.openclaw/openclaw.json config.json

# 或使用提供的工具
./bin/copy-config
```

## 架构

### 多层架构
```
客户端层 (Web UI/CLI) → API 层 (REST) → 核心服务 → 存储层 → AI 集成
```

### 核心组件

**核心服务** (`internal/`):
- **chat/**: 会话管理和消息流转
- **memory/**: 三层记忆系统（短期记忆、长期记忆、工作记忆）
- **vector/**: 嵌入向量生成和语义搜索
- **cron/**: 支持 cron 表达式的任务调度
- **config/**: 多层配置（默认 → 本地 → 全局）

**AI 集成** (`pkg/ai/`):
- 多提供商抽象（Minimax、Qwen、Zhipu）
- 基于模型名称的提供商选择
- 自动降级和模拟响应的优雅处理

**入口点** (`cmd/`):
- `server/`: HTTP API + Web UI 服务器
- `openclaw/`: CLI 工具

### 配置系统

配置按优先级分层加载：**本地 → 全局 → 默认**

```json
{
  "models": {
    "providers": {
      "minimax": {
        "apiKey": "your_key",
        "baseUrl": "https://api.minimax.chat/v1"
      },
      "qwen-portal": {
        "apiKey": "your_key",
        "baseUrl": "https://portal.qwen.ai/v1"
      },
      "zhipu": {
        "apiKey": "your_key",
        "model": "glm-4"
      }
    }
  }
}
```

### 记忆架构

- **短期记忆**: 对话缓冲区（FIFO，50条消息）
- **长期记忆**: 基于向量嵌入的语义记忆
- **工作记忆**: 基于优先级的活跃任务

记忆上下文会自动提供给 AI 请求，以增强对话连续性。

### AI 提供商模式

系统使用提供商模式进行自动选择：
- 模型名称包含 "minimax" 或 "minimax-m2" → Minimax 提供商
- 模型名称包含 "qwen" 或 "coder-model" → Qwen 提供商
- 模型名称包含 "zhipu" 或 "glm" → Zhipu 提供商

所有提供商都实现 `Client` 接口，具有 `ChatCompletion(ctx, req)` 方法。

### 包约定

- **internal/**: 应用程序特定代码，不可在此项目外复用
- **pkg/**: 共享的可复用包，具有清晰接口
- **cmd/**: 应用程序入口点

## API 端点（端口 55789）

- `GET /` - Web 界面 (PWA)
- `GET /health` - 健康检查
- `POST /api/chat` - 与助手对话
- `POST /api/memory/search` - 搜索记忆
- `GET /api/memory/stats` - 记忆统计
- `GET /api/sessions` - 列出会话
- `GET /api/cron/tasks` - 列出定时任务
- `POST /api/cron/tasks` - 创建任务
- `DELETE /api/cron/tasks/{id}` - 删除任务
- `POST /api/cron/tasks/{id}/execute` - 立即执行任务

## 重要实现细节

### 模拟响应行为
当 AI 提供商无法访问或返回错误时，系统通过提供模拟响应优雅降级。这是出于演示/开发目的的故意设计。不要"修复"此行为 - 它允许系统在没有 API 密钥的情况下运行。

### Go 模块
- 模块名：`goclaw`
- Go 版本：1.19+
- 主要依赖：`github.com/gorilla/mux`、`github.com/robfig/cron/v3`

### 嵌入式 Web UI
Web 界面的 HTML/CSS/JS 在运行时生成，不是作为静态文件提供。查看 `cmd/server/` 中的模板生成代码。

### FRP 配置（内网）
对于需要内网穿透的部署，已预配置 FRP：
- 服务器：82.156.152.146:7000
- Token：wodefrpmima
- 端口：35789 (TCP) 或子域名 `goclaw` (HTTP)

## 常见模式

### 添加新的 AI 提供商
1. 在 `pkg/ai/client.go` 中创建实现 `Client` 接口的新客户端类型
2. 在 `MultiProviderClient.ChatCompletion()` 中添加提供商检测逻辑
3. 在 `internal/config/config.go` 中添加配置架构
4. 在服务器初始化中注册提供商

### 记忆操作
记忆操作通过 `internal/memory/` 中的记忆管理器进行。长期记忆会自动向量化并存储以进行语义搜索。

### 定时任务
任务由 `internal/cron/` 中的 cron 系统管理。任务支持标准 cron 表达式，可通过 API 按需执行。
