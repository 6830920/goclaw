#!/bin/bash

# Goclaw 本地隧道脚本
# 用于快速将本地服务暴露到公网，便于预览

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🌐 Goclaw 本地隧道部署${NC}"
echo "========================"

# 检查依赖
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}❌ curl 未安装${NC}"
        exit 1
    fi
    
    # 检查是否有隧道工具
    if command -v ngrok &> /dev/null; then
        TUNNEL_TOOL="ngrok"
    elif command -v cloudflared &> /dev/null; then
        TUNNEL_TOOL="cloudflared"
    elif command -v frpc &> /dev/null; then
        TUNNEL_TOOL="frp"
    else
        echo -e "${YELLOW}⚠️  未检测到隧道工具${NC}"
        echo -e "${BLUE}💡 请先安装其中一个工具:${NC}"
        echo "   ngrok: https://ngrok.com/download"
        echo "   Cloudflare Tunnel: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/"
        echo "   FRP: https://github.com/fatedier/frp/releases"
        echo ""
        echo -e "${GREEN}📝 或者您可以使用在线服务:${NC}"
        echo "   1. Railway: https://railway.app"
        echo "   2. Render: https://render.com"
        echo "   3. Heroku: https://heroku.com"
        exit 0
    fi
}

# 启动 Goclaw 服务器
start_server() {
    echo -e "${BLUE}🚀 启动 Goclaw 服务器...${NC}"
    
    # 杀死之前的服务
    pkill -f goclaw-server 2>/dev/null || true
    
    # 启动服务器
    ./bin/goclaw-server &
    SERVER_PID=$!
    echo "Server PID: $SERVER_PID"
    
    # 等待服务器启动
    sleep 3
    
    # 检查服务器是否启动成功
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        echo -e "${RED}❌ 服务器启动失败${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 服务器已在端口 55789 启动${NC}"
}

# 使用 ngrok 创建隧道
use_ngrok() {
    echo -e "${BLUE}🔌 使用 ngrok 创建隧道...${NC}"
    
    if ! command -v ngrok &> /dev/null; then
        echo -e "${RED}❌ ngrok 未安装${NC}"
        return 1
    fi
    
    # 检查ngrok是否已认证
    if ! ngrok config check &> /dev/null; then
        echo -e "${YELLOW}💡 请先运行 'ngrok config add-authtoken YOUR_TOKEN'${NC}"
        echo "   获取token: https://dashboard.ngrok.com/get-started/your-authtoken"
    fi
    
    # 创建隧道
    echo -e "${BLUE}🌐 创建公共URL...${NC}"
    ngrok http 55789 &
    NGROK_PID=$!
    
    # 等待ngrok启动
    sleep 5
    
    # 获取公共URL
    PUBLIC_URL=$(curl -s http://localhost:4040/api/tunnels | python3 -c "import sys, json; print(json.load(sys.stdin)['tunnels'][0]['public_url'])" 2>/dev/null || echo "http://localhost:4040")
    
    if [ "$PUBLIC_URL" != "http://localhost:4040" ]; then
        echo -e "${GREEN}🎉 Goclaw 在线预览地址:${NC}"
        echo -e "${GREEN}🔗 $PUBLIC_URL${NC}"
        echo ""
        echo -e "${BLUE}📱 现在您可以通过该链接访问 Goclaw${NC}"
        echo -e "${BLUE}⚡ 实时同步本地更改${NC}"
    else
        echo -e "${YELLOW}⚠️  无法获取ngrok URL，检查 http://localhost:4040${NC}"
    fi
    
    # 等待用户终止
    echo -e "${YELLOW}🏃 按 Ctrl+C 停止服务${NC}"
    trap 'kill $NGROK_PID $SERVER_PID 2>/dev/null; exit' INT TERM
    wait $NGROK_PID $SERVER_PID 2>/dev/null
}

# 使用 Cloudflare Tunnel 创建隧道
use_cloudflare() {
    echo -e "${BLUE}🔌 使用 Cloudflare Tunnel 创建隧道...${NC}"
    
    if ! command -v cloudflared &> /dev/null; then
        echo -e "${RED}❌ cloudflared 未安装${NC}"
        return 1
    fi
    
    # 登录Cloudflare（首次需要）
    echo -e "${BLUE}🔑 检查 Cloudflare 登录状态...${NC}"
    
    # 创建隧道
    echo -e "${BLUE}🌐 创建隧道...${NC}"
    cloudflared tunnel --url http://localhost:55789 &
    CLOUDFLARE_PID=$!
    
    # 等待启动
    sleep 5
    
    # 获取隧道信息
    echo -e "${GREEN}💡 检查 Cloudflare 仪表板获取URL: https://dash.teams.cloudflare.com/warp-tunnel${NC}"
    echo -e "${GREEN}🔗 或查看终端输出获取公共URL${NC}"
    
    # 等待用户终止
    echo -e "${YELLOW}🏃 按 Ctrl+C 停止服务${NC}"
    trap 'kill $CLOUDFLARE_PID $SERVER_PID 2>/dev/null; exit' INT TERM
    wait $CLOUDFLARE_PID $SERVER_PID 2>/dev/null
}

# 使用 FRP 创建隧道
use_frp() {
    echo -e "${BLUE}🔌 使用 FRP 创建隧道...${NC}"
    
    if ! command -v frpc &> /dev/null; then
        echo -e "${RED}❌ frpc 未安装${NC}"
        return 1
    fi
    
    # 检查配置文件是否存在
    if [ ! -f "frpc.ini" ]; then
        echo -e "${YELLOW}⚠️  frpc.ini 配置文件不存在，正在创建示例配置...${NC}"
        
        cat > frpc.ini << 'EOF'
# Goclaw FRP 客户端配置文件
# 用于将本地 Goclaw 服务通过 FRP 暴露到公网

[common]
# 请替换为您的 FRP 服务器地址
server_addr = your-frp-server.com
server_port = 7000

# 如果服务器启用了 token 验证
# token = your-token-here

# 日志配置
log_file = ./frpc.log
log_level = info
log_max_days = 3

# Goclaw Web 服务
[goclaw-web]
type = http
local_ip = 127.0.0.1
local_port = 55789
# 自定义子域名（如果服务器支持）
# subdomain = goclaw
# 或者使用自定义域名
# custom_domains = your-domain.com

# 如果需要 TCP 端口转发
[goclaw-tcp]
type = tcp
local_ip = 127.0.0.1
local_port = 55789
# 远程端口（在FRP服务器上开放的端口）
# remote_port = 65789
EOF
        
        echo -e "${GREEN}✅ 已创建 frpc.ini 示例配置文件${NC}"
        echo -e "${YELLOW}💡 请编辑 frpc.ini 文件，填入您的 FRP 服务器信息${NC}"
        echo "   1. 修改 server_addr 为您的 FRP 服务器地址"
        echo "   2. 修改 server_port 为您的 FRP 服务器端口"
        echo "   3. 如需要，填入 token 验证信息"
        echo ""
        return 1
    fi
    
    # 启动 FRP 客户端
    echo -e "${BLUE}🌐 启动 FRP 客户端...${NC}"
    frpc -c frpc.ini &
    FRP_PID=$!
    
    # 等待启动
    sleep 3
    
    echo -e "${GREEN}✅ FRP 隧道已启动${NC}"
    echo -e "${BLUE}💡 请检查您的 FRP 服务器配置以获取访问地址${NC}"
    echo -e "${BLUE}⚡ Goclaw 服务现在可通过 FRP 隧道访问${NC}"
    
    # 等待用户终止
    echo -e "${YELLOW}🏃 按 Ctrl+C 停止服务${NC}"
    trap 'kill $FRP_PID $SERVER_PID 2>/dev/null; exit' INT TERM
    wait $FRP_PID $SERVER_PID 2>/dev/null
}

# 主函数
main() {
    check_dependencies
    
    # 构建服务器
    echo -e "${BLUE}🔨 构建 Goclaw 服务器...${NC}"
    export GOPROXY=https://goproxy.cn,direct
    go build -o bin/goclaw-server ./cmd/server
    
    # 启动服务器
    start_server
    
    # 询问使用哪种隧道工具
    echo ""
    echo -e "${BLUE}📋 选择隧道工具:${NC}"
    echo "1) ngrok"
    echo "2) Cloudflare Tunnel"
    echo "3) FRP (Fast Reverse Proxy)"
    echo "4) 仅启动本地服务"
    echo ""
    read -p "请选择 (1-4): " choice
    
    case $choice in
        1)
            use_ngrok
            ;;
        2)
            use_cloudflare
            ;;
        3)
            use_frp
            ;;
        4)
            echo -e "${GREEN}✅ 服务器已在 http://localhost:55789 运行${NC}"
            echo -e "${YELLOW}🏃 按 Ctrl+C 停止服务${NC}"
            trap 'kill $SERVER_PID 2>/dev/null; exit' INT TERM
            wait $SERVER_PID 2>/dev/null
            ;;
        *)
            echo -e "${RED}❌ 无效选择${NC}"
            ;;
    esac
}

# 运行主函数
main