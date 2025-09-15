#!/bin/bash
# force-fix-database.sh - 强制修复数据库结构

echo "========================================="
echo "   JobView 数据库强制修复"
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

# 1. 备份
info "备份当前数据库..."
BACKUP_FILE="/tmp/jobview_force_backup_$(date +%Y%m%d_%H%M%S).sql"
docker exec jobview-db pg_dump -U jobview -d jobview_db > "$BACKUP_FILE"
success "备份完成: $BACKUP_FILE"

# 2. 分步修复（每步独立事务）
info "开始分步修复数据库..."

# Step 1: 创建缺失的表
info "Step 1: 创建基础表..."
docker exec jobview-db psql -U jobview -d jobview_db << 'EOF'
-- 创建schema_migrations表
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建status_transitions表
CREATE TABLE IF NOT EXISTS status_transitions (
    id SERIAL PRIMARY KEY,
    template_id INTEGER,
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    allow_skip BOOLEAN DEFAULT false,
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建status_tracking_history表
CREATE TABLE IF NOT EXISTS status_tracking_history (
    id SERIAL PRIMARY KEY,
    application_id INTEGER,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    changed_by INTEGER,
    change_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

SELECT 'Tables created' as status;
EOF

# Step 2: 修复user_status_preferences表结构
info "Step 2: 修复user_status_preferences表..."
docker exec jobview-db psql -U jobview -d jobview_db << 'EOF'
-- 检查并添加缺失的列
DO $$
BEGIN
    -- 添加template_id列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'user_status_preferences'
                  AND column_name = 'template_id') THEN
        ALTER TABLE user_status_preferences ADD COLUMN template_id INTEGER;
    END IF;

    -- 添加allow_backward列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'user_status_preferences'
                  AND column_name = 'allow_backward') THEN
        ALTER TABLE user_status_preferences ADD COLUMN allow_backward BOOLEAN DEFAULT true;
    END IF;

    -- 添加allow_skip列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'user_status_preferences'
                  AND column_name = 'allow_skip') THEN
        ALTER TABLE user_status_preferences ADD COLUMN allow_skip BOOLEAN DEFAULT true;
    END IF;

    -- 添加require_note列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'user_status_preferences'
                  AND column_name = 'require_note') THEN
        ALTER TABLE user_status_preferences ADD COLUMN require_note BOOLEAN DEFAULT false;
    END IF;

    -- 添加updated_at列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'user_status_preferences'
                  AND column_name = 'updated_at') THEN
        ALTER TABLE user_status_preferences ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
    END IF;

    -- 添加created_at列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'user_status_preferences'
                  AND column_name = 'created_at') THEN
        ALTER TABLE user_status_preferences ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
    END IF;
END $$;

SELECT 'user_status_preferences fixed' as status;
EOF

# Step 3: 添加外键约束（如果可能）
info "Step 3: 添加约束..."
docker exec jobview-db psql -U jobview -d jobview_db << 'EOF'
-- 添加唯一约束到status_transitions
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint
                  WHERE conname = 'unique_status_transition') THEN
        ALTER TABLE status_transitions
        ADD CONSTRAINT unique_status_transition
        UNIQUE(template_id, from_status, to_status);
    END IF;
END $$;

-- 尝试添加外键（如果表存在）
DO $$
BEGIN
    -- status_transitions的外键
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'status_flow_templates') THEN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_status_transitions_template') THEN
            ALTER TABLE status_transitions
            ADD CONSTRAINT fk_status_transitions_template
            FOREIGN KEY (template_id) REFERENCES status_flow_templates(id) ON DELETE CASCADE;
        END IF;
    END IF;

    -- status_tracking_history的外键
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_status_history_app') THEN
        ALTER TABLE status_tracking_history
        ADD CONSTRAINT fk_status_history_app
        FOREIGN KEY (application_id) REFERENCES job_applications(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_status_history_user') THEN
        ALTER TABLE status_tracking_history
        ADD CONSTRAINT fk_status_history_user
        FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;

    -- user_status_preferences的外键
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'status_flow_templates') THEN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_prefs_template') THEN
            ALTER TABLE user_status_preferences
            ADD CONSTRAINT fk_user_prefs_template
            FOREIGN KEY (template_id) REFERENCES status_flow_templates(id) ON DELETE SET NULL;
        END IF;
    END IF;
END $$;

SELECT 'Constraints added' as status;
EOF

# Step 4: 插入状态流转规则
info "Step 4: 插入状态流转规则..."
docker exec jobview-db psql -U jobview -d jobview_db << 'EOF'
-- 获取template_id
DO $$
DECLARE
    v_template_id INTEGER;
BEGIN
    -- 获取default_flow的ID
    SELECT id INTO v_template_id FROM status_flow_templates WHERE name = 'default_flow' LIMIT 1;

    IF v_template_id IS NULL THEN
        -- 如果没有default_flow，获取任意默认模板
        SELECT id INTO v_template_id FROM status_flow_templates WHERE is_default = true LIMIT 1;
    END IF;

    IF v_template_id IS NOT NULL THEN
        -- 删除旧规则（如果有）
        DELETE FROM status_transitions WHERE template_id = v_template_id;

        -- 插入完整的状态流转规则
        INSERT INTO status_transitions (template_id, from_status, to_status, allow_skip, is_enabled) VALUES
        -- 正向流转
        (v_template_id, '未投递', '已投递', false, true),
        (v_template_id, '已投递', '简历筛选中', false, true),
        (v_template_id, '简历筛选中', '笔试中', false, true),
        (v_template_id, '简历筛选中', '一面中', true, true),
        (v_template_id, '笔试中', '一面中', false, true),
        (v_template_id, '一面中', '二面中', false, true),
        (v_template_id, '二面中', '三面中', false, true),
        (v_template_id, '二面中', 'HR面中', true, true),
        (v_template_id, '三面中', 'HR面中', false, true),
        (v_template_id, 'HR面中', '已通过', false, true),

        -- 失败状态
        (v_template_id, '简历筛选中', '简历挂', false, true),
        (v_template_id, '笔试中', '笔试挂', false, true),
        (v_template_id, '一面中', '一面挂', false, true),
        (v_template_id, '二面中', '二面挂', false, true),
        (v_template_id, '三面中', '三面挂', false, true),
        (v_template_id, 'HR面中', 'HR面挂', false, true),

        -- 直接通过
        (v_template_id, '简历筛选中', '已通过', true, true),
        (v_template_id, '笔试中', '已通过', true, true),
        (v_template_id, '一面中', '已通过', true, true),
        (v_template_id, '二面中', '已通过', true, true),
        (v_template_id, '三面中', '已通过', true, true),

        -- 回退流转
        (v_template_id, '简历筛选中', '已投递', true, true),
        (v_template_id, '笔试中', '简历筛选中', true, true),
        (v_template_id, '一面中', '简历筛选中', true, true),
        (v_template_id, '一面中', '笔试中', true, true),
        (v_template_id, '二面中', '一面中', true, true),
        (v_template_id, '三面中', '二面中', true, true),
        (v_template_id, 'HR面中', '三面中', true, true),

        -- 从失败恢复
        (v_template_id, '简历挂', '简历筛选中', true, true),
        (v_template_id, '笔试挂', '笔试中', true, true),
        (v_template_id, '一面挂', '一面中', true, true),
        (v_template_id, '二面挂', '二面中', true, true),
        (v_template_id, '三面挂', '三面中', true, true),
        (v_template_id, 'HR面挂', 'HR面中', true, true);

        RAISE NOTICE 'Inserted status transitions for template_id: %', v_template_id;
    ELSE
        RAISE NOTICE 'No template found, skipping status transitions';
    END IF;
END $$;

SELECT COUNT(*) as transition_rules FROM status_transitions;
EOF

# Step 5: 更新用户偏好
info "Step 5: 更新用户偏好设置..."
docker exec jobview-db psql -U jobview -d jobview_db << 'EOF'
-- 清空并重建用户偏好
TRUNCATE TABLE user_status_preferences;

-- 为所有用户插入默认偏好
INSERT INTO user_status_preferences (user_id, template_id, allow_backward, allow_skip, require_note)
SELECT
    u.id,
    sft.id,
    true,
    true,
    false
FROM users u
CROSS JOIN status_flow_templates sft
WHERE sft.is_default = true OR sft.name = 'default_flow'
LIMIT 1;

-- 如果没有模板，仍然创建偏好（template_id为NULL）
INSERT INTO user_status_preferences (user_id, allow_backward, allow_skip, require_note)
SELECT
    u.id,
    true,
    true,
    false
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM user_status_preferences WHERE user_id = u.id);

SELECT COUNT(*) as users_with_preferences FROM user_status_preferences;
EOF

# 3. 验证结果
echo ""
echo "========================================="
echo "   验证修复结果"
echo "========================================="

info "检查所有表是否存在..."
docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT table_name,
       '✓ 已创建' as status
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN ('status_transitions', 'status_tracking_history', 'user_status_preferences', 'schema_migrations')
ORDER BY table_name;"

info "检查状态流转规则数量..."
docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT COUNT(*) as total_rules FROM status_transitions;"

info "检查关键流转规则..."
docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT from_status || ' → ' || to_status as transition,
       CASE WHEN is_enabled THEN '✓ 启用' ELSE '✗ 禁用' END as status
FROM status_transitions
WHERE (from_status = '已投递' AND to_status = '简历筛选中')
   OR (from_status = '简历筛选中' AND to_status IN ('笔试中', '一面中'))
   OR (from_status = '一面中' AND to_status = '二面中');"

# 4. 重启服务
info "重启后端服务..."
docker-compose -f docker-compose.baota.yml restart backend 2>/dev/null || docker-compose restart backend

echo ""
echo "========================================="
echo -e "${GREEN}强制修复完成！${NC}"
echo ""
echo "下一步操作："
echo "1. 清除浏览器缓存 (Ctrl+F5)"
echo "2. 退出并重新登录"
echo "3. 测试拖拽功能"
echo ""
echo "如果仍有问题，请手动检查："
echo "  docker exec -it jobview-db psql -U jobview -d jobview_db"
echo "  \\dt  -- 查看所有表"
echo "  SELECT * FROM status_transitions LIMIT 5;"
echo "========================================="