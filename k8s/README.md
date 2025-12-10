# YS-Cloud Kubernetes 部署指南

本文档介绍了如何使用提供的脚本来部署 ys-cloud 系统到 Kubernetes 集群。

## 部署架构

系统现在采用独立部署的架构，分为以下组件：
- **PostgreSQL 数据库** - 持久化数据存储
- **Redis 缓存** - 缓存和会话存储
- **后端服务** - Go 语言后端应用
- **前端应用** - React 前端应用 (Nginx 静态服务)

## 部署脚本说明

### 1. 单独部署脚本

| 脚本文件 | 功能描述 |
|---------|---------|
| `deploy-postgres.sh` | 部署 PostgreSQL 数据库 |
| `deploy-redis.sh` | 部署 Redis 缓存 |
| `deploy.sh` | 部署后端服务 |
| `deploy-frontend.sh` | 部署前端应用 |
| `deploy-all.sh` | 按顺序部署所有组件 |

### 2. 一键完整部署

```bash
# 部署所有组件
./k8s/deploy-all.sh

# 跳过数据库和Redis部署（如果已存在）
./k8s/deploy-all.sh --skip-deps

# 跳过前端部署
./k8s/deploy-all.sh --skip-frontend

# 查看帮助
./k8s/deploy-all.sh --help
```

## 部署前准备

### 1. 启动 Minikube

```bash
minikube start
```

### 2. 检查环境

```bash
# 检查 minikube 状态
minikube status

# 检查 kubectl 连接
kubectl cluster-info
```

### 3. 安装依赖（前端构建需要）

```bash
# 检查 Node.js
node --version
npm --version

# 如果没有安装，请访问 https://nodejs.org/ 安装
```

## 部署步骤

### 方式一：一键部署（推荐）

```bash
./k8s/deploy-all.sh
```

### 方式二：分步部署

1. **部署数据库**
   ```bash
   ./k8s/deploy-postgres.sh
   ```

2. **部署 Redis**
   ```bash
   ./k8s/deploy-redis.sh
   ```

3. **部署后端服务**
   ```bash
   ./k8s/deploy.sh
   ```

4. **部署前端应用**
   ```bash
   ./k8s/deploy-frontend.sh
   ```

## 服务访问

部署完成后，可以通过以下方式访问服务：

### 获取访问地址

```bash
# 后端服务地址
minikube service ys-cloud-app-service -n ys-cloud --url

# 前端应用地址
minikube service ys-cloud-frontend-service -n ys-cloud --url
```

### 服务端点

- **前端应用**: `http://<minikube-url>`
- **后端API**: `http://<minikube-url>/api`
- **数据库**: `postgres-service.ys-cloud.svc.cluster.local:5432`
- **Redis**: `redis-service.ys-cloud.svc.cluster.local:6379`

## 管理和维护

### 查看服务状态

```bash
# 查看所有Pod
kubectl get pods -n ys-cloud

# 查看所有服务
kubectl get services -n ys-cloud

# 查看所有资源
kubectl get all -n ys-cloud
```

### 查看日志

```bash
# 后端服务日志
kubectl logs -f deployment/ys-cloud-app -n ys-cloud

# 前端应用日志
kubectl logs -f deployment/ys-cloud-frontend -n ys-cloud

# 数据库日志
kubectl logs -f deployment/postgres -n ys-cloud

# Redis日志
kubectl logs -f deployment/redis -n ys-cloud
```

### 数据库连接

```bash
# 连接到PostgreSQL
kubectl exec -it deployment/postgres -n ys-cloud -- psql -U postgres -d ys_cloud

# 连接到Redis
kubectl exec -it deployment/redis -n ys-cloud -- redis-cli -a redispass
```

### 重启服务

```bash
# 重启后端服务
kubectl rollout restart deployment/ys-cloud-app -n ys-cloud

# 重启前端服务
kubectl rollout restart deployment/ys-cloud-frontend -n ys-cloud

# 重启数据库
kubectl rollout restart deployment/postgres -n ys-cloud

# 重启Redis
kubectl rollout restart deployment/redis -n ys-cloud
```

## 清理资源

```bash
# 清理所有资源
./k8s/cleanup.sh

# 或者手动清理特定组件
kubectl delete -f k8s/frontend-deployment.yaml  # 前端
kubectl delete -f k8s/ys-cloud-app-deployment.yaml  # 后端
kubectl delete -f k8s/postgres-deployment.yaml  # 数据库
kubectl delete -f k8s/redis-deployment.yaml  # Redis
```

## 故障排除

### 常见问题

1. **Pod 启动失败**
   ```bash
   kubectl describe pod <pod-name> -n ys-cloud
   ```

2. **服务无法访问**
   ```bash
   kubectl get services -n ys-cloud
   kubectl describe service <service-name> -n ys-cloud
   ```

3. **镜像构建失败**
   ```bash
   # 检查 Docker 是否运行
   docker info

   # 检查 minikube docker 环境
   eval $(minikube docker-env)
   docker info
   ```

4. **前端构建失败**
   ```bash
   # 检查 Node.js 版本
   node --version
   npm --version

   # 手动构建测试
   cd web && npm install && npm run build
   ```

### 调试命令

```bash
# 进入Pod调试
kubectl exec -it <pod-name> -n ys-cloud -- /bin/sh

# 查看事件
kubectl get events -n ys-cloud --sort-by=.metadata.creationTimestamp

# 查看资源使用情况
kubectl top pods -n ys-cloud
```

## 配置说明

### 环境变量配置

主要的配置文件：
- `k8s/configmaps.yaml` - 配置映射
- `k8s/secrets.yaml` - 密钥信息

### 数据库配置

- **数据库名**: `ys_cloud`
- **用户名**: `postgres`
- **密码**: `password` (在 secrets.yaml 中)

### Redis配置

- **密码**: `redispass` (在 secrets.yaml 中)
- **端口**: 6379

### 前端配置

前端应用会自动代理 API 请求到后端服务，无需额外配置。

## 开发模式

如果需要在开发模式下运行：

```bash
# 启动数据库和Redis
./k8s/deploy-postgres.sh
./k8s/deploy-redis.sh

# 本地运行后端
cd .. && go run main.go

# 本地运行前端
cd web && npm start
```

这样可以实现数据库和Redis在Kubernetes中运行，而应用在本地开发。