#!/bin/bash

# 一键启动脚本：启动 minikube 并自动开启 tunnel
# 使用方法: ./k8s/minikube-start.sh

echo "🚀 启动 minikube..."
minikube start

echo ""
echo "⏳ 等待 minikube 完全启动..."
sleep 5

# 检查 minikube 是否成功启动
if ! minikube status >/dev/null 2>&1; then
    echo "❌ Minikube 启动失败"
    exit 1
fi

echo "✓ Minikube 启动成功"
echo ""

# 启动 tunnel
echo "🌐 启动 minikube tunnel..."
nohup minikube tunnel > "$HOME/.minikube-tunnel.log" 2>&1 &
TUNNEL_PID=$!

# 保存 PID
echo $TUNNEL_PID > "$HOME/.minikube-tunnel.pid"

echo "⏳ 等待 tunnel 初始化..."
sleep 5

if ps -p $TUNNEL_PID > /dev/null 2>&1; then
    echo "✓ minikube tunnel 启动成功 (PID: $TUNNEL_PID)"
    echo ""
    echo "📝 可以使用以下服务:"
    echo "  MySQL:    127.0.0.1:3306 (root/my-password)"
    echo "  Redis:    127.0.0.1:6379 (使用 redis-cli 或 port-forward)"
    echo "  PostgreSQL: 127.0.0.1:5432 (使用 psql 或 port-forward)"
    echo ""
    echo "📊 查看所有服务:"
    kubectl get svc
else
    echo "❌ Tunnel 启动失败"
    rm -f "$HOME/.minikube-tunnel.pid"
    exit 1
fi
