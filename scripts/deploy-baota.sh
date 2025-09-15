#!/bin/bash
# deploy-baota.sh - JobView 宝塔面板部署脚本

set -e

echo "========================================="
echo "   JobView 宝塔面板部署脚本"
echo "========================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置变量
WEB_ROOT="/www/wwwroot/jobview.bfsmlt.top"
PROJECT_DIR="/opt/jobview"

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

# 函数：打印信息
info() {
    echo -e "→ $1"
}

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    error "请使用root用户运行此脚本"
fi

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        error "Docker未安装，请先安装Docker"
    fi
    success "Docker已安装"

    if ! command -v docker-compose &> /dev/null; then
        warning "Docker Compose未安装，正在安装..."
        curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
        success "Docker Compose安装完成"
    else
        success "Docker Compose已安装"
    fi
}

# 检查环境变量文件
check_env() {
    cd "$PROJECT_DIR"

    if [ ! -f .env ]; then
        warning ".env文件不存在，从.env.example创建..."
        cp .env.example .env
        warning "请编辑.env文件配置必要的环境变量"

        # 提示必须修改的配置
        echo ""
        echo "必须修改的配置项："
        echo "  1. DB_PASSWORD - 数据库密码"
        echo "  2. JWT_SECRET - JWT密钥（至少32字符）"
        echo ""
        echo "编辑命令：nano $PROJECT_DIR/.env"
        echo ""
        read -p "是否现在编辑配置文件？[y/N]: " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            nano .env
        else
            error "请先配置.env文件后再运行部署"
        fi
    fi
    success ".env文件已配置"
}

# 构建前端
build_frontend() {
    info "构建前端应用..."

    cd "$PROJECT_DIR/frontend"

    # 检查是否安装了Node.js
    if ! command -v node &> /dev/null; then
        warning "Node.js未安装，正在安装..."
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
    fi

    # 安装依赖
    info "安装前端依赖..."
    npm ci

    # 构建生产版本
    info "构建生产版本..."
    VITE_API_BASE=/api npm run build

    success "前端构建完成"
}

# 部署前端到宝塔网站目录
deploy_frontend() {
    info "部署前端文件到宝塔网站目录..."

    # 创建网站目录（如果不存在）
    mkdir -p "$WEB_ROOT"

    # 备份旧文件（如果存在）
    if [ -d "$WEB_ROOT/backup" ]; then
        info "备份现有文件..."
        tar -czf "$WEB_ROOT/backup_$(date +%Y%m%d_%H%M%S).tar.gz" \
            --exclude="$WEB_ROOT/backup*" \
            "$WEB_ROOT"/* 2>/dev/null || true
    fi

    # 清理旧文件（保留备份）
    find "$WEB_ROOT" -mindepth 1 -maxdepth 1 ! -name 'backup*' -exec rm -rf {} +

    # 复制新文件
    cp -r "$PROJECT_DIR/frontend/dist/"* "$WEB_ROOT/"

    # 设置权限
    chown -R www:www "$WEB_ROOT"
    chmod -R 755 "$WEB_ROOT"

    success "前端部署完成"
}

# 启动后端服务
start_backend() {
    info "启动后端服务..."

    cd "$PROJECT_DIR"

    # 使用宝塔专用的docker-compose配置
    if [ -f docker-compose.baota.yml ]; then
        docker-compose -f docker-compose.baota.yml down 2>/dev/null || true
        docker-compose -f docker-compose.baota.yml up -d --build
    else
        error "找不到docker-compose.baota.yml配置文件"
    fi

    success "后端服务启动完成"
}

# 检查服务健康状态
check_health() {
    info "检查服务健康状态..."

    # 等待服务启动
    sleep 10

    # 检查数据库
    if docker exec jobview-db pg_isready -U jobview > /dev/null 2>&1; then
        success "数据库服务正常"
    else
        warning "数据库服务可能还在启动中..."
    fi

    # 检查后端API
    if curl -f http://127.0.0.1:8010/api/auth/health > /dev/null 2>&1; then
        success "后端API服务正常"
    else
        warning "后端API服务可能还在启动中..."
    fi

    # 检查前端（通过宝塔Nginx）
    if curl -f https://jobview.bfsmlt.top > /dev/null 2>&1; then
        success "前端服务正常（HTTPS）"
    elif curl -f http://jobview.bfsmlt.top > /dev/null 2>&1; then
        success "前端服务正常（HTTP）"
    else
        warning "前端服务可能需要配置DNS或等待生效"
    fi
}

# 重启Nginx
restart_nginx() {
    info "重启Nginx服务..."

    # 宝塔面板的Nginx重启命令
    if [ -f /etc/init.d/nginx ]; then
        /etc/init.d/nginx reload
    else
        systemctl reload nginx
    fi

    success "Nginx已重启"
}

# 设置开机自启
setup_autostart() {
    info "设置Docker服务开机自启..."

    # 创建systemd服务文件
    cat > /etc/systemd/system/jobview-backend.service << EOF
[Unit]
Description=JobView Backend Service
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$PROJECT_DIR
ExecStart=/usr/local/bin/docker-compose -f docker-compose.baota.yml up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose.baota.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable jobview-backend.service

    success "开机自启设置完成"
}

# 显示部署信息
show_info() {
    echo ""
    echo "========================================="
    echo "   部署完成！"
    echo "========================================="
    echo ""
    echo "访问地址："
    echo "  - 网站: https://jobview.bfsmlt.top"
    echo "  - API: https://jobview.bfsmlt.top/api"
    echo ""
    echo "服务状态："
    docker-compose -f docker-compose.baota.yml ps
    echo ""
    echo "常用命令："
    echo "  查看日志: docker-compose -f docker-compose.baota.yml logs -f"
    echo "  重启后端: docker-compose -f docker-compose.baota.yml restart"
    echo "  停止服务: docker-compose -f docker-compose.baota.yml down"
    echo ""
    echo "前端文件位置: $WEB_ROOT"
    echo "项目目录: $PROJECT_DIR"
    echo ""
}

# 主菜单
show_menu() {
    echo ""
    echo "请选择操作："
    echo "  1) 完整部署（前端+后端）"
    echo "  2) 仅部署后端"
    echo "  3) 仅部署前端"
    echo "  4) 更新部署"
    echo "  5) 查看服务状态"
    echo "  6) 退出"
    echo ""
    read -p "请输入选项 [1-6]: " choice
}

# 主流程
main() {
    show_menu

    case $choice in
        1)
            info "开始完整部署..."
            check_docker
            check_env
            build_frontend
            deploy_frontend
            start_backend
            restart_nginx
            setup_autostart
            check_health
            show_info
            ;;
        2)
            info "仅部署后端..."
            check_docker
            check_env
            start_backend
            setup_autostart
            check_health
            ;;
        3)
            info "仅部署前端..."
            build_frontend
            deploy_frontend
            restart_nginx
            ;;
        4)
            info "更新部署..."
            cd "$PROJECT_DIR"
            git pull origin main || warning "Git拉取失败，继续使用现有代码"
            build_frontend
            deploy_frontend
            start_backend
            restart_nginx
            check_health
            show_info
            ;;
        5)
            docker-compose -f docker-compose.baota.yml ps
            check_health
            ;;
        6)
            exit 0
            ;;
        *)
            error "无效的选项"
            ;;
    esac
}

# 运行主流程
main