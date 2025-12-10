#!/bin/bash

set -e

echo "🚀 开始部署 ys-cloud 后端服务到 Kubernetes..."

# 检查 minikube 是否运行
if ! minikube status >/dev/null 2>&1; then
    echo "❌ Minikube 未运行，请先启动 minikube"
    echo "💡 运行: minikube start"
    exit 1
fi

echo "✅ Minikube 状态正常"

# 构建 Docker 镜像 (使用本地 Docker daemon)
echo "🔨 构建 ys-cloud Docker 镜像..."
cd "$(dirname "$0")/.."  # Go to script's parent directory (project root)
docker build -t ys-cloud:latest .

# 设置 minikube docker 环境并推送镜像
echo "📦 推送镜像到 minikube..."
eval $(minikube docker-env)
minikube image load ys-cloud:latest

# 按顺序部署资源
echo "📦 部署 Kubernetes 资源..."
cd "$(dirname "$0")"  # Go to script's directory (k8s)

# 1. 创建命名空间
echo "  - 创建命名空间..."
kubectl apply -f namespace.yaml

# 2. 创建配置和密钥
echo "  - 创建配置映射..."
kubectl apply -f configmaps.yaml

echo "  - 创建密钥..."
kubectl apply -f secrets.yaml

# 3. 部署后端应用
echo "  - 部署 ys-cloud 后端服务..."
kubectl apply -f ys-cloud-app-deployment.yaml

# 等待应用就绪（增加超时时间）
echo "⏳ 等待后端服务就绪..."
kubectl wait --for=condition=ready pod -l app=ys-cloud-app -n ys-cloud --timeout=300s

# 获取服务信息
echo ""
echo "✅ 后端服务部署完成！"
echo ""
echo "📋 后端服务状态:"
kubectl get pods -l app=ys-cloud-app -n ys-cloud
kubectl get services -l app=ys-cloud-app -n ys-cloud

echo ""
echo "🌐 获取后端服务访问地址:"
minikube service ys-cloud-app-service -n ys-cloud --url

echo ""
echo "🔍 查看日志:"
echo "  kubectl logs -f deployment/ys-cloud-app -n ys-cloud"

echo ""
echo "📝 注意: 请确保数据库和Redis服务已先部署"
echo "   数据库部署: ./k8s/deploy-postgres.sh"
echo "   Redis部署:   ./k8s/deploy-redis.sh"

echo ""
echo "💡 如需清理，运行: ./k8s/cleanup.sh"