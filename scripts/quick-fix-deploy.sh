#!/bin/bash
# quick-fix-deploy.sh - 快速修复部署脚本

set -e

echo "========================================="
echo "   JobView 快速修复部署"
echo "========================================="

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 函数
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# 配置变量
WEB_ROOT="/www/wwwroot/jobview.bfsmlt.top"
PROJECT_DIR="/opt/jobview"

# 1. 清理前端目录（保留宝塔文件）
info "清理前端目录..."
cd "$WEB_ROOT"

# 列出要保留的文件
KEEP_FILES=(".user.ini" ".htaccess" ".well-known")

# 删除其他所有文件和目录
for item in *; do
    skip=false
    for keep in "${KEEP_FILES[@]}"; do
        if [[ "$item" == "$keep" ]]; then
            skip=true
            break
        fi
    done

    if [ "$skip" = false ] && [ -e "$item" ]; then
        rm -rf "$item" 2>/dev/null || true
        echo "  删除: $item"
    fi
done

# 处理隐藏文件（除了要保留的）
for item in .[^.]*; do
    if [ -e "$item" ]; then
        skip=false
        for keep in "${KEEP_FILES[@]}"; do
            if [[ "$item" == "$keep" ]]; then
                skip=true
                break
            fi
        done

        if [ "$skip" = false ]; then
            rm -rf "$item" 2>/dev/null || true
            echo "  删除: $item"
        fi
    fi
done

success "目录清理完成"

# 2. 构建前端（如果需要）
if [ ! -d "$PROJECT_DIR/frontend/dist" ]; then
    info "构建前端应用..."
    cd "$PROJECT_DIR/frontend"

    # 安装依赖
    if [ ! -d "node_modules" ]; then
        info "安装依赖..."
        npm ci
    fi

    # 构建
    info "构建生产版本..."
    VITE_API_BASE=/api npm run build
    success "前端构建完成"
else
    info "使用已存在的构建文件"
fi

# 3. 复制前端文件
info "部署前端文件..."
cp -r "$PROJECT_DIR/frontend/dist/"* "$WEB_ROOT/"

# 4. 设置权限
info "设置文件权限..."
chown -R www:www "$WEB_ROOT"
chmod -R 755 "$WEB_ROOT"

# 保护.user.ini
if [ -f "$WEB_ROOT/.user.ini" ]; then
    chattr +i "$WEB_ROOT/.user.ini" 2>/dev/null || true
fi

success "前端部署完成"

# 5. 启动后端服务
info "启动后端服务..."
cd "$PROJECT_DIR"

# 检查使用哪个配置文件
if [ -f "docker-compose.baota.yml" ]; then
    COMPOSE_FILE="docker-compose.baota.yml"
else
    COMPOSE_FILE="docker-compose.yml"
fi

# 停止旧服务
docker-compose -f "$COMPOSE_FILE" down 2>/dev/null || true

# 启动新服务
docker-compose -f "$COMPOSE_FILE" up -d

success "后端服务已启动"

# 6. 检查服务状态
echo ""
echo "========================================="
echo "   服务状态"
echo "========================================="

# 等待服务启动
sleep 5

# 检查后端
if curl -f http://127.0.0.1:8010/api/auth/health > /dev/null 2>&1; then
    success "后端API正常"
else
    echo "⚠ 后端API可能还在启动中..."
fi

# 检查数据库
if docker exec jobview-db pg_isready > /dev/null 2>&1; then
    success "数据库正常"
else
    echo "⚠ 数据库可能还在启动中..."
fi

# 7. 显示访问信息
echo ""
echo "========================================="
echo "   部署完成"
echo "========================================="
echo ""
echo "访问地址："
echo "  网站: https://jobview.bfsmlt.top"
echo "  API: https://jobview.bfsmlt.top/api"
echo ""
echo "查看日志："
echo "  docker-compose -f $COMPOSE_FILE logs -f"
echo ""

# 8. 重启Nginx（可选）
read -p "是否重启Nginx服务？[y/N]: " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    /etc/init.d/nginx reload || systemctl reload nginx
    success "Nginx已重启"
fi