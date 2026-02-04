# Goclaw 工具系统实现计划

## 目标
实现类似OpenClaw的工具调用系统，让AI能够调用外部工具和服务。

## 核心设计

### 1. 工具定义 (Tool Definition)
```go
type Tool struct {
    Name        string                 // 工具名称
    Description string                 // 工具描述（给AI看的）
    Parameters  map[string]Parameter   // 参数定义
    Execute     func(ctx context.Context, params map[string]interface{}) (interface{}, error) // 执行函数
}

type Parameter struct {
    Type        string      // 类型：string, number, boolean, array, object
    Description string      // 参数描述
    Required    bool        // 是否必填
    Default     interface{} // 默认值
}
```

### 2. 工具注册器 (Tool Registry)
```go
type ToolRegistry struct {
    tools map[string]*Tool
    mu    sync.RWMutex
}

// 注册工具
func (r *ToolRegistry) Register(tool *Tool)

// 获取工具
func (r *ToolRegistry) Get(name string) (*Tool, error)

// 列出所有工具
func (r *ToolRegistry) List() []*Tool

// 格式化为AI可用的工具描述
func (r *ToolRegistry) FormatForAI() string
```

### 3. 工具调用器 (Tool Executor)
```go
type ToolExecutor struct {
    registry *ToolRegistry
}

// 执行工具调用
func (e *ToolExecutor) Execute(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error)

// 解析AI的工具调用请求
func (e *ToolExecutor) ParseToolCall(aiResponse string) (string, map[string]interface{}, error)
```

## 实现步骤

### Phase 1: 基础框架 (当前)
1. ✅ 定义Tool、Parameter等核心数据结构
2. ✅ 实现ToolRegistry
3. ✅ 实现ToolExecutor
4. ✅ 添加单元测试

### Phase 2: 内置工具
- read: 读取文件
- write: 写入文件
- exec: 执行命令
- web_search: 网络搜索
- memory_search: 记忆搜索
- web_fetch: 获取网页内容

### Phase 3: AI集成
- 修改prompt生成逻辑，包含工具描述
- 解析AI的工具调用意图
- 处理工具调用结果
- 支持多轮工具调用

### Phase 4: 高级特性
- 工具调用超时控制
- 工具调用权限控制
- 工具调用日志记录
- 工具依赖管理

## 目录结构
```
internal/tools/
├── tool.go          # Tool定义和核心数据结构
├── registry.go      # 工具注册器
├── executor.go      # 工具调用器
├── builtin/         # 内置工具
│   ├── read.go
│   ├── write.go
│   ├── exec.go
│   └── ...
└── registry_test.go # 测试
```

## 开发状态
- [x] 核心数据结构设计
- [ ] ToolRegistry实现
- [ ] ToolExecutor实现
- [ ] 内置工具实现
- [ ] AI集成
- [ ] 测试完善
- [ ] 文档编写
