#!/bin/bash
# manual-run-migrations.sh - 手动执行所有迁移文件

echo "========================================="
echo "   手动执行数据库迁移"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
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

error() {
    echo -e "${RED}✗ $1${NC}"
}

# 检查迁移文件目录
MIGRATION_DIR="/opt/jobview/backend/migrations"

if [ ! -d "$MIGRATION_DIR" ]; then
    error "迁移文件目录不存在: $MIGRATION_DIR"
    exit 1
fi

# 备份数据库
info "备份数据库..."
BACKUP_FILE="/tmp/jobview_migration_backup_$(date +%Y%m%d_%H%M%S).sql"
docker exec jobview-db pg_dump -U jobview -d jobview_db > "$BACKUP_FILE"
success "备份完成: $BACKUP_FILE"

# 创建迁移记录表
info "创建迁移记录表..."
docker exec jobview-db psql -U jobview -d jobview_db << 'EOF'
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
EOF

# 获取已执行的迁移
info "检查已执行的迁移..."
EXECUTED_MIGRATIONS=$(docker exec jobview-db psql -U jobview -d jobview_db -t -c "SELECT version FROM schema_migrations;" 2>/dev/null || echo "")

# 执行迁移文件
cd "$MIGRATION_DIR"
for migration_file in *.sql; do
    migration_name="${migration_file%.*}"

    # 检查是否已执行
    if echo "$EXECUTED_MIGRATIONS" | grep -q "$migration_name"; then
        echo "  ⏭️  跳过已执行: $migration_file"
        continue
    fi

    info "执行迁移: $migration_file"

    # 执行迁移文件
    if docker exec -i jobview-db psql -U jobview -d jobview_db < "$migration_file" 2>/tmp/migration_error.log; then
        # 记录成功的迁移
        docker exec jobview-db psql -U jobview -d jobview_db -c "
            INSERT INTO schema_migrations (version) VALUES ('$migration_name')
            ON CONFLICT (version) DO NOTHING;
        "
        success "成功: $migration_file"
    else
        error "失败: $migration_file"
        echo "错误详情:"
        cat /tmp/migration_error.log

        # 询问是否继续
        read -p "是否继续执行其他迁移？[y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
done

# 验证关键表
echo ""
echo "========================================="
echo "   验证数据库结构"
echo "========================================="

info "检查关键表..."
docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT table_name,
       CASE WHEN table_name IS NOT NULL THEN '✓' ELSE '✗' END as exists
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN (
    'job_applications',
    'users',
    'status_flow_templates',
    'status_transitions',
    'status_tracking_history',
    'user_status_preferences',
    'export_tasks',
    'schema_migrations'
)
ORDER BY table_name;"

info "检查状态流转规则..."
docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT COUNT(*) as rule_count FROM status_transitions;" 2>/dev/null || echo "status_transitions表不存在"

# 重启后端
info "重启后端服务..."
docker-compose -f docker-compose.baota.yml restart backend || docker-compose restart backend

echo ""
echo "========================================="
echo -e "${GREEN}迁移执行完成！${NC}"
echo ""
echo "请测试系统功能"
echo "========================================="