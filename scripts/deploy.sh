#!/bin/bash
# deploy.sh - JobView 部署脚本

set -e

echo "========================================="
echo "   JobView Docker 部署脚本"
echo "========================================="

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
    exit 1
}

# 函数：打印警告消息
warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        error "Docker未安装，请先安装Docker"
    fi
    success "Docker已安装"

    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose未安装，请先安装Docker Compose"
    fi
    success "Docker Compose已安装"
}

# 检查环境变量文件
check_env() {
    if [ ! -f .env ]; then
        warning ".env文件不存在，从.env.example创建..."
        cp .env.example .env
        warning "请编辑.env文件配置必要的环境变量后重新运行"
        exit 1
    fi
    success ".env文件已配置"
}

# 构建镜像
build_images() {
    echo ""
    echo "开始构建Docker镜像..."

    # 使用生产环境的docker-compose文件
    if [ -f docker-compose.production.yml ]; then
        docker-compose -f docker-compose.production.yml build
    else
        docker-compose build
    fi

    success "镜像构建完成"
}

# 启动服务
start_services() {
    echo ""
    echo "启动服务..."

    if [ -f docker-compose.production.yml ]; then
        docker-compose -f docker-compose.production.yml up -d
    else
        docker-compose up -d
    fi

    success "服务启动完成"
}

# 检查服务健康状态
check_health() {
    echo ""
    echo "检查服务健康状态..."

    # 等待服务启动
    sleep 10

    # 检查数据库
    if docker exec jobview-db pg_isready -U jobview > /dev/null 2>&1; then
        success "数据库服务正常"
    else
        warning "数据库服务可能还在启动中..."
    fi

    # 检查后端API
    if curl -f http://localhost:8010/api/auth/health > /dev/null 2>&1; then
        success "后端API服务正常"
    else
        warning "后端API服务可能还在启动中..."
    fi

    # 检查前端
    if curl -f http://localhost > /dev/null 2>&1; then
        success "前端服务正常"
    else
        warning "前端服务可能还在启动中..."
    fi
}

# 显示服务状态
show_status() {
    echo ""
    echo "服务状态："
    if [ -f docker-compose.production.yml ]; then
        docker-compose -f docker-compose.production.yml ps
    else
        docker-compose ps
    fi
}

# 显示访问信息
show_info() {
    echo ""
    echo "========================================="
    echo "   部署完成！"
    echo "========================================="
    echo ""
    echo "访问地址："
    echo "  - 前端: http://localhost"
    echo "  - 后端API: http://localhost:8010"
    echo ""
    echo "查看日志："
    echo "  docker-compose logs -f"
    echo ""
    echo "停止服务："
    echo "  docker-compose down"
    echo ""
}

# 主流程
main() {
    check_docker
    check_env
    build_images
    start_services
    check_health
    show_status
    show_info
}

# 运行主流程
main