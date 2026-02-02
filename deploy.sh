#!/bin/bash

# Goclaw 自动部署脚本
# 用于自动化部署到各种托管平台

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Goclaw 部署脚本${NC}"
echo "==============================="

# 函数：打印带颜色的信息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 检查是否在项目根目录
if [ ! -f "go.mod" ]; then
    print_error "请在项目根目录运行此脚本"
    exit 1
fi

# 部署选项菜单
show_menu() {
    echo ""
    echo -e "${BLUE}📋 部署选项:${NC}"
    echo "1) 构建 Docker 镜像"
    echo "2) 部署到本地 Docker Compose"
    echo "3) 构建二进制文件"
    echo "4) 生成静态资源"
    echo "5) 部署到 GitHub Pages (前端资源)"
    echo "6) 运行所有构建步骤"
    echo "0) 退出"
    echo ""
    read -p "请选择部署选项 (0-6): " choice
    
    case $choice in
        1) build_docker ;;
        2) deploy_docker_compose ;;
        3) build_binaries ;;
        4) generate_static ;;
        5) deploy_github_pages ;;
        6) build_all ;;
        0) 
            echo "👋 退出部署脚本"
            exit 0
            ;;
        *)
            print_error "❌ 无效选择，请重试"
            show_menu
            ;;
    esac
}

# 构建 Docker 镜像
build_docker() {
    print_info "正在构建 Docker 镜像..."
    
    if command -v docker >/dev/null 2>&1; then
        docker build -t goclaw:latest .
        print_success "Docker 镜像构建完成"
    else
        print_error "Docker 未安装"
        return 1
    fi
}

# 部署到 Docker Compose
deploy_docker_compose() {
    print_info "正在部署到 Docker Compose..."
    
    if command -v docker-compose >/dev/null 2>&1; then
        docker-compose up -d
        print_success "Docker Compose 部署完成"
        echo "访问地址: http://localhost"
    else
        print_error "Docker Compose 未安装"
        return 1
    fi
}

# 构建二进制文件
build_binaries() {
    print_info "正在构建跨平台二进制文件..."
    
    mkdir -p dist
    
    # Linux AMD64
    print_info "构建 Linux AMD64..."
    GOOS=linux GOARCH=amd64 go build -o dist/goclaw-linux-amd64 -ldflags="-w -s" ./cmd/server
    
    # Linux ARM64
    print_info "构建 Linux ARM64..."
    GOOS=linux GOARCH=arm64 go build -o dist/goclaw-linux-arm64 -ldflags="-w -s" ./cmd/server
    
    # Windows AMD64
    print_info "构建 Windows AMD64..."
    GOOS=windows GOARCH=amd64 go build -o dist/goclaw-windows-amd64.exe -ldflags="-w -s" ./cmd/server
    
    # macOS AMD64
    print_info "构建 macOS AMD64..."
    GOOS=darwin GOARCH=amd64 go build -o dist/goclaw-darwin-amd64 -ldflags="-w -s" ./cmd/server
    
    # macOS ARM64
    print_info "构建 macOS ARM64..."
    GOOS=darwin GOARCH=arm64 go build -o dist/goclaw-darwin-arm64 -ldflags="-w -s" ./cmd/server
    
    print_success "所有平台二进制文件构建完成"
    ls -la dist/
}

# 生成静态资源
generate_static() {
    print_info "正在生成静态资源..."
    
    # 创建静态资源目录
    mkdir -p static
    
    # 创建基本的前端文件
    cat > static/index.html << 'EOF'
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Goclaw - 个人AI助手</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 10px;
            padding: 30px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
        }
        .status {
            text-align: center;
            margin: 20px 0;
        }
        .api-info {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 Goclaw - 个人AI助手</h1>
        
        <div class="status">
            <h2>✅ 服务运行正常</h2>
            <p>欢迎使用 Goclaw 个人AI助手框架</p>
        </div>
        
        <div class="api-info">
            <h3>🔗 API 接口</h3>
            <ul>
                <li><strong>聊天接口:</strong> <code>/api/chat</code></li>
                <li><strong>记忆搜索:</strong> <code>/api/memory/search</code></li>
                <li><strong>会话管理:</strong> <code>/api/sessions</code></li>
                <li><strong>定时任务:</strong> <code>/api/cron/tasks</code></li>
            </ul>
        </div>
        
        <div style="text-align: center; margin-top: 30px;">
            <p>Powered by Goclaw | Made with ❤️</p>
        </div>
    </div>
</body>
</html>
EOF

    print_success "静态资源生成完成"
}

# 部署到 GitHub Pages
deploy_github_pages() {
    print_info "准备部署到 GitHub Pages..."
    
    # 检查是否在正确的分支
    current_branch=$(git branch --show-current)
    if [ "$current_branch" != "main" ]; then
        print_warning "当前不在 main 分支，GitHub Pages 部署通常在 main 分支进行"
        read -p "是否继续? (y/N): " confirm
        if [[ ! $confirm =~ ^[Yy]$ ]]; then
            return 0
        fi
    fi
    
    # 生成静态资源
    generate_static
    
    # 创建 gh-pages 分支（如果不存在）
    git branch -D gh-pages 2>/dev/null || true
    git subtree push --prefix static origin gh-pages
    
    print_success "GitHub Pages 部署完成"
    print_info "访问地址: https://6830920.github.io/goclaw/"
}

# 运行所有构建步骤
build_all() {
    print_info "运行所有构建步骤..."
    
    build_binaries
    build_docker
    generate_static
    
    print_success "所有构建步骤完成"
}

# 如果没有参数，显示菜单；否则执行指定命令
if [ $# -eq 0 ]; then
    show_menu
else
    case "$1" in
        "docker") build_docker ;;
        "compose") deploy_docker_compose ;;
        "build") build_binaries ;;
        "static") generate_static ;;
        "pages") deploy_github_pages ;;
        "all") build_all ;;
        *) 
            echo "用法: $0 [docker|compose|build|static|pages|all]"
            exit 1
            ;;
    esac
fi