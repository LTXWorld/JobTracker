#!/bin/bash

# 数据库优化迁移脚本
# 创建时间: 2025-09-07
# 用途: 执行数据库查询优化相关的迁移

set -e

# 配置数据库连接信息
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-jobview_db}"
DB_USER="${DB_USER:-ltx}"

# 迁移文件路径
MIGRATION_DIR="./migrations"
OPTIMIZATION_MIGRATION="004_add_performance_indexes.sql"

echo "🚀 开始执行数据库查询优化迁移..."
echo "数据库: $DB_HOST:$DB_PORT/$DB_NAME"
echo "用户: $DB_USER"

# 检查迁移文件是否存在
if [ ! -f "$MIGRATION_DIR/$OPTIMIZATION_MIGRATION" ]; then
    echo "❌ 迁移文件不存在: $MIGRATION_DIR/$OPTIMIZATION_MIGRATION"
    exit 1
fi

echo "📝 执行索引优化迁移..."

# 执行迁移
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_DIR/$OPTIMIZATION_MIGRATION"

if [ $? -eq 0 ]; then
    echo "✅ 索引优化迁移执行成功！"
else
    echo "❌ 索引优化迁移执行失败！"
    exit 1
fi

# 验证索引创建结果
echo "🔍 验证索引创建结果..."

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    indexname, 
    indexdef,
    CASE 
        WHEN indexdef LIKE '%CONCURRENTLY%' THEN '并发创建'
        ELSE '标准创建'
    END as creation_type
FROM pg_indexes 
WHERE tablename = 'job_applications' 
    AND indexname LIKE 'idx_job_applications_%'
ORDER BY indexname;
"

echo "📊 检查表统计信息更新..."

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
ANALYZE job_applications;
SELECT 'Statistics updated for job_applications table' as result;
"

echo "🎯 优化完成！建议的下一步操作："
echo "1. 运行性能测试: go test -bench=. ./tests/service/"
echo "2. 监控慢查询日志"
echo "3. 检查应用程序日志中的数据库连接池状态"
echo "4. 在生产环境部署前进行负载测试"

echo "✨ 数据库查询优化迁移全部完成！"