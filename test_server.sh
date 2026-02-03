#!/bin/bash

# Goclaw 测试服务器脚本
# 用于实时测试代码修改

set -e

echo "🔄 启动 Goclaw 测试系统..."

# 设置端口
PORT=55789  # Goclaw标准端口

# 函数：构建服务器
build_server() {
    echo "🔨 正在构建服务器..."
    export GOPROXY=https://goproxy.cn,direct
    go build -o bin/goclaw-test-server ./cmd/server
    echo "✅ 构建完成"
}

# 函数：启动测试服务器
start_test_server() {
    echo "🚀 启动测试服务器 (端口: $PORT)..."
    
    # 杀死之前的进程
    pkill -f goclaw-test-server 2>/dev/null || true
    
    # 启动服务器
    ./bin/goclaw-test-server &
    SERVER_PID=$!
    sleep 3  # 等待服务器启动
    
    # 检查服务器是否成功启动
    if kill -0 $SERVER_PID 2>/dev/null; then
        echo "✅ 测试服务器已在端口 $PORT 启动 (PID: $SERVER_PID)"
    else
        echo "❌ 测试服务器启动失败"
        return 1
    fi
}

# 函数：停止测试服务器
stop_test_server() {
    echo "🛑 停止测试服务器..."
    pkill -f goclaw-test-server 2>/dev/null || true
    echo "✅ 测试服务器已停止"
}

# 函数：运行单元测试
run_unit_tests() {
    echo "🧪 运行单元测试..."
    export GOPROXY=https://goproxy.cn,direct
    go test ./... -v
    echo "✅ 单元测试完成"
}

# 函数：运行API测试
run_api_tests() {
    echo "📡 测试API端点..."
    
    # 检查健康状态
    if curl -sf http://localhost:$PORT/health >/dev/null 2>&1; then
        echo "✅ 健康检查: 正常"
    else
        echo "❌ 健康检查: 失败"
    fi
    
    # 测试根路径
    if curl -sf http://localhost:$PORT/ >/dev/null 2>&1; then
        echo "✅ Web界面: 可访问"
    else
        echo "❌ Web界面: 不可访问"
    fi
    
    # 测试内存统计
    if curl -sf http://localhost:$PORT/api/memory/stats >/dev/null 2>&1; then
        echo "✅ 内存API: 正常"
    else
        echo "❌ 内存API: 失败"
    fi
    
    # 测试会话API
    if curl -sf http://localhost:$PORT/api/sessions >/dev/null 2>&1; then
        echo "✅ 会话API: 正常"
    else
        echo "❌ 会话API: 失败"
    fi
    
    # 测试定时任务API
    if curl -sf http://localhost:$PORT/api/cron/tasks >/dev/null 2>&1; then
        echo "✅ 定时任务API: 正常"
    else
        echo "❌ 定时任务API: 失败"
    fi
}

# 函数：重新加载并测试
reload_and_test() {
    echo "🔄 重新加载并测试..."
    
    # 停止当前服务器
    stop_test_server
    
    # 重新构建
    build_server
    
    # 重启服务器
    start_test_server
    
    # 运行API测试
    run_api_tests
}

# 主菜单
show_menu() {
    echo ""
    echo "📋 Goclaw 测试系统菜单:"
    echo "1) 构建服务器"
    echo "2) 启动测试服务器"
    echo "3) 停止测试服务器"
    echo "4) 运行单元测试"
    echo "5) 运行API测试"
    echo "6) 重新加载并测试 (构建+启动+测试)"
    echo "7) 完整测试流程 (构建+启动+单元测试+API测试)"
    echo "0) 退出"
    echo ""
    read -p "请选择操作 (0-7): " choice
    
    case $choice in
        1) build_server ;;
        2) start_test_server ;;
        3) stop_test_server ;;
        4) run_unit_tests ;;
        5) run_api_tests ;;
        6) reload_and_test ;;
        7) 
            build_server
            start_test_server
            run_unit_tests
            run_api_tests
            ;;
        0) 
            stop_test_server
            echo "👋 退出测试系统"
            exit 0
            ;;
        *)
            echo "❌ 无效选择，请重试"
            show_menu
            ;;
    esac
}

# 如果没有参数，显示菜单；否则执行指定命令
if [ $# -eq 0 ]; then
    show_menu
else
    case "$1" in
        "build") build_server ;;
        "start") start_test_server ;;
        "stop") stop_test_server ;;
        "test-unit") run_unit_tests ;;
        "test-api") run_api_tests ;;
        "reload") reload_and_test ;;
        "full-test") 
            build_server
            start_test_server
            run_unit_tests
            run_api_tests
            ;;
        *) 
            echo "用法: $0 [build|start|stop|test-unit|test-api|reload|full-test]"
            exit 1
            ;;
    esac
fi