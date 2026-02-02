# Goclaw

OpenClaw 个人AI助手框架的 Go 语言实现。

## 🇨🇳 中文文档 (默认)

## 🎯 状态：全栈实现完成！

### 🌐 Web界面
在 `http://localhost:55789` 访问聊天界面  
支持渐进式Web应用（PWA）以便移动设备安装

### 🤖 AI模型支持
- **Minimax AI**: MiniMax-M2.1 模型支持
- **通义千问**: 通义Coder模型支持
- **可配置**: 易于配置远程模型
- **降级**: 未配置AI时内置响应

### 🧠 记忆系统
- **短期记忆**: 对话历史管理
- **长期记忆**: 基于向量嵌入的语义搜索
- **工作记忆**: 活跃任务管理

## 概述

这是原始 OpenClaw (https://github.com/openclaw/openclaw) 项目的 Go 语言重构实现。它提供了具有以下功能的个人AI助手：

- **Web界面**: 可从任何设备访问的现代化UI
- **移动就绪**: 支持主屏幕安装的PWA
- **记忆系统**: 具有语义搜索功能的短期、长期和工作记忆
- **AI集成**: 可配置的模型提供商（Minimax、通义千问等）
- **API后端**: 用于程序访问的RESTful API
- **定时任务**: 支持cron表达式的任务调度系统

## 已实现功能

| 功能 | 状态 |
|---------|--------|
| Web界面 (PWA) | ✅ 已完成 |
| 移动端安装 | ✅ 已完成 |
| Minimax/通义千问集成 | ✅ 可配置 |
| 记忆系统 | ✅ 已完成 |
| 向量存储与搜索 | ✅ 已完成 |
| 短期记忆 | ✅ 已完成 |
| 长期记忆 | ✅ 已完成 |
| 工作记忆 | ✅ 已完成 |
| 聊天会话 | ✅ 已完成 |
| REST API | ✅ 已完成 |
| 配置系统 | ✅ 已完成 |
| 定时任务 | ✅ 已完成 |

## 快速开始

```bash
# 构建
cd ~/projects/goclaw
./build.sh

# 配置（可选 - 用于AI模型）
cp config.example.json config.json
# 在config.json中编辑您的API密钥

# 运行服务器
./bin/goclaw-server

# 在 http://localhost:55789 访问Web界面
```

## API端点 (端口 55789)

- `GET /` - Web界面
- `GET /health` - 健康检查
- `POST /api/chat` - 与助手聊天
- `POST /api/memory/search` - 搜索记忆
- `GET /api/memory/stats` - 记忆统计
- `GET /api/sessions` - 列出会话
- `GET /api/cron/tasks` - 列出定时任务
- `POST /api/cron/tasks` - 创建新任务
- `DELETE /api/cron/tasks/{id}` - 删除任务
- `POST /api/cron/tasks/{id}/execute` - 立即执行任务

## 配置

详细设置说明请参见 [CONFIGURATION.md](CONFIGURATION.md)。

### 一次性配置复制

要将您现有的OpenClaw配置从 `~/.openclaw/openclaw.json` 复制到此项目：

```bash
# 复制现有配置（一次性操作）
cp ~/.openclaw/openclaw.json ~/projects/goclaw/config.json

# 或使用提供的工具：
./bin/copy-config
```

### 支持的AI提供商
配置支持多种AI提供商：
- **Minimax**: MiniMax-M2.1 模型支持
- **通义千问**: 通义Coder和视觉模型
- **智谱AI**: GLM-4模型支持
- **其他提供商**: 可通过models.providers配置

示例配置：
```json
{
  "models": {
    "providers": {
      "minimax": {
        "apiKey": "您的minimax_api_key",
        "baseUrl": "https://api.minimax.chat/v1"
      },
      "qwen-portal": {
        "apiKey": "您的通义千问_api_key",
        "baseUrl": "https://portal.qwen.ai/v1"
      },
      "zhipu": {
        "apiKey": "您的智谱_api_key",
        "model": "glm-4"
      }
    }
  }
}
```

## 项目结构

```
goclaw/
├── cmd/
│   └── server/            # HTTP API + Web UI服务器
├── internal/
│   ├── chat/              # 聊天会话管理
│   ├── config/            # 配置系统
│   ├── core/              # 核心类型
│   ├── cron/              # 定时任务系统
│   ├── memory/            # 记忆管理
│   └── vector/            # 向量操作
├── pkg/
│   └── ai/                # AI模型接口
├── static/                # Web界面文件（生成）
├── bin/                   # 编译后的二进制文件
├── config.example.json    # 示例配置
├── FEATURES.md            # 功能特性文档
├── VECTOR_SEARCH.md       # 向量检索文档
├── ARCHITECTURE.md        # 系统架构文档
└── CONFIGURATION.md       # 设置指南
```

## 要求

- Go 1.19+
- 用于UI访问的Web浏览器
- AI提供商的API密钥（可选）

## 文档

- [FEATURES.md](FEATURES.md) - 功能特性详解
- [VECTOR_SEARCH.md](VECTOR_SEARCH.md) - 向量检索系统
- [ARCHITECTURE.md](ARCHITECTURE.md) - 系统架构设计
- [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md) - 完整开发路线图
- [CONFIGURATION.md](CONFIGURATION.md) - 设置指南
- [docs/project_structure.md](docs/project_structure.md) - 架构详情

## 测试系统

Goclaw包含全面的测试系统以确保代码质量和功能稳定性：

### 单元测试
```bash
# 运行所有单元测试
go test ./...

# 运行特定包的测试
go test ./internal/vector -v
go test ./internal/cron -v
```

### 集成测试
```bash
# 使用测试脚本运行完整测试流程
./test_server.sh full-test
```

### 测试覆盖范围
- 向量存储和检索功能
- 定时任务管理系统  
- API端点功能验证
- 并发访问安全性
- 数据持久化功能

详细测试信息请参见 [TESTING.md](TESTING.md)。

## 部署

### Docker 部署
```bash
# 构建并运行
docker build -t goclaw .
docker run -p 55789:55789 goclaw

# 或使用 Docker Compose
docker-compose up -d
```

### 一键部署脚本
```bash
# 查看部署选项
./deploy.sh

# 构建所有平台二进制文件
./deploy.sh build

# 构建 Docker 镜像
./deploy.sh docker
```

### 快速在线预览
```bash
# 启动本地服务并通过隧道公开访问
./local-tunnel.sh

# 然后可以选择:
# 1. 使用 ngrok 创建公共URL
# 2. 使用 Cloudflare Tunnel
# 3. 仅本地访问
```

### CI/CD
项目集成了 GitHub Actions 自动化流程：
- 代码提交后自动运行测试
- 自动构建跨平台二进制文件
- 自动创建 GitHub Releases

## 许可证

MIT 许可证

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