#!/bin/bash

set -e

echo "🐶 开始部署 PostgreSQL 数据库到 Kubernetes..."

# 检查 minikube 是否运行
if ! minikube status >/dev/null 2>&1; then
    echo "❌ Minikube 未运行，请先启动 minikube"
    echo "💡 运行: minikube start"
    exit 1
fi

echo "✅ Minikube 状态正常"

# 按顺序部署数据库资源
echo "📦 部署 PostgreSQL 资源..."
cd "$(dirname "$0")"  # Go to script's directory (k8s)

# 1. 创建命名空间
echo "  - 创建命名空间..."
kubectl apply -f namespace.yaml

# 2. 创建存储
echo "  - 创建持久化存储..."
kubectl apply -f storage.yaml

# 3. 创建配置和密钥
echo "  - 创建配置映射..."
kubectl apply -f configmaps.yaml

echo "  - 创建密钥..."
kubectl apply -f secrets.yaml

# 4. 部署数据库
echo "  - 部署 PostgreSQL..."
kubectl apply -f postgres-deployment.yaml

# 等待数据库就绪
echo "⏳ 等待 PostgreSQL 数据库就绪..."
kubectl wait --for=condition=ready pod -l app=postgres -n default --timeout=120s

# 获取服务信息
echo ""
echo "✅ PostgreSQL 部署完成！"
echo ""
echo "📋 PostgreSQL 状态:"
kubectl get pods -l app=postgres -n default
kubectl get services -l app=postgres -n default

echo ""
echo "🔍 查看 PostgreSQL 日志:"
echo "  kubectl logs -f deployment/postgres -n default"

echo ""
echo "🔗 连接信息:"
echo "  主机: postgres-service.default.svc.cluster.local"
echo "  端口: 5432"
echo "  数据库: ys_cloud"
echo "  用户名: postgres"
echo "  密码: password (在 secrets.yaml 中配置)"