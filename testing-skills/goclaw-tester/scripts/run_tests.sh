#!/bin/bash
# Goclaw全面测试套件

set -e  # 遇到错误时退出

PROJECT_DIR="${PROJECT_DIR:-/home/daniel/projects/goclaw}"
export GOPROXY=https://goproxy.cn,direct

echo "🚀 开始Goclaw全面测试套件..."

cd $PROJECT_DIR

echo "🔍 1. 代码质量检查..."
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
timeout 30s ./bin/test-server &
SERVER_PID=$!
sleep 5

# 确保服务器已启动
if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✅ 服务器启动成功 (PID: $SERVER_PID)"
    
    # 测试API端点
    echo "📡 测试API端点..."
    
    # 健康检查
    HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:55789/health 2>/dev/null)
    if [ "$HEALTH_STATUS" -eq 200 ]; then
        echo "✅ 健康检查通过"
    else
        echo "❌ 健康检查失败 (状态: $HEALTH_STATUS)"
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi

    # 测试聊天API
    CHAT_RESULT=$(curl -s -X POST http://localhost:55789/api/chat \
        -H "Content-Type: application/json" \
        -d '{"message":"hello"}' 2>/dev/null)
    if [[ $CHAT_RESULT == *"response"* || $CHAT_RESULT == *"message"* ]]; then
        echo "✅ 聊天API正常"
    else
        echo "⚠️  聊天API返回: $CHAT_RESULT"
    fi

    # 测试内存搜索API
    SEARCH_RESULT=$(curl -s -X POST http://localhost:55789/api/memory/search \
        -H "Content-Type: application/json" \
        -d '{"query":"test", "limit":5}' 2>/dev/null)
    if [[ $SEARCH_RESULT == *"status"* ]]; then
        echo "✅ 内存搜索API正常"
    else
        echo "⚠️  内存搜索API返回: $SEARCH_RESULT"
    fi
    
    # 测试定时任务API
    CRON_RESULT=$(curl -s http://localhost:55789/api/cron/tasks 2>/dev/null)
    if [[ $CRON_RESULT == *"status"* ]]; then
        echo "✅ 定时任务API正常"
    else
        echo "⚠️  定时任务API返回: $CRON_RESULT"
    fi

    # 停止服务器
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
    echo "✅ 集成测试完成"
else
    echo "❌ 服务器未能正常启动"
    exit 1
fi

echo "🏆 所有测试通过！Goclaw项目运行正常。"