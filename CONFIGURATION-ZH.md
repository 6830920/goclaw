# Goclaw 配置指南

## 配置文件

复制 `config.example.json` 到 `config.json` 并填写您的凭据：

```bash
cp config.example.json config.json
```

## 可用配置选项

### 基础设置
- `gateway.port`: API服务器端口（默认：55789）
- `gateway.bind`: 绑定地址（默认："127.0.0.1"）

### AI模型提供商

#### Minimax AI（推荐）
```json
{
  "minimax": {
    "apiKey": "您的_MINIMAX_API_KEY_在这里",
    "model": "MiniMax-M2.1",
    "baseUrl": "https://api.minimax.chat/v1"
  }
}
```

1. 从 [Minimax 控制台](https://www.minimaxi.com/) 获取您的API密钥
2. 将 `您的_MINIMAX_API_KEY_在这里` 替换为您的实际API密钥
3. 可选更改模型（默认：MiniMax-M2.1）

#### 通义千问（Qwen）
```json
{
  "qwen-portal": {
    "apiKey": "您的_通义_API_KEY_在这里",
    "model": "coder-model",
    "baseUrl": "https://portal.qwen.ai/v1"
  }
}
```

1. 从 [通义千问开发者平台](https://www.aliyun.com/product/dashscope) 获取API密钥
2. 替换API密钥并根据需要调整模型

#### 智谱AI（Zhipu AI）
```json
{
  "zhipu": {
    "apiKey": "您的_智谱_API_KEY_在这里",
    "model": "glm-4",
    "baseUrl": "https://open.bigmodel.cn/api/paas/v4/"
  }
}
```

1. 从 [智谱AI控制台](https://bigmodel.cn/) 获取API密钥
2. 替换API密钥并根据需要调整模型

## 记忆设置
- `agent.sandbox`: 安全沙箱设置
- `memory.shortTermMax`: 最大短期记忆保留数量
- `memory.workingMax`: 最大工作记忆项数
- `memory.similarityCut`: 长期记忆检索阈值

## 使用配置运行

```bash
# 创建您的配置
cp config.example.json config.json
# 在config.json中编辑您的API密钥

# 启动服务器
./bin/goclaw-server
```

服务器会自动加载存在的 `config.json` 文件。