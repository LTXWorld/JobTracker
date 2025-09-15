#!/bin/bash
# health-check.sh - JobView 健康检查脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 函数：打印成功消息
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# 函数：打印错误消息
error() {
    echo -e "${RED}✗ $1${NC}"
}

# 函数：打印警告消息
warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

echo "========================================="
echo "   JobView 系统健康检查"
echo "========================================="
echo ""

# 检查结果计数
TOTAL_CHECKS=0
PASSED_CHECKS=0

# 函数：执行健康检查
check() {
    local name=$1
    local command=$2
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    if eval "$command" > /dev/null 2>&1; then
        success "$name"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        error "$name"
        return 1
    fi
}

# 1. 检查Docker服务
echo "检查Docker服务..."
check "Docker守护进程" "docker version"
check "Docker Compose" "docker-compose version"
echo ""

# 2. 检查容器状态
echo "检查容器状态..."
check "数据库容器运行中" "docker ps | grep -q jobview-db"
check "后端容器运行中" "docker ps | grep -q jobview-backend"
check "前端容器运行中" "docker ps | grep -q jobview-frontend"
echo ""

# 3. 检查数据库连接
echo "检查数据库..."
check "PostgreSQL响应" "docker exec jobview-db pg_isready -U jobview"
echo ""

# 4. 检查后端API
echo "检查后端API..."
check "健康检查端点" "curl -f http://localhost:8010/api/auth/health"

# 检查认证端点
if curl -f -X POST http://localhost:8010/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"test"}' > /dev/null 2>&1; then
    success "认证端点响应"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    warning "认证端点响应（这是正常的，因为使用了测试凭据）"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
echo ""

# 5. 检查前端服务
echo "检查前端服务..."
check "前端首页" "curl -f http://localhost"
check "前端静态资源" "curl -f http://localhost/ | grep -q '<script'"
echo ""

# 6. 检查网络连通性
echo "检查容器网络..."
check "后端到数据库连接" "docker exec jobview-backend ping -c 1 postgres"
echo ""

# 7. 资源使用情况
echo "资源使用情况："
echo "----------------------------------------"
docker stats --no-stream jobview-db jobview-backend jobview-frontend || warning "无法获取资源统计"
echo ""

# 8. 磁盘使用情况
echo "磁盘使用情况："
echo "----------------------------------------"
docker system df
echo ""

# 9. 最近的日志
echo "最近的错误日志（如果有）："
echo "----------------------------------------"
docker-compose logs --tail=10 2>&1 | grep -i error || echo "没有发现错误日志"
echo ""

# 总结
echo "========================================="
echo "   健康检查结果"
echo "========================================="
echo ""

if [ $PASSED_CHECKS -eq $TOTAL_CHECKS ]; then
    success "所有检查通过 ($PASSED_CHECKS/$TOTAL_CHECKS)"
    echo ""
    echo "系统运行正常！"
    exit 0
else
    warning "部分检查未通过 ($PASSED_CHECKS/$TOTAL_CHECKS)"
    echo ""
    echo "请检查失败的项目并修复问题。"
    echo "查看详细日志："
    echo "  docker-compose logs [service_name]"
    exit 1
fi