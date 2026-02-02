# OpenClaw-Go Configuration Guide

## Configuration File

Copy `config.example.json` to `config.json` and fill in your credentials:

```bash
cp config.example.json config.json
```

## Available Configuration Options

### Basic Settings
- `gateway.port`: Port for the API server (default: 18888)
- `gateway.bind`: Bind address (default: "127.0.0.1")

### AI Model Providers

#### Zhipu AI (Recommended)
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
2. Replace `YOUR_ZHIPU_API_KEY_HERE` with your actual API key
3. Optionally change the model (default: glm-4)

#### Other Providers (Coming Soon)
- Anthropic Claude
- OpenAI GPT
- Custom endpoints

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
./bin/openclaw-server
```

The server will automatically load `config.json` if it exists.