#!/bin/bash

set -e

echo "🌐 开始部署前端应用到 Kubernetes..."

# 检查 minikube 是否运行
if ! minikube status >/dev/null 2>&1; then
    echo "❌ Minikube 未运行，请先启动 minikube"
    echo "💡 运行: minikube start"
    exit 1
fi

echo "✅ Minikube 状态正常"

# 检查 Node.js 是否安装
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js"
    echo "💡 访问: https://nodejs.org/"
    exit 1
fi

echo "✅ Node.js 版本: $(node --version)"

# 构建 Docker 镜像 (使用本地 Docker daemon)
echo "🔨 构建前端 Docker 镜像..."
cd "$(dirname "$0")/.."  # Go to script's parent directory (project root)

# 检查前端目录是否存在
if [ ! -d "web" ]; then
    echo "❌ 未找到前端目录 'web'"
    exit 1
fi

# 构建前端应用
echo "  - 构建前端应用..."
cd web
npm install
npm run build
cd ..

# 构建 Docker 镜像
docker build -t ys-cloud-frontend:latest -f Dockerfile.frontend .

# 设置 minikube docker 环境并推送镜像
echo "📦 推送镜像到 minikube..."
eval $(minikube docker-env)
minikube image load ys-cloud-frontend:latest

# 创建前端 Kubernetes 部署文件（如果不存在）
echo "📦 准备前端 Kubernetes 资源..."
cd "$(dirname "$0")"  # Go to script's directory (k8s)

# 创建前端部署文件
if [ ! -f "frontend-deployment.yaml" ]; then
    echo "  - 创建前端部署文件..."
    cat > frontend-deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ys-cloud-frontend
  namespace: ys-cloud
  labels:
    app: ys-cloud-frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ys-cloud-frontend
  template:
    metadata:
      labels:
        app: ys-cloud-frontend
    spec:
      containers:
      - name: frontend
        image: ys-cloud-frontend:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: ys-cloud-frontend-service
  namespace: ys-cloud
spec:
  selector:
    app: ys-cloud-frontend
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
  type: LoadBalancer
EOF
fi

# 部署前端应用
echo "  - 部署前端应用到 Kubernetes..."
kubectl apply -f frontend-deployment.yaml

# 等待前端应用就绪
echo "⏳ 等待前端应用就绪..."
kubectl wait --for=condition=ready pod -l app=ys-cloud-frontend -n ys-cloud --timeout=180s

# 获取服务信息
echo ""
echo "✅ 前端应用部署完成！"
echo ""
echo "📋 前端应用状态:"
kubectl get pods -l app=ys-cloud-frontend -n ys-cloud
kubectl get services -l app=ys-cloud-frontend -n ys-cloud

echo ""
echo "🌐 获取前端应用访问地址:"
minikube service ys-cloud-frontend-service -n ys-cloud --url

echo ""
echo "🔍 查看前端应用日志:"
echo "  kubectl logs -f deployment/ys-cloud-frontend -n ys-cloud"

echo ""
echo "💡 如需清理前端应用，运行: kubectl delete -f frontend-deployment.yaml"