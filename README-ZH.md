# Goclaw

OpenClaw 个人AI助手框架的 Go 语言实现。

## 🎯 状态：全栈实现完成！

### 🌐 Web 界面
在 `http://localhost:55789` 访问聊天界面  
支持渐进式 Web 应用（PWA）以便移动设备安装

### 🤖 AI 模型支持
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
│   ├── memory/            # 记忆管理
│   └── vector/            # 向量操作
├── pkg/
│   └── ai/                # AI模型接口
├── static/                # Web界面文件（生成）
├── bin/                   # 编译后的二进制文件
├── config.example.json    # 示例配置
└── CONFIGURATION.md       # 设置指南
```

## 要求

- Go 1.19+
- 用于UI访问的Web浏览器
- AI提供商的API密钥（可选）

## 文档

- [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md) - 完整开发路线图
- [CONFIGURATION.md](CONFIGURATION.md) - 设置指南
- [docs/project_structure.md](docs/project_structure.md) - 架构详情

## 许可证

MIT 许可证