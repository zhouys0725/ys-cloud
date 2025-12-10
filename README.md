# YS Cloud - 云平台自动化部署系统

YS Cloud 是一个现代化的云平台自动化部署系统，支持从 Git 仓库自动拉取代码、构建 Docker 镜像，并部署到 Kubernetes 集群中。

## 功能特性

- 🔧 **项目管理**: 支持多个项目的管理和配置
- 🚀 **流水线**: 可视化流水线配置，支持多阶段构建和部署
- 🐳 **Docker 集成**: 自动构建和推送 Docker 镜像
- ☸️ **Kubernetes 部署**: 支持多环境部署（开发、测试、生产）
- 🔄 **Webhook 触发**: 支持 GitHub、GitLab、Gitee 等 Git 平台的 Webhook
- 📊 **监控面板**: 实时查看构建和部署状态
- 👥 **用户管理**: 基于角色的访问控制
- 📝 **日志查看**: 完整的构建和部署日志记录

## 技术栈

### 后端
- **语言**: Go 1.24
- **框架**: Gin
- **数据库**: PostgreSQL
- **缓存**: Redis
- **容器**: Docker
- **编排**: Kubernetes

### 前端
- **语言**: TypeScript
- **框架**: React 18
- **UI 库**: Ant Design
- **状态管理**: React Hooks

## 快速开始

### 环境要求

- **Minikube** - 本地Kubernetes集群
- **Docker** - 镜像构建
- **Node.js 18+** - 前端构建
- **Go 1.24+** - 后端开发（可选）

### 部署

#### 1. 启动 Minikube
```bash
minikube start
minikube status
```

#### 2. 一键部署
```bash
# 部署所有服务（推荐）
./k8s/deploy-all.sh

# 或者分步部署
./k8s/deploy-postgres.sh    # 部署数据库
./k8s/deploy-redis.sh       # 部署缓存
./k8s/deploy.sh             # 部署后端服务
./k8s/deploy-frontend.sh    # 部署前端应用
```

#### 3. 访问应用
```bash
# 获取前端访问地址
minikube service ys-cloud-frontend-service -n ys-cloud --url

# 获取后端API地址
minikube service ys-cloud-app-service -n ys-cloud --url
```

### 验证部署
```bash
kubectl get pods -n ys-cloud
kubectl get services -n ys-cloud
```

成功部署后，你应该能看到以下所有服务都在运行：
- ✅ PostgreSQL数据库 (1/1 Running)
- ✅ Redis缓存 (1/1 Running)
- ✅ 后端服务 (1/1 Running)
- ✅ 前端应用 (1/1 Running)

## 🛠️ 使用指南

### 1. 创建项目

1. 登录系统后，点击"项目管理"
2. 点击"新建项目"
3. 填写项目信息：
   - 项目名称
   - 项目描述
   - Git 仓库地址
   - Git 提供商（GitHub/GitLab/Gitee）

### 2. 配置流水线

1. 进入项目详情，点击"流水线"
2. 点击"新建流水线"
3. 配置流水线步骤（YAML 格式）：

```yaml
version: '1.0'

stages:
  - name: build
    image: golang:1.24
    commands:
      - go mod download
      - go build -o app .
    artifacts:
      - path: ./app

  - name: docker
    image: docker:latest
    commands:
      - docker build -t your-app:\${BUILD_NUMBER} .
      - docker push your-app:\${BUILD_NUMBER}

  - name: deploy
    image: kubectl:latest
    commands:
      - kubectl apply -f k8s/
```

### 3. 部署应用

1. 运行流水线，系统会自动：
   - 拉取代码
   - 构建 Docker 镜像
   - 部署到 Kubernetes 集群

2. 在"部署管理"中查看部署状态和日志

### 4. 配置 Webhook

在 Git 平台中配置 Webhook，实现代码提交自动触发构建：

- GitHub: `http://your-domain.com/webhooks/github/{project-secret}`
- GitLab: `http://your-domain.com/webhooks/gitlab/{project-secret}`
- Gitee: `http://your-domain.com/webhooks/gitee/{project-secret}`

## 🔧 管理命令

### 查看服务状态
```bash
kubectl get pods -n ys-cloud
kubectl get services -n ys-cloud
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
# 连接PostgreSQL
kubectl exec -it deployment/postgres -n ys-cloud -- psql -U postgres -d ys_cloud

# 连接Redis
kubectl exec -it deployment/redis -n ys-cloud -- redis-cli -a redispass
```

### 清理环境
```bash
# 清理所有资源
./k8s/cleanup.sh
```

## 🏗️ 项目结构

```
ys-cloud/
├── k8s/                   # Kubernetes 部署配置
│   ├── deploy-*.sh        # 部署脚本
│   ├── *.yaml             # K8s 资源定义
│   └── README.md          # 部署详细文档
├── web/                   # 前端代码
│   ├── src/
│   └── public/
├── internal/              # 核心业务逻辑
│   ├── handler/           # HTTP 处理器
│   ├── service/           # 业务服务
│   └── model/             # 数据模型
├── pkg/                   # 包文件
├── cmd/                   # 应用程序入口点
├── bin/                   # 构建输出目录
├── Dockerfile             # 后端Docker文件
├── Dockerfile.frontend    # 前端Docker文件
├── main.go                # 主程序入口
└── entrypoint.sh          # 容器入口脚本
```

## 📚 详细文档

- [`k8s/README.md`](k8s/README.md) - 详细的Kubernetes部署指南
- [`QUICK_START.md`](QUICK_START.md) - 快速开始指南

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

### 监控指标

- 构建成功率
- 部署成功率
- 平均构建时间
- 系统资源使用情况

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📈 版本历史

### v1.0.0 (2024-01-01)
- 初始版本发布
- 基础项目管理功能
- 流水线配置和执行
- Docker 镜像构建
- Kubernetes 部署
- Web 前端界面
- 用户认证和权限管理

---

## 🙋‍♂️ 支持

如果您有任何问题或建议，请通过以下方式联系：

- 提交 Issue
- 发送邮件
- 参与讨论

**感谢使用 YS Cloud！** 🚀