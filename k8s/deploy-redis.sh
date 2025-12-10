#!/bin/bash

set -e

echo "🔴 开始部署 Redis 缓存到 Kubernetes..."

# 检查 minikube 是否运行
if ! minikube status >/dev/null 2>&1; then
    echo "❌ Minikube 未运行，请先启动 minikube"
    echo "💡 运行: minikube start"
    exit 1
fi

echo "✅ Minikube 状态正常"

# 按顺序部署Redis资源
echo "📦 部署 Redis 资源..."
cd "$(dirname "$0")"  # Go to script's directory (k8s)

# 1. 创建命名空间
echo "  - 创建命名空间..."
kubectl apply -f namespace.yaml

# 2. 创建配置和密钥
echo "  - 创建配置映射..."
kubectl apply -f configmaps.yaml

echo "  - 创建密钥..."
kubectl apply -f secrets.yaml

# 3. 部署 Redis
echo "  - 部署 Redis..."
kubectl apply -f redis-deployment.yaml

# 等待 Redis 就绪
echo "⏳ 等待 Redis 缓存就绪..."
kubectl wait --for=condition=ready pod -l app=redis -n ys-cloud --timeout=60s

# 获取服务信息
echo ""
echo "✅ Redis 部署完成！"
echo ""
echo "📋 Redis 状态:"
kubectl get pods -l app=redis -n ys-cloud
kubectl get services -l app=redis -n ys-cloud

echo ""
echo "🔍 查看 Redis 日志:"
echo "  kubectl logs -f deployment/redis -n ys-cloud"

echo ""
echo "🔗 连接信息:"
echo "  主机: redis-service.ys-cloud.svc.cluster.local"
echo "  端口: 6379"
echo "  密码: redispass (在 secrets.yaml 中配置)"

echo ""
echo "🧪 测试 Redis 连接:"
echo "  kubectl exec -it deployment/redis -n ys-cloud -- redis-cli -a redispass ping"