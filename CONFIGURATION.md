# Goclaw Configuration Guide

## 🇨🇳 [中文文档](CONFIGURATION-ZH.md) | 🇺🇸 English Documentation

## Configuration File

Copy `config.example.json` to `config.json` and fill in your credentials:

```bash
cp config.example.json config.json
```

## Available Configuration Options

### Basic Settings
- `gateway.port`: Port for the API server (default: 55789)
- `gateway.bind`: Bind address (default: "127.0.0.1")

### AI Model Providers

#### Minimax AI (Recommended)
```json
{
  "minimax": {
    "apiKey": "YOUR_MINIMAX_API_KEY_HERE",
    "model": "MiniMax-M2.1",
    "baseUrl": "https://api.minimax.chat/v1"
  }
}
```

1. Get your API key from [Minimax Console](https://www.minimaxi.com/)
2. Replace `YOUR_MINIMAX_API_KEY_HERE` with your actual API key
3. Optionally change the model (default: MiniMax-M2.1)

#### Qwen (Tongyi Qianwen)
```json
{
  "qwen-portal": {
    "apiKey": "YOUR_QWEN_API_KEY_HERE",
    "model": "coder-model",
    "baseUrl": "https://portal.qwen.ai/v1"
  }
}
```

1. Get your API key from [Qwen Developer Platform](https://www.aliyun.com/product/dashscope)
2. Replace API key and adjust model as needed

#### Zhipu AI (GLM Series)
```json
{
  "zhipu": {
    "apiKey": "YOUR_ZHIPU_API_KEY_HERE",
    "model": "glm-4",
    "baseUrl": "https://open.bigmodel.cn/api/paas/v4/"
  }
}
```

1. Get your API key from [Zhipu AI Console](https://bigmodel.cn/)
2. Replace API key and adjust model as needed

## Memory Settings
- `agent.sandbox`: Security sandbox settings
- `memory.shortTermMax`: Max short-term memories to retain
- `memory.workingMax`: Max working memory items
- `memory.similarityCut`: Threshold for long-term memory retrieval

## Running with Configuration

```bash
# Create your config
cp config.example.json config.json
# Edit config.json with your API keys

# Start the server
./bin/goclaw-server
```

The server will automatically load `config.json` if it exists.