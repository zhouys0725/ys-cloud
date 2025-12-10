#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧹 开始清理 ys-cloud Kubernetes 资源...${NC}"

# 切换到脚本目录
cd "$(dirname "$0")"

# 检查命名空间是否存在
NAMESPACE_EXISTS=$(kubectl get namespace ys-cloud --no-headers 2>/dev/null | wc -l)
if [ "$NAMESPACE_EXISTS" -eq 0 ]; then
    echo -e "${YELLOW}⚠️  命名空间 ys-cloud 不存在，无需清理${NC}"
    exit 0
fi

echo -e "${BLUE}📋 当前 ys-cloud 命名空间中的资源:${NC}"
kubectl get all -n ys-cloud 2>/dev/null || echo "未找到资源"

# 解析命令行参数
FORCE_CLEANUP=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--force)
            FORCE_CLEANUP=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            echo "用法: $0 [选项]"
            echo "选项:"
            echo "  -f, --force    强制清理所有资源，包括 finalizers"
            echo "  -v, --verbose  显示详细输出"
            echo "  -h, --help     显示此帮助信息"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 -h 或 --help 查看帮助"
            exit 1
            ;;
    esac
done

# 清理函数
cleanup_resource() {
    local file=$1
    local resource_name=$2

    if [ -f "$file" ]; then
        echo -e "${YELLOW}🗑️  删除 ${resource_name}...${NC}"
        if [ "$VERBOSE" = true ]; then
            kubectl delete -f "$file" --ignore-not-found=true
        else
            kubectl delete -f "$file" --ignore-not-found=true > /dev/null 2>&1
        fi
    else
        echo -e "${YELLOW}⚠️  文件 $file 不存在，跳过${NC}"
    fi
}

# 按依赖关系顺序清理资源
echo -e "\n${BLUE}🔄 开始清理资源...${NC}"

# 1. 清理应用层资源 (会自动清理关联的 services)
cleanup_resource "frontend-deployment.yaml" "前端应用"
cleanup_resource "ys-cloud-app-deployment.yaml" "后端应用"

# 2. 清理数据库层资源
cleanup_resource "postgres-deployment.yaml" "PostgreSQL 数据库"
cleanup_resource "redis-deployment.yaml" "Redis 缓存"

# 3. 清理配置和密钥
cleanup_resource "configmaps.yaml" "配置映射"
cleanup_resource "secrets.yaml" "密钥"

# 4. 清理存储资源
cleanup_resource "storage.yaml" "存储资源"

# 5. 最后清理命名空间 (这会清理所有剩余资源)
echo -e "${YELLOW}🗑️  删除命名空间...${NC}"
kubectl delete -f namespace.yaml --ignore-not-found=true

# 等待资源清理
echo -e "\n${BLUE}⏳ 等待资源清理完成...${NC}"
sleep 5

# 强制清理选项
if [ "$FORCE_CLEANUP" = true ]; then
    echo -e "\n${RED}🔨 执行强制清理...${NC}"

    # 强制删除命名空间 (如果有 finalizers)
    if kubectl get namespace ys-cloud --no-headers 2>/dev/null | grep -q "Terminating"; then
        echo -e "${RED}⚡ 强制删除终止中的命名空间...${NC}"
        kubectl patch namespace ys-cloud -p '{"metadata":{"finalizers":[]}}' --type=merge
    fi

    # 清理可能的孤立资源
    echo -e "${RED}🧹 清理孤立资源...${NC}"
    kubectl delete all,pvc,configmap,secret -n ys-cloud --all --ignore-not-found=true > /dev/null 2>&1
fi

# 验证清理结果
echo -e "\n${BLUE}🔍 验证清理结果...${NC}"

# 检查命名空间是否已删除
NAMESPACE_REMAINS=$(kubectl get namespace ys-cloud --no-headers 2>/dev/null | wc -l)
if [ "$NAMESPACE_REMAINS" -eq 0 ]; then
    echo -e "${GREEN}✅ 命名空间 ys-cloud 已成功删除${NC}"
else
    echo -e "${RED}❌ 命名空间 ys-cloud 仍然存在${NC}"
    if kubectl get namespace ys-cloud --no-headers 2>/dev/null | grep -q "Terminating"; then
        echo -e "${YELLOW}⚠️  命名空间正在终止中，请稍等或使用 --force 强制清理${NC}"
    fi
fi

# 检查是否有剩余的 pod
REMAINING_PODS=$(kubectl get pods -n ys-cloud --no-headers 2>/dev/null | wc -l)
if [ "$REMAINING_PODS" -eq 0 ]; then
    echo -e "${GREEN}✅ 所有 Pod 已清理${NC}"
else
    echo -e "${YELLOW}⚠️  仍有 $REMAINING_PODS 个 Pod 存在${NC}"
fi

# 检查是否有剩余的 PVC
REMAINING_PVCS=$(kubectl get pvc -n ys-cloud --no-headers 2>/dev/null | wc -l)
if [ "$REMAINING_PVCS" -eq 0 ]; then
    echo -e "${GREEN}✅ 所有 PVC 已清理${NC}"
else
    echo -e "${YELLOW}⚠️  仍有 $REMAINING_PVCS 个 PVC 存在${NC}"
fi

echo -e "\n${GREEN}🎉 清理完成！${NC}"

# 显示清理统计
if [ "$VERBOSE" = true ]; then
    echo -e "\n${BLUE}📊 清理统计:${NC}"
    echo "- 清理模式: $([ "$FORCE_CLEANUP" = true ] && echo "强制" || echo "标准")"
    echo "- 命名空间状态: $([ "$NAMESPACE_REMAINS" -eq 0 ] && echo "已删除" || echo "仍存在")"
    echo "- 剩余 Pod: $REMAINING_PODS"
    echo "- 剩余 PVC: $REMAINING_PVCS"
fi