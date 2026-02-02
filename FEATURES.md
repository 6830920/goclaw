# Goclaw 功能特性

## 🎯 核心功能

### 1. AI助手服务
- **多AI提供商支持**: 支持Minimax、通义千问、智谱AI等多种AI模型
- **实时对话**: 基于真实AI API的实时对话功能
- **智能响应**: 支持复杂查询的理解和响应

### 2. Web界面
- **中文界面**: 完整的中文用户界面
- **PWA支持**: 渐进式Web应用，支持移动设备安装
- **实时聊天**: Web端实时聊天界面
- **响应式设计**: 适配桌面和移动设备

### 3. 记忆系统
- **短期记忆**: 会话级别的对话历史管理
- **长期记忆**: 基于向量存储的知识库
- **语义搜索**: 基于嵌入向量的相似性搜索
- **上下文管理**: 智能上下文维护

### 4. 定时任务系统
- **Cron调度**: 基于cron表达式的任务调度
- **提醒功能**: 支持自定义提醒消息
- **通知系统**: 推送通知功能
- **任务管理**: 任务的CRUD操作

### 5. API服务
- **RESTful API**: 完整的REST API接口
- **聊天接口**: `/api/chat` 实时对话
- **记忆搜索**: `/api/memory/search` 语义搜索
- **任务管理**: `/api/cron/tasks` 任务调度
- **会话管理**: `/api/sessions` 会话控制

## 🚀 技术特性

### 架构特点
- **Go语言实现**: 高性能Go语言后端
- **模块化设计**: 清晰的模块分离
- **配置灵活**: 支持多种AI提供商配置
- **扩展性强**: 易于扩展新功能

### 端口约定
- **端口规则**: 采用55xxx端口系列（基于OpenClaw的18xxx映射）
- **当前端口**: 55789

### 安全特性
- **配置隔离**: API密钥等敏感信息自动排除
- **会话隔离**: 独立的会话管理
- **输入验证**: 请求参数验证

## 📋 API端点

### 聊天服务
- `POST /api/chat` - 发送消息并获取AI响应
- `GET /health` - 健康检查

### 记忆系统
- `POST /api/memory/search` - 搜索记忆库
- `GET /api/memory/stats` - 获取记忆统计

### 会话管理
- `GET /api/sessions` - 列出会话

### 定时任务
- `GET /api/cron/tasks` - 列出所有任务
- `POST /api/cron/tasks` - 创建新任务
- `GET /api/cron/tasks/{id}` - 获取特定任务
- `PUT /api/cron/tasks/{id}` - 更新任务
- `DELETE /api/cron/tasks/{id}` - 删除任务
- `POST /api/cron/tasks/{id}/execute` - 立即执行任务

## 🌐 使用场景

### 个人助手
- 日程管理
- 信息查询
- 学习辅助
- 生活提醒

### 知识管理
- 个人知识库
- 笔记整理
- 信息检索
- 学习资料管理

### 自动化任务
- 定时提醒
- 信息推送
- 数据处理
- 工作流自动化

## 📦 项目结构

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
├── static/                # Web UI文件（生成）
├── bin/                   # 编译后的二进制文件
├── config.example.json    # 示例配置
└── CONFIGURATION.md       # 设置指南
```

## 🚀 快速开始

### 构建
```bash
cd ~/projects/goclaw
./build.sh
```

### 配置
```bash
cp config.example.json config.json
# 编辑config.json以添加您的API密钥
```

### 运行
```bash
./bin/goclaw-server
```

访问 `http://localhost:55789`

## 🌍 本地化支持

### 语言支持
- **默认语言**: 中文
- **界面语言**: 中文Web界面
- **文档语言**: 中英双语文档

### 本地化特性
- 中文界面文本
- 中文提示信息
- 本土化功能设计