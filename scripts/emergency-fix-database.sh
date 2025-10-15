#!/bin/bash
# emergency-fix-database.sh - 紧急修复数据库结构

echo "========================================="
echo "   JobView 数据库紧急修复"
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

# 1. 备份当前数据库
info "备份当前数据库..."
docker exec jobview-db pg_dump -U jobview -d jobview_db > /tmp/jobview_backup_$(date +%Y%m%d_%H%M%S).sql
success "备份完成: /tmp/jobview_backup_$(date +%Y%m%d_%H%M%S).sql"

# 2. 创建缺失的表和字段
info "创建缺失的数据库结构..."

cat > /tmp/emergency_fix.sql << 'EOF'
-- 开始事务
BEGIN;

-- 1. 创建schema_migrations表（如果不存在）
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. 创建status_transitions表（状态流转规则）
CREATE TABLE IF NOT EXISTS status_transitions (
    id SERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL,
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    allow_skip BOOLEAN DEFAULT false,
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES status_flow_templates(id) ON DELETE CASCADE,
    UNIQUE(template_id, from_status, to_status)
);

-- 3. 创建status_tracking_history表（状态变更历史）
CREATE TABLE IF NOT EXISTS status_tracking_history (
    id SERIAL PRIMARY KEY,
    application_id INTEGER NOT NULL,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    changed_by INTEGER,
    change_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES job_applications(id) ON DELETE CASCADE,
    FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE SET NULL
);

-- 4. 修改user_status_preferences表（添加缺失的字段）
-- 首先检查表是否存在
DO $$
BEGIN
    -- 如果表不存在，创建它
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_status_preferences') THEN
        CREATE TABLE user_status_preferences (
            id SERIAL PRIMARY KEY,
            user_id INTEGER NOT NULL UNIQUE,
            template_id INTEGER,
            allow_backward BOOLEAN DEFAULT true,
            allow_skip BOOLEAN DEFAULT true,
            require_note BOOLEAN DEFAULT false,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
            FOREIGN KEY (template_id) REFERENCES status_flow_templates(id) ON DELETE SET NULL
        );
    ELSE
        -- 如果表存在，添加缺失的列
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                      WHERE table_name = 'user_status_preferences' AND column_name = 'template_id') THEN
            ALTER TABLE user_status_preferences ADD COLUMN template_id INTEGER;
            ALTER TABLE user_status_preferences ADD CONSTRAINT fk_template
                FOREIGN KEY (template_id) REFERENCES status_flow_templates(id) ON DELETE SET NULL;
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                      WHERE table_name = 'user_status_preferences' AND column_name = 'allow_backward') THEN
            ALTER TABLE user_status_preferences ADD COLUMN allow_backward BOOLEAN DEFAULT true;
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                      WHERE table_name = 'user_status_preferences' AND column_name = 'allow_skip') THEN
            ALTER TABLE user_status_preferences ADD COLUMN allow_skip BOOLEAN DEFAULT true;
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                      WHERE table_name = 'user_status_preferences' AND column_name = 'require_note') THEN
            ALTER TABLE user_status_preferences ADD COLUMN require_note BOOLEAN DEFAULT false;
        END IF;
    END IF;
END $$;

-- 5. 创建索引以提高性能
CREATE INDEX IF NOT EXISTS idx_status_transitions_template ON status_transitions(template_id);
CREATE INDEX IF NOT EXISTS idx_status_transitions_from_to ON status_transitions(from_status, to_status);
CREATE INDEX IF NOT EXISTS idx_status_history_app ON status_tracking_history(application_id);
CREATE INDEX IF NOT EXISTS idx_status_history_date ON status_tracking_history(changed_at);

-- 6. 插入默认的状态流转规则
-- 获取default_flow模板的ID
WITH default_template AS (
    SELECT id FROM status_flow_templates WHERE name = 'default_flow' LIMIT 1
)
INSERT INTO status_transitions (template_id, from_status, to_status, allow_skip, is_enabled)
SELECT
    dt.id,
    v.from_status,
    v.to_status,
    v.allow_skip,
    true
FROM default_template dt
CROSS JOIN (
    VALUES
    -- 正向流转（基础流程）
    ('未投递', '已投递', false),
    ('已投递', '简历筛选中', false),
    ('简历筛选中', '笔试中', true),
    ('简历筛选中', '一面中', true),  -- 可以跳过笔试
    ('笔试中', '一面中', false),
    ('一面中', '二面中', false),
    ('一面中', 'HR面中', true),         -- 可以跳过二面直接HR
    ('二面中', '三面中', true),         -- 可以跳过三面
    ('二面中', 'HR面中', true),         -- 可以跳过三面直接HR
    ('三面中', 'HR面中', false),
    ('HR面中', '已通过', false),

    -- 失败状态流转
    ('简历筛选中', '简历挂', false),
    ('笔试中', '笔试挂', false),
    ('一面中', '一面挂', false),
    ('二面中', '二面挂', false),
    ('三面中', '三面挂', false),
    ('HR面中', 'HR面挂', false),

    -- 任何面试阶段都可以直接通过
    ('简历筛选中', '已通过', true),
    ('笔试中', '已通过', true),
    ('一面中', '已通过', true),
    ('二面中', '已通过', true),
    ('三面中', '已通过', true),

    -- 回退流转（允许状态回退）
    ('简历筛选中', '已投递', true),
    ('笔试中', '简历筛选中', true),
    ('一面中', '简历筛选中', true),
    ('一面中', '笔试中', true),
    ('二面中', '一面中', true),
    ('三面中', '二面中', true),
    ('HR面中', '三面中', true),
    ('HR面中', '二面中', true),
    ('HR面中', '一面中', true),

    -- 从失败状态恢复
    ('简历挂', '简历筛选中', true),
    ('笔试挂', '笔试中', true),
    ('一面挂', '一面中', true),
    ('二面挂', '二面中', true),
    ('三面挂', '三面中', true),
    ('HR面挂', 'HR面中', true),

    -- 从失败状态重新开始
    ('简历挂', '已投递', true),
    ('笔试挂', '简历筛选中', true),
    ('一面挂', '简历筛选中', true),
    ('二面挂', '一面中', true),
    ('三面挂', '二面中', true),
    ('HR面挂', '三面中', true)
) AS v(from_status, to_status, allow_skip)
ON CONFLICT (template_id, from_status, to_status) DO UPDATE
SET
    allow_skip = EXCLUDED.allow_skip,
    is_enabled = true,
    updated_at = CURRENT_TIMESTAMP;

-- 7. 为所有用户创建默认偏好设置
INSERT INTO user_status_preferences (user_id, template_id, allow_backward, allow_skip, require_note)
SELECT
    u.id,
    sft.id,
    true,  -- 允许回退
    true,  -- 允许跳过
    false  -- 不强制要求备注
FROM users u
CROSS JOIN status_flow_templates sft
WHERE sft.name = 'default_flow'
ON CONFLICT (user_id) DO UPDATE
SET
    template_id = EXCLUDED.template_id,
    allow_backward = true,
    allow_skip = true,
    updated_at = CURRENT_TIMESTAMP;

-- 8. 记录迁移版本
INSERT INTO schema_migrations (version, applied_at)
VALUES
    ('001_initial_schema', CURRENT_TIMESTAMP),
    ('002_status_tracking', CURRENT_TIMESTAMP),
    ('003_status_transitions', CURRENT_TIMESTAMP),
    ('004_emergency_fix', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO NOTHING;

-- 提交事务
COMMIT;

-- 验证修复结果
SELECT 'Tables created:' as info;
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN ('status_transitions', 'status_tracking_history', 'user_status_preferences', 'schema_migrations');

SELECT '' as blank, 'Status transitions count:' as info;
SELECT COUNT(*) as rules_count FROM status_transitions WHERE is_enabled = true;

SELECT '' as blank, 'User preferences count:' as info;
SELECT COUNT(*) as users_with_preferences FROM user_status_preferences;

EOF

# 执行修复
info "执行数据库修复..."
docker exec -i jobview-db psql -U jobview -d jobview_db < /tmp/emergency_fix.sql

if [ $? -eq 0 ]; then
    success "数据库结构修复成功"
else
    error "数据库修复失败，请检查错误信息"
    exit 1
fi

# 3. 重启后端服务
info "重启后端服务以应用更改..."
docker-compose -f docker-compose.baota.yml restart backend 2>/dev/null || docker-compose restart backend

success "服务重启完成"

# 4. 验证修复
echo ""
echo "========================================="
echo "   验证修复结果"
echo "========================================="

echo -e "\n${YELLOW}检查创建的表：${NC}"
docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT table_name,
       CASE WHEN table_name IS NOT NULL THEN '✓ 已创建' ELSE '✗ 未创建' END as status
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN ('status_transitions', 'status_tracking_history', 'user_status_preferences', 'schema_migrations')
ORDER BY table_name;"

echo -e "\n${YELLOW}检查状态流转规则：${NC}"
docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT from_status, to_status,
       CASE WHEN is_enabled THEN '✓' ELSE '✗' END as enabled,
       CASE WHEN allow_skip THEN '是' ELSE '否' END as can_skip
FROM status_transitions
WHERE from_status IN ('已投递', '简历筛选中')
   OR to_status IN ('简历筛选中', '笔试中', '一面中')
ORDER BY from_status, to_status
LIMIT 10;"

echo ""
echo "========================================="
echo -e "${GREEN}修复完成！${NC}"
echo ""
echo "请执行以下操作："
echo "1. 清除浏览器缓存（Ctrl+F5）"
echo "2. 重新登录系统"
echo "3. 测试拖拽功能"
echo ""
echo "如果还有问题，查看日志："
echo "  docker logs jobview-backend --tail 50"
echo "========================================="
