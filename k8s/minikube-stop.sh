#!/bin/bash

# 一键停止脚本：停止 minikube 和 tunnel

echo "⏹️  停止 minikube tunnel..."
PIDFILE="$HOME/.minikube-tunnel.pid"

if [ -f "$PIDFILE" ]; then
    PID=$(cat "$PIDFILE")
    if ps -p $PID > /dev/null 2>&1; then
        kill $PID
        echo "✓ Tunnel 已停止 (PID: $PID)"
    fi
    rm -f "$PIDFILE"
fi

# 停止所有 port-forward
ps aux | grep "kubectl port-forward" | grep -v grep | awk '{print $2}' | xargs kill -9 2>/dev/null

echo "⏹️  停止 minikube..."
minikube stop

echo "✓ 所有服务已停止"
