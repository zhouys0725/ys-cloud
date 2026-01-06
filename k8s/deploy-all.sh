#!/bin/bash

set -e

echo "🎯 开始完整部署 ys-cloud 系统到 Kubernetes..."
echo "部署顺序: 数据库 -> Redis -> 后端服务 -> 前端应用"
echo ""

# 脚本目录
SCRIPT_DIR="$(dirname "$0")"

# 部署函数
deploy_component() {
    local component_name=$1
    local script_path=$2

    echo "=================================="
    echo "🚀 开始部署 $component_name..."
    echo "=================================="

    if [ -f "$script_path" ]; then
        bash "$script_path"
        echo ""
        echo "✅ $component_name 部署完成！"
        echo ""
    else
        echo "❌ 找不到部署脚本: $script_path"
        exit 1
    fi
}

# 等待用户确认
confirm_deploy() {
    local component=$1
    echo "准备部署 $component，按回车键继续..."
    read -r
}

# 开始部署
echo "🔍 检查部署前环境..."

# 检查 minikube 是否运行
if ! minikube status >/dev/null 2>&1; then
    echo "❌ Minikube 未运行，请先启动 minikube"
    echo "💡 运行: minikube start"
    exit 1
fi

echo "✅ Minikube 状态正常"
echo ""

# 可选择跳过某些组件
SKIP_DEPS=false
SKIP_FRONTEND=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        --skip-frontend)
            SKIP_FRONTEND=true
            shift
            ;;
        --help)
            echo "用法: $0 [选项]"
            echo "选项:"
            echo "  --skip-deps      跳过数据库和Redis部署"
            echo "  --skip-frontend  跳过前端部署"
            echo "  --help          显示此帮助信息"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 --help 查看可用选项"
            exit 1
            ;;
    esac
done

echo "📋 部署计划:"
if [ "$SKIP_DEPS" = false ]; then
    echo "  ✓ PostgreSQL 数据库"
    echo "  ✓ Redis 缓存"
else
    echo "  ✗ 跳过数据库和Redis (跳过依赖)"
fi
echo "  ✓ 后端服务"
if [ "$SKIP_FRONTEND" = false ]; then
    echo "  ✓ 前端应用"
else
    echo "  ✗ 跳过前端应用"
fi
echo ""

# 确认开始部署
echo "🎯 准备开始部署，按回车键继续..."
read -r

# 1. 部署数据库和Redis
if [ "$SKIP_DEPS" = false ]; then
    deploy_component "PostgreSQL 数据库" "$SCRIPT_DIR/deploy-postgres.sh"

    # 等待数据库完全就绪
    echo "⏳ 等待数据库完全初始化..."
    sleep 10

    deploy_component "Redis 缓存" "$SCRIPT_DIR/deploy-redis.sh"

    # 等待Redis完全就绪
    echo "⏳ 等待Redis完全初始化..."
    sleep 5
fi

# 2. 部署后端服务
deploy_component "后端服务" "$SCRIPT_DIR/deploy.sh"

# 3. 部署前端应用
if [ "$SKIP_FRONTEND" = false ]; then
    deploy_component "前端应用" "$SCRIPT_DIR/deploy-frontend.sh"
fi

# 部署完成
echo "=================================="
echo "🎉 ys-cloud 系统部署完成！"
echo "=================================="
echo ""
echo "📋 所有服务状态:"
kubectl get all -n ys-cloud

echo ""
echo "🌐 服务访问地址:"

# 获取各服务的访问地址
if [ "$SKIP_DEPS" = false ]; then
    echo ""
    echo "📊 PostgreSQL:"
    echo "  主机: postgres-service.default.svc.cluster.local:5432"
    echo ""
    echo "🔴 Redis:"
    echo "  主机: redis-service.default.svc.cluster.local:6379"
fi

echo ""
echo "🚀 后端服务:"
if command -v minikube &> /dev/null; then
    minikube service ys-cloud-app-service -n ys-cloud --url
fi

if [ "$SKIP_FRONTEND" = false ]; then
    echo ""
    echo "🌐 前端应用:"
    if command -v minikube &> /dev/null; then
        minikube service ys-cloud-frontend-service -n ys-cloud --url
    fi
fi

echo ""
echo "🔍 常用命令:"
echo "  查看所有Pod:      kubectl get pods -n ys-cloud"
echo "  查看所有服务:      kubectl get services -n ys-cloud"
echo "  查看后端日志:     kubectl logs -f deployment/ys-cloud-app -n ys-cloud"
if [ "$SKIP_FRONTEND" = false ]; then
    echo "  查看前端日志:     kubectl logs -f deployment/ys-cloud-frontend -n ys-cloud"
fi
echo ""
echo "  连接数据库:       kubectl exec -it deployment/postgres -n default -- psql -U postgres -d ys_cloud"
echo "  连接Redis:        kubectl exec -it deployment/redis -n default -- redis-cli -a redispass"

echo ""
echo "💡 清理所有服务:    ./k8s/cleanup.sh"
echo "💡 单独重新部署组件: ./k8s/deploy-xxx.sh"