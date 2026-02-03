---
name: goclaw-tester
description: Comprehensive testing and development support for Goclaw Go project. Provides automated testing, build verification, error detection, and real-time feedback for code modifications. Use when developing, debugging, or verifying changes to Goclaw codebase.
---

# Goclaw 测试与开发技能

## 概述

该技能为 Goclaw Go 项目提供全面的测试和开发支持，包括自动化测试、构建验证、错误检测和代码修改的实时反馈。

## 何时使用此技能

当需要以下操作时使用此技能：
1. 修改 Goclaw 代码后进行测试
2. 验证构建是否成功
3. 检测代码错误
4. 运行单元测试
5. 实时监控代码修改的影响
6. 性能测试和基准测试

## 测试流程

### 1. 代码修改验证流程

当修改代码后，按以下顺序执行验证：

```bash
# 1. 语法检查
cd ~/projects/goclaw && go vet ./...

# 2. 构建测试
cd ~/projects/goclaw && go build -o bin/test-server ./cmd/server

# 3. 单元测试
cd ~/projects/goclaw && go test ./... -v

# 4. 启动服务器测试
cd ~/projects/goclaw && ./bin/test-server &
SERVER_PID=$!
sleep 2
kill $SERVER_PID

echo "✓ 所有基本测试通过"
```

### 2. 实时错误检测脚本

使用 `scripts/check_and_build.sh` 脚本来实时检测错误：

```bash
#!/bin/bash
# scripts/check_and_build.sh

PROJECT_DIR="$1"
if [ -z "$PROJECT_DIR" ]; then
    PROJECT_DIR="~/projects/goclaw"
fi

echo "🔍 检查项目: $PROJECT_DIR"

# 语法检查
echo "📝 运行 go vet..."
go vet $PROJECT_DIR/... 2>&1
VET_RESULT=$?
if [ $VET_RESULT -ne 0 ]; then
    echo "❌ go vet 发现问题"
    exit 1
else
    echo "✅ go vet 通过"
fi

# 导入检查
echo "📦 检查未使用导入..."
go vet -vettool=$(which shadow) $PROJECT_DIR/... 2>/dev/null || echo "继续 - shadow 工具可能未安装"

# 构建测试
echo "🔨 尝试构建..."
go build -o $PROJECT_DIR/bin/test-build-$$.tmp $PROJECT_DIR/cmd/server 2>&1
BUILD_RESULT=$?
if [ $BUILD_RESULT -ne 0 ]; then
    echo "❌ 构建失败"
    exit 1
else
    echo "✅ 构建成功"
    rm -f $PROJECT_DIR/bin/test-build-$$.tmp
fi

# 测试运行
echo "🧪 运行测试..."
go test $PROJECT_DIR/... -short 2>&1
TEST_RESULT=$?
if [ $TEST_RESULT -ne 0 ]; then
    echo "❌ 测试失败"
    exit 1
else
    echo "✅ 测试通过"
fi

echo "🎉 所有检查通过！"
```

### 3. 详细的测试命令

#### 单元测试
```bash
# 运行所有测试
cd ~/projects/goclaw && go test ./... -v

# 运行特定包测试
cd ~/projects/goclaw && go test ./internal/vector -v

# 运行性能测试
cd ~/projects/goclaw && go test ./internal/vector -bench=. -benchmem
```

#### 集成测试
```bash
# 构建服务器
cd ~/projects/goclaw && go build -o bin/goclaw-dev-server ./cmd/server

# 启动服务器进行集成测试
cd ~/projects/goclaw && ./bin/goclaw-dev-server &
SERVER_PID=$!
sleep 3

# API端点测试
curl -s http://localhost:55789/health
curl -s -X POST http://localhost:55789/api/chat -H "Content-Type: application/json" -d '{"message":"hello"}'
curl -s -X POST http://localhost:55789/api/memory/search -H "Content-Type: application/json" -d '{"query":"test"}'

# 清理
kill $SERVER_PID
```

## 自动化测试脚本

### scripts/run_tests.sh
```bash
#!/bin/bash
# 全面的Goclaw测试套件

set -e  # 遇到错误时退出

PROJECT_DIR="${PROJECT_DIR:-~/projects/goclaw}"
echo "🚀 开始Goclaw测试套件..."

echo "🔍 1. 代码检查..."
cd $PROJECT_DIR
go fmt ./...
go vet ./...
go mod tidy

echo "🔨 2. 构建测试..."
go build -o bin/test-server ./cmd/server
if [ $? -eq 0 ]; then
    echo "✅ 构建成功"
else
    echo "❌ 构建失败"
    exit 1
fi

echo "🧪 3. 单元测试..."
go test ./... -short
if [ $? -eq 0 ]; then
    echo "✅ 单元测试通过"
else
    echo "❌ 单元测试失败"
    exit 1
fi

echo "🔌 4. 集成测试..."
# 启动服务器
./bin/test-server &
SERVER_PID=$!
sleep 5

# 测试API端点
echo "📡 测试API端点..."
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:55789/health)
if [ "$HEALTH_STATUS" -eq 200 ]; then
    echo "✅ 健康检查通过"
else
    echo "❌ 健康检查失败 (状态: $HEALTH_STATUS)"
    kill $SERVER_PID
    exit 1
fi

# 测试内存搜索
SEARCH_RESULT=$(curl -s -X POST http://localhost:55789/api/memory/search \
    -H "Content-Type: application/json" \
    -d '{"query":"test", "limit":5}')
if [[ $SEARCH_RESULT == *"status"* ]]; then
    echo "✅ 内存搜索API正常"
else
    echo "❌ 内存搜索API异常"
    kill $SERVER_PID
    exit 1
fi

# 停止服务器
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null || true
echo "✅ 集成测试通过"

echo "🏆 所有测试通过！"
```

### scripts/dev_watch.sh
```bash
#!/bin/bash
# 开发监视脚本 - 监视文件变化并自动测试

PROJECT_DIR="${PROJECT_DIR:-~/projects/goclaw}"
WATCH_FILE="$1"

if [ -z "$WATCH_FILE" ]; then
    echo "用法: $0 <file_to_watch>"
    echo "例如: $0 internal/vector/store.go"
    exit 1
fi

echo "👀 监视文件: $WATCH_FILE"
echo "🔧 按 Ctrl+C 停止监视"

while true; do
    inotifywait -q -e modify $WATCH_FILE 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "🔄 检测到文件修改，正在测试..."
        
        # 运行快速检查
        cd $PROJECT_DIR
        go vet $(dirname $WATCH_FILE) 2>&1
        if [ $? -eq 0 ]; then
            echo "✅ 语法检查通过"
            
            # 尝试构建
            go build -o bin/watch-test.tmp ./cmd/server 2>/dev/null
            if [ $? -eq 0 ]; then
                echo "✅ 构建通过"
                rm -f bin/watch-test.tmp
            else
                echo "❌ 构建失败"
            fi
        else
            echo "❌ 语法检查失败"
        fi
        
        echo "⏳ 等待下次修改..."
    fi
done
```

## 常见错误检测

### 1. Go语法错误
- 未使用的导入
- 重复的变量声明
- 类型不匹配
- 接口实现错误

### 2. 构建错误
- 依赖缺失
- 版本冲突
- CGO相关错误

### 3. 运行时错误
- 空指针引用
- 数组越界
- 并发竞争

## 性能测试

### 基准测试脚本
```bash
# 运行向量存储基准测试
cd ~/projects/goclaw && go test ./internal/vector -bench=Benchmark -benchmem

# 运行API性能测试
ab -n 100 -c 10 http://localhost:55789/health
```

## 调试技巧

### 1. 调试构建错误
```bash
# 详细构建输出
go build -x -v ./cmd/server

# 检查依赖
go list -m all
go mod graph
```

### 2. 调试运行时错误
```bash
# 启用调试信息
go build -gcflags="-N -l" -o bin/debug-server ./cmd/server
dlv exec ./bin/debug-server
```

### 3. 内存分析
```bash
# 构建时启用竞态检测
go build -race -o bin/race-server ./cmd/server

# 运行并测试是否有竞态条件
./bin/race-server &
SERVER_PID=$!
# 运行一些测试...
kill $SERVER_PID
```

## 持续集成检查清单

每次代码修改后，执行以下检查：

- [ ] `go fmt ./...` - 代码格式化
- [ ] `go vet ./...` - 静态分析
- [ ] `go build ./...` - 构建测试
- [ ] `go test ./...` - 单元测试
- [ ] 服务器启动测试
- [ ] API端点功能测试
- [ ] 内存使用检查
- [ ] 错误处理验证

## 使用示例

### 示例1: 修改向量存储后测试
```bash
# 修改代码后...
cd ~/projects/goclaw
./testing-skills/goclaw-tester/scripts/check_and_build.sh

# 如果通过，运行完整测试
./testing-skills/goclaw-tester/scripts/run_tests.sh
```

### 示例2: 监视特定文件
```bash
# 监视向量存储文件的变化
./testing-skills/goclaw-tester/scripts/dev_watch.sh internal/vector/store.go
```

### 示例3: 运行性能测试
```bash
# 检查修改对性能的影响
go test ./internal/vector -bench=Benchmark -benchmem -count=3
```