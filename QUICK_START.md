# 🚀 YS Cloud 快速开始指南

## 📋 目录说明

- `k8s/` - Kubernetes部署配置和脚本
- `web/` - 前端应用
- `internal/` - 后端核心代码
- `pkg/` - 包文件
- `cmd/` - 应用程序入口点
- `bin/` - 构建输出目录

## 🎯 快速部署

### 1. Kubernetes部署（推荐）

```bash
# 一键部署所有服务
./k8s/deploy-all.sh

# 或者分步部署
./k8s/deploy-postgres.sh    # 部署PostgreSQL数据库
./k8s/deploy-redis.sh       # 部署Redis缓存
./k8s/deploy.sh             # 部署后端服务
./k8s/deploy-frontend.sh    # 部署前端应用
```

### 2. 前置要求

- **Minikube** - 本地Kubernetes集群
- **Docker** - 镜像构建
- **Node.js** - 前端构建（v18+）

```bash
# 启动Minikube
minikube start

# 检查状态
minikube status
```

### 3. 访问应用

```bash
# 获取前端访问地址
minikube service ys-cloud-frontend-service -n ys-cloud --url

# 获取后端API地址
minikube service ys-cloud-app-service -n ys-cloud --url
```

## 🔧 管理命令

### 查看服务状态
```bash
kubectl get pods -n ys-cloud
kubectl get services -n ys-cloud
```

### 查看日志
```bash
# 后端服务日志
kubectl logs -f deployment/ys-cloud-app -n ys-cloud

# 前端应用日志
kubectl logs -f deployment/ys-cloud-frontend -n ys-cloud

# 数据库日志
kubectl logs -f deployment/postgres -n ys-cloud
```

### 清理环境
```bash
# 清理所有资源
./k8s/cleanup.sh
```

## 📚 详细文档

- [`k8s/README.md`](k8s/README.md) - 详细的Kubernetes部署指南
- [`README.md`](README.md) - 项目说明文档

## 🛠️ 开发模式

如果要进行开发，可以只部署依赖服务，本地运行应用：

```bash
# 只部署数据库和Redis
./k8s/deploy-postgres.sh
./k8s/deploy-redis.sh

# 本地运行后端
go run main.go

# 本地运行前端
cd web && npm install && npm start
```

## 🔍 故障排除

### 常见问题

1. **Minikube未启动**
   ```bash
   minikube start
   ```

2. **镜像拉取失败**
   ```bash
   eval $(minikube docker-env)
   ```

3. **服务无法访问**
   ```bash
   minikube service <service-name> -n ys-cloud --url
   ```

4. **Pod启动失败**
   ```bash
   kubectl describe pod <pod-name> -n ys-cloud
   kubectl logs pod/<pod-name> -n ys-cloud
   ```

## 🎉 部署验证

成功部署后，你应该能看到以下所有服务都在运行：
- ✅ PostgreSQL数据库 (1/1 Running)
- ✅ Redis缓存 (1/1 Running)
- ✅ 后端服务 (1/1 Running)
- ✅ 前端应用 (1/1 Running)