# Goclaw 向量检索系统

## 🧠 系统概述

Goclaw的向量检索系统基于OpenClaw的设计理念，提供高效的语义搜索和记忆管理功能。系统采用向量嵌入技术实现语义相似性匹配。

## 🔧 技术架构

### 向量存储
- **存储类型**: 内存存储（支持持久化到JSON文件）
- **数据结构**: 向量 + 元数据组合存储
- **索引方式**: 基于余弦相似度的相似性匹配

### 嵌入引擎
- **嵌入模型**: 通过API调用远程嵌入服务
- **支持格式**: Float32向量数组
- **维度处理**: 自适应向量维度匹配

## 🎯 核心功能

### 1. 向量存储
```go
// 添加向量条目
Add(ctx context.Context, vector []float32, metadata MemoryMetadata) (string, error)

// 批量添加带嵌入
AddWithEmbedding(ctx context.Context, content string, tags []string, custom map[string]string) (string, error)
```

### 2. 语义搜索
```go
// 向量搜索
Search(ctx context.Context, query []float32, limit int) ([]SearchResult, error)

// 文本搜索（自动生成嵌入）
SearchByText(ctx context.Context, query string, limit int) ([]SearchResult, error)
```

### 3. 数据管理
```go
// 获取单个项目
Get(ctx context.Context, id string) (*VectorEntry, error)

// 删除项目
Delete(ctx context.Context, id string) error

// 列表和计数
List(ctx context.Context, limit, offset int) ([]VectorEntry, error)
Count(ctx context.Context) (int, error)
```

## 📊 数据结构

### VectorEntry
```go
type VectorEntry struct {
    Vector   []float32      // 嵌入向量
    Metadata MemoryMetadata // 元数据
}
```

### MemoryMetadata
```go
type MemoryMetadata struct {
    ID        string            `json:"id"`        // 唯一标识符
    Content   string            `json:"content"`   // 内容文本
    Timestamp int64             `json:"timestamp"` // 时间戳
    Tags      []string          `json:"tags"`      // 标签数组
    Custom    map[string]string `json:"custom"`    // 自定义字段
}
```

### SearchResult
```go
type SearchResult struct {
    ID       string         // 匹配项ID
    Score    float32        // 相似度分数
    Content  string         // 内容文本
    Metadata MemoryMetadata // 元数据
}
```

## ⚡ 搜索算法

### 余弦相似度计算
系统使用标准的余弦相似度公式计算向量相似性：

```
similarity = (A · B) / (||A|| × ||B||)
```

其中：
- A · B 是向量点积
- ||A|| 和 ||B|| 是向量的模长

### 搜索流程
1. **查询嵌入**: 将查询文本转换为向量
2. **相似度计算**: 计算查询向量与存储向量的相似度
3. **排序**: 按相似度降序排列结果
4. **限制**: 返回前N个最佳匹配

## 💾 持久化

### JSON文件存储
- **格式**: JSON序列化格式
- **压缩**: Indented JSON便于阅读
- **备份**: 支持完整导出/导入

### 持久化接口
```go
// 保存到文件
Save(ctx context.Context, path string) error

// 从文件加载
Load(ctx context.Context, path string) error
```

## 🚀 性能特点

### 时间复杂度
- **搜索**: O(n)，其中n是向量库大小
- **插入**: O(1)
- **删除**: O(1)

### 内存使用
- **存储**: 向量 + 元数据
- **索引**: 无额外索引内存开销

## 🔍 使用场景

### 1. 对话记忆
- 存储历史对话内容
- 检索相关上下文
- 提供连续对话体验

### 2. 知识检索
- 存储知识片段
- 语义化搜索知识
- 提供智能问答

### 3. 个性化推荐
- 存储用户偏好
- 检索相似内容
- 提供个性化服务

## 🛠️ 配置选项

### 存储配置
- **持久化路径**: 可配置的JSON文件路径
- **最大容量**: 可限制存储数量
- **自动清理**: 可配置的过期策略

### 搜索配置
- **返回数量**: 搜索结果数量限制
- **相似度阈值**: 最低相似度要求
- **分页支持**: 支持分页查询

## 🔐 安全考虑

### 数据隐私
- 本地存储，不上传云端
- 加密存储（可选）
- 访问控制

### 性能安全
- 查询超时保护
- 内存使用限制
- 异常处理机制

## 📈 扩展性

### 存储后端
- 当前：内存存储 + JSON文件
- 未来：可扩展至SQLite、向量数据库

### 索引优化
- 当前：线性搜索
- 未来：近似最近邻搜索(ANN)

## 📋 API接口

### 内存搜索API
```
POST /api/memory/search - 语义搜索
GET  /api/memory/stats  - 内存统计
```

### 参数说明
- `query`: 搜索查询文本
- `limit`: 结果数量限制
- `tags`: 标签过滤条件

## 🎨 应用示例

### 智能问答
```javascript
// 用户问题: "昨天我们聊了什么?"
// 系统: 搜索记忆库中与"昨天"和"聊天"相关的条目
```

### 上下文感知
```javascript
// 根据对话历史提供个性化回复
// 检索相关话题的先前讨论
```

## 🔄 与其他系统集成

### 与AI模型集成
- 搜索结果作为上下文提供给AI
- 增强AI回复的相关性

### 与会话管理集成
- 自动保存对话到向量库
- 检索历史对话上下文

---
*该系统设计参考了OpenClaw的向量检索理念，实现了类似的语义搜索功能*