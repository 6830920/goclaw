# Goclaw 工具调用系统

## 🎯 概述

Goclaw提供了强大的工具调用能力，允许通过API执行系统命令、读写文件等操作。这些功能类似于OpenClaw的工具系统。

## 🛠️ 工具类型

### 1. 系统命令执行

执行任意系统命令：

```bash
# 执行 ls 命令
curl -X POST http://localhost:55789/api/tools/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": "ls",
    "args": ["-la", "/home/daniel/projects/goclaw"]
  }'

# 执行 Node.js 脚本
curl -X POST http://localhost:55789/api/tools/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": "node",
    "args": ["-e", "console.log(\"Hello from Goclaw!\")"]
  }'

# 查看系统信息
curl -X POST http://localhost:55789/api/tools/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": "uname",
    "args": ["-a"]
  }'
```

响应示例：
```json
{
  "stdout": "total 32\ndrwxr-xr-x  3 daniel daniel 4096 Feb  2 19:00 .\ndrwxr-xr-x  2 daniel daniel 4096 Feb  2 19:00 bin\ndrwxr-xr-x  2 daniel daniel 4096 Feb  2 19:00 cmd\n...",
  "stderr": "",
  "exitCode": 0,
  "duration": 0.005
}
```

### 2. 文件读取

读取文件内容：

```bash
# 读取 README
curl -X POST http://localhost:55789/api/tools/file/read \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "/home/daniel/projects/goclaw/README.md"
  }'

# 读取日志文件
curl -X POST http://localhost:55789/api/tools/file/read \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "/var/log/syslog",
    "tail": 100
  }'
```

响应示例：
```json
{
  "content": "# Goclaw\n\nOpenClaw 个人AI助手框架的 Go 语言实现。\n\n## 🎯 状态：全栈实现完成！\n..."
}
```

### 3. 文件写入

创建或覆盖文件：

```bash
# 创建新文件
curl -X POST http://localhost:55789/api/tools/file/write \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "/tmp/goclaw-test.txt",
    "content": "这是测试内容\n第二行\n第三行"
  }'

# 写入配置文件
curl -X POST http://localhost:55789/api/tools/file/write \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "/home/daniel/projects/goclaw/config.json",
    "content": "{\n  \"server\": {\n    \"port\": 55789\n  }\n}"
  }'
```

响应示例：
```json
{
  "success": true
}
```

## 📊 队列系统

### 消息队列

Goclaw使用消息队列来处理对话请求，确保高并发下的稳定性。

### 队列统计

```bash
# 获取队列状态
curl http://localhost:55789/api/queue/stats
```

响应示例：
```json
{
  "stats": {
    "queue_length": 5,
    "workers": 5,
    "capacity": 100
  }
}
```

### 会话管理

```bash
# 获取所有会话
curl http://localhost:55789/api/sessions

# 创建新会话
curl -X POST http://localhost:55789/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user123"
  }'

# 发送消息（自动加入队列）
curl -X POST http://localhost:55789/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，请帮我执行 ls 命令",
    "userId": "user123",
    "sessionId": "sess_123456_user123"
  }'
```

## 🔧 高级用法

### 1. 管道命令

```bash
# 组合命令
curl -X POST http://localhost:55789/api/tools/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": "bash",
    "args": ["-c", "ls -la | grep go | head -5"]
  }'
```

### 2. 读取并处理文件

```bash
# 读取文件并统计行数
curl -X POST http://localhost:55789/api/tools/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": "bash",
    "args": ["-c", "wc -l /home/daniel/projects/goclaw/README.md"]
  }'
```

### 3. 定时任务中的工具调用

通过cron任务执行系统命令：

```bash
# 创建定时任务执行备份
curl -X POST http://localhost:55789/api/cron/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "自动备份",
    "schedule": "0 2 * * *",
    "command": "backup",
    "payload": {
      "source": "/home/daniel/projects/goclaw/data",
      "destination": "/backup/goclaw"
    },
    "description": "每天凌晨2点自动备份数据"
  }'
```

## ⚠️ 安全注意事项

1. **命令注入风险**: 避免直接将用户输入拼接到命令中
2. **权限控制**: 限制可执行的命令范围
3. **文件路径验证**: 验证文件路径防止目录遍历攻击
4. **超时设置**: 设置命令执行超时防止无限运行
5. **日志记录**: 记录所有工具调用以便审计

### 建议的安全措施

```go
// 允许列表示例
allowedCommands := map[string]bool{
	"ls":    true,
	"cat":   true,
	"echo":  true,
	"date":  true,
	"whoami": true,
	// 只允许安全的命令
}

// 路径验证示例
func safePath(filename string) bool {
	// 防止目录遍历
	if filename == ".." || 
	   filename == "../" ||
	   contains(filename, "..") {
		return false
	}
	
	// 允许特定目录
	allowedDirs := []string{"/home/daniel/projects/goclaw", "/tmp"}
	for _, dir := range allowedDirs {
		if hasPrefix(filename, dir) {
			return true
		}
	}
	return false
}
```

## 📈 性能考虑

1. **命令超时**: 默认30秒超时
2. **并发限制**: 同时处理的消息数量有限制
3. **资源消耗**: 监控CPU和内存使用
4. **队列大小**: 队列容量100条消息

### 性能优化建议

- 使用轻量级命令
- 避免长时间运行的命令
- 使用流式输出处理大文件
- 定期清理临时文件

## 🔍 故障排除

### 常见问题

1. **命令执行失败**
   - 检查命令是否存在
   - 验证参数格式
   - 查看stderr输出

2. **文件访问被拒绝**
   - 检查文件权限
   - 验证路径是否正确
   - 确保有足够的访问权限

3. **超时错误**
   - 命令运行时间过长
   - 增加超时时间
   - 优化命令逻辑

### 调试命令

```bash
# 测试命令执行
curl -X POST http://localhost:55789/api/tools/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": "echo",
    "args": ["test"]
  }'

# 测试文件读取
curl -X POST http://localhost:55789/api/tools/file/read \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "/etc/hostname"
  }'
```

## 📚 API 参考

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/tools/exec` | POST | 执行系统命令 |
| `/api/tools/file/read` | POST | 读取文件 |
| `/api/tools/file/write` | POST | 写入文件 |
| `/api/queue/stats` | GET | 获取队列统计 |
| `/api/sessions` | GET/POST | 会话管理 |
| `/api/chat` | POST | 发送消息（自动队列） |