#!/bin/bash
# Goclaw 服务更新脚本
# 编译新版本并重启服务

set -e

echo "🔄 开始更新Goclaw服务..."

# 1. 编译新版本
echo "📦 编译Goclaw服务器..."
cd /home/daniel/projects/goclaw
go build -o /home/daniel/goclaw-server ./cmd/server/

# 2. 停止服务
echo "⏹️  停止Goclaw服务..."
sudo systemctl stop goclaw.service

# 3. 等待服务完全停止
sleep 2

# 4. 启动服务
echo "▶️  启动Goclaw服务..."
sudo systemctl start goclaw.service

# 5. 等待服务启动
sleep 3

# 6. 检查服务状态
echo "📊 检查服务状态..."
if sudo systemctl is-active --quiet goclaw.service; then
    echo "✅ Goclaw服务运行正常！"
    echo ""
    echo "📍 访问地址："
    echo "   - 本地: http://localhost:55789"
    echo "   - 外网: http://82.156.152.146:35789"
    echo ""
    echo "📋 查看日志："
    echo "   sudo journalctl -u goclaw -f"
    echo "   tail -f /var/log/goclaw.log"
else
    echo "❌ Goclaw服务启动失败！"
    echo "📋 错误日志："
    sudo journalctl -u goclaw --no-pager -n 20
    exit 1
fi
