#!/bin/bash
# backup.sh - JobView 数据备份脚本

set -e

# 配置
BACKUP_DIR="${BACKUP_DIR:-/backup/jobview}"
DATE=$(date +%Y%m%d_%H%M%S)
DB_CONTAINER="${DB_CONTAINER:-jobview-db}"
DB_USER="${DB_USER:-jobview}"
DB_NAME="${DB_NAME:-jobview_db}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

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

# 函数：打印信息
info() {
    echo -e "${YELLOW}→ $1${NC}"
}

echo "========================================="
echo "   JobView 数据备份"
echo "========================================="
echo ""
echo "备份时间: $(date)"
echo "备份目录: $BACKUP_DIR"
echo ""

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 1. 备份数据库
info "备份PostgreSQL数据库..."
if docker exec "$DB_CONTAINER" pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_DIR/db_backup_$DATE.sql.gz"; then
    success "数据库备份完成: db_backup_$DATE.sql.gz"

    # 显示备份文件大小
    SIZE=$(du -h "$BACKUP_DIR/db_backup_$DATE.sql.gz" | cut -f1)
    echo "  文件大小: $SIZE"
else
    error "数据库备份失败"
fi

# 2. 备份上传的文件（如果存在）
if [ -d "./backend/uploads" ]; then
    info "备份上传文件..."
    if tar -czf "$BACKUP_DIR/uploads_backup_$DATE.tar.gz" ./backend/uploads; then
        success "上传文件备份完成: uploads_backup_$DATE.tar.gz"

        # 显示备份文件大小
        SIZE=$(du -h "$BACKUP_DIR/uploads_backup_$DATE.tar.gz" | cut -f1)
        echo "  文件大小: $SIZE"
    else
        error "上传文件备份失败"
    fi
else
    info "没有找到上传文件目录，跳过"
fi

# 3. 备份环境配置（不包含敏感信息）
info "备份配置文件..."
if tar -czf "$BACKUP_DIR/config_backup_$DATE.tar.gz" \
    --exclude='.env' \
    --exclude='*.key' \
    --exclude='*.pem' \
    docker-compose*.yml \
    frontend/nginx*.conf \
    backend/migrations \
    2>/dev/null; then
    success "配置文件备份完成: config_backup_$DATE.tar.gz"
else
    warning "部分配置文件可能未备份"
fi

# 4. 清理旧备份
info "清理超过 $RETENTION_DAYS 天的旧备份..."
DELETED_COUNT=0

# 清理数据库备份
find "$BACKUP_DIR" -name "db_backup_*.sql.gz" -mtime +$RETENTION_DAYS -delete -print | while read -r file; do
    echo "  删除: $(basename "$file")"
    DELETED_COUNT=$((DELETED_COUNT + 1))
done

# 清理上传文件备份
find "$BACKUP_DIR" -name "uploads_backup_*.tar.gz" -mtime +$RETENTION_DAYS -delete -print | while read -r file; do
    echo "  删除: $(basename "$file")"
    DELETED_COUNT=$((DELETED_COUNT + 1))
done

# 清理配置备份
find "$BACKUP_DIR" -name "config_backup_*.tar.gz" -mtime +$RETENTION_DAYS -delete -print | while read -r file; do
    echo "  删除: $(basename "$file")"
    DELETED_COUNT=$((DELETED_COUNT + 1))
done

if [ $DELETED_COUNT -eq 0 ]; then
    echo "  没有需要清理的旧备份"
else
    success "已清理 $DELETED_COUNT 个旧备份文件"
fi

# 5. 显示备份统计
echo ""
echo "备份统计："
echo "----------------------------------------"
echo "当前备份文件："
ls -lh "$BACKUP_DIR"/*.gz 2>/dev/null | tail -5 || echo "  暂无备份文件"

echo ""
echo "磁盘使用情况："
df -h "$BACKUP_DIR" | tail -1

echo ""
echo "========================================="
success "备份完成！"
echo "备份位置: $BACKUP_DIR"
echo "========================================="

# 可选：将备份上传到远程存储
# 取消注释以下代码以启用远程备份
# if command -v rclone &> /dev/null; then
#     info "上传备份到远程存储..."
#     rclone copy "$BACKUP_DIR/db_backup_$DATE.sql.gz" remote:jobview-backups/
#     success "远程备份完成"
# fi