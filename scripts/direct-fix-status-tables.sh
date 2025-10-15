#!/bin/bash
# direct-fix-status-tables.sh - 直接创建状态相关表

echo "========================================="
echo "   直接修复状态跟踪表"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() {
    echo -e "${YELLOW}→ $1${NC}"
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# 直接复制关键的SQL并执行
info "创建所有必需的状态跟踪表..."

docker exec jobview-db psql -U jobview -d jobview_db << 'EOSQL'
-- 开始事务
BEGIN;

-- 1. 创建状态流转模板表（如果不存在）
CREATE TABLE IF NOT EXISTS status_flow_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. 创建状态流转规则表
DROP TABLE IF EXISTS status_transitions CASCADE;
CREATE TABLE status_transitions (
    id SERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES status_flow_templates(id) ON DELETE CASCADE,
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    allow_skip BOOLEAN DEFAULT false,
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(template_id, from_status, to_status)
);

-- 3. 创建状态跟踪历史表
DROP TABLE IF EXISTS status_tracking_history CASCADE;
CREATE TABLE status_tracking_history (
    id SERIAL PRIMARY KEY,
    application_id INTEGER NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    changed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    change_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. 创建或更新用户状态偏好表
DROP TABLE IF EXISTS user_status_preferences CASCADE;
CREATE TABLE user_status_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    template_id INTEGER REFERENCES status_flow_templates(id) ON DELETE SET NULL,
    allow_backward BOOLEAN DEFAULT true,
    allow_skip BOOLEAN DEFAULT true,
    require_note BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. 创建索引
CREATE INDEX idx_status_transitions_template ON status_transitions(template_id);
CREATE INDEX idx_status_transitions_from_to ON status_transitions(from_status, to_status);
CREATE INDEX idx_status_history_app ON status_tracking_history(application_id);
CREATE INDEX idx_status_history_date ON status_tracking_history(changed_at);

-- 6. 确保有默认模板
INSERT INTO status_flow_templates (name, description, is_default)
VALUES ('default_flow', '默认招聘流程', true)
ON CONFLICT (name) DO UPDATE SET is_default = true;

-- 7. 插入完整的状态流转规则
WITH template AS (
    SELECT id FROM status_flow_templates WHERE name = 'default_flow'
)
INSERT INTO status_transitions (template_id, from_status, to_status, allow_skip, is_enabled)
SELECT
    t.id,
    v.from_status,
    v.to_status,
    v.allow_skip,
    v.is_enabled
FROM template t
CROSS JOIN (VALUES
    -- 基础流转
    ('未投递', '已投递', false, true),
    ('已投递', '简历筛选中', false, true),
    ('简历筛选中', '笔试中', false, true),
    ('简历筛选中', '一面中', true, true),
    ('笔试中', '一面中', false, true),
    ('一面中', '二面中', false, true),
    ('一面中', 'HR面中', true, true),
    ('二面中', '三面中', false, true),
    ('二面中', 'HR面中', true, true),
    ('三面中', 'HR面中', false, true),
    ('HR面中', '已通过', false, true),

    -- 失败流转
    ('简历筛选中', '简历挂', false, true),
    ('笔试中', '笔试挂', false, true),
    ('一面中', '一面挂', false, true),
    ('二面中', '二面挂', false, true),
    ('三面中', '三面挂', false, true),
    ('HR面中', 'HR面挂', false, true),

    -- 直接通过
    ('一面中', '已通过', true, true),
    ('二面中', '已通过', true, true),
    ('三面中', '已通过', true, true),

    -- 允许回退
    ('简历筛选中', '已投递', true, true),
    ('笔试中', '简历筛选中', true, true),
    ('一面中', '笔试中', true, true),
    ('二面中', '一面中', true, true),
    ('三面中', '二面中', true, true),
    ('HR面中', '三面中', true, true)
) AS v(from_status, to_status, allow_skip, is_enabled);

-- 8. 为所有用户创建偏好
INSERT INTO user_status_preferences (user_id, template_id, allow_backward, allow_skip, require_note)
SELECT u.id, t.id, true, true, false
FROM users u
CROSS JOIN status_flow_templates t
WHERE t.name = 'default_flow'
ON CONFLICT (user_id) DO UPDATE SET
    template_id = EXCLUDED.template_id,
    allow_backward = true,
    allow_skip = true;

COMMIT;

-- 显示结果
SELECT 'Tables Created:' as info;
SELECT table_name FROM information_schema.tables
WHERE table_name IN ('status_transitions', 'status_tracking_history', 'user_status_preferences')
ORDER BY table_name;

SELECT '' as x, 'Transition Rules:' as info;
SELECT COUNT(*) as count FROM status_transitions;

SELECT '' as x, 'Key Rules:' as info;
SELECT from_status || ' → ' || to_status as rule
FROM status_transitions
WHERE from_status = '已投递' OR to_status = '简历筛选中'
LIMIT 5;
EOSQL

if [ $? -eq 0 ]; then
    success "表创建成功！"
else
    echo "创建过程中有错误，但可能部分成功"
fi

# 重启后端
info "重启后端服务..."
docker-compose -f docker-compose.baota.yml restart backend

echo ""
echo "========================================="
echo "修复完成！"
echo ""
echo "现在请："
echo "1. 清除浏览器缓存 (Ctrl+Shift+Del)"
echo "2. 退出并重新登录系统"
echo "3. 测试拖拽功能"
echo "========================================="
