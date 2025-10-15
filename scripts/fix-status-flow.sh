#!/bin/bash
# fix-status-flow.sh - 修复状态流转配置

echo "========================================="
echo "   JobView 状态流转修复脚本"
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

# 检查参数
if [ "$1" == "--force" ]; then
    FORCE_MODE=true
    echo -e "${YELLOW}强制模式：将重置所有状态流转配置${NC}"
else
    FORCE_MODE=false
fi

# 1. 备份当前配置
info "备份当前数据库配置..."
docker exec jobview-db pg_dump -U jobview -d jobview_db \
    -t status_flow_templates \
    -t status_transitions \
    -t user_status_preferences \
    > /tmp/status_flow_backup_$(date +%Y%m%d_%H%M%S).sql
success "备份完成"

# 2. 修复状态流转配置
info "修复状态流转配置..."

# 创建修复SQL
cat > /tmp/fix_status_flow.sql << 'EOF'
-- 开始事务
BEGIN;

-- 1. 确保默认模板存在
INSERT INTO status_flow_templates (name, description, is_default, created_at, updated_at)
SELECT
    '标准招聘流程',
    '标准的招聘流程，包含从投递到最终结果的完整流转',
    true,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM status_flow_templates WHERE is_default = true
);

-- 获取默认模板ID
WITH default_template AS (
    SELECT id FROM status_flow_templates WHERE is_default = true LIMIT 1
)

-- 2. 插入完整的状态流转规则（如果不存在）
INSERT INTO status_transitions (template_id, from_status, to_status, allow_skip, is_enabled, created_at, updated_at)
SELECT
    dt.id,
    v.from_status,
    v.to_status,
    v.allow_skip,
    true,
    NOW(),
    NOW()
FROM default_template dt
CROSS JOIN (
    VALUES
    -- 正向流转（进行中状态）
    ('未投递', '已投递', false),
    ('已投递', '简历筛选中', false),
    ('简历筛选中', '笔试中', true),
    ('简历筛选中', '一面中', true),
    ('笔试中', '一面中', false),
    ('一面中', '二面中', false),
    ('一面中', 'HR面中', true),
    ('二面中', '三面中', false),
    ('二面中', 'HR面中', true),
    ('三面中', 'HR面中', false),
    ('HR面中', '已通过', false),

    -- 失败状态流转
    ('简历筛选中', '简历挂', false),
    ('笔试中', '笔试挂', false),
    ('一面中', '一面挂', false),
    ('二面中', '二面挂', false),
    ('三面中', '三面挂', false),
    ('HR面中', 'HR面挂', false),

    -- 直接到已通过（跳过中间步骤）
    ('简历筛选中', '已通过', true),
    ('笔试中', '已通过', true),
    ('一面中', '已通过', true),
    ('二面中', '已通过', true),
    ('三面中', '已通过', true),

    -- 回退流转（如果允许）
    ('简历筛选中', '已投递', false),
    ('笔试中', '简历筛选中', false),
    ('一面中', '简历筛选中', false),
    ('一面中', '笔试中', false),
    ('二面中', '一面中', false),
    ('三面中', '二面中', false),
    ('HR面中', '三面中', false),

    -- 从失败状态恢复
    ('简历挂', '简历筛选中', false),
    ('笔试挂', '笔试中', false),
    ('一面挂', '一面中', false),
    ('二面挂', '二面中', false),
    ('三面挂', '三面中', false),
    ('HR面挂', 'HR面中', false)
) AS v(from_status, to_status, allow_skip)
ON CONFLICT (template_id, from_status, to_status) DO UPDATE
SET
    is_enabled = true,
    allow_skip = EXCLUDED.allow_skip,
    updated_at = NOW();

-- 3. 更新用户偏好设置（允许回退）
UPDATE user_status_preferences
SET
    allow_backward = true,
    allow_skip = true,
    updated_at = NOW()
WHERE template_id IN (
    SELECT id FROM status_flow_templates WHERE is_default = true
);

-- 4. 如果用户没有偏好设置，创建默认的
INSERT INTO user_status_preferences (user_id, template_id, allow_backward, allow_skip, require_note, created_at, updated_at)
SELECT
    u.id,
    sft.id,
    true,  -- 允许回退
    true,  -- 允许跳过
    false, -- 不强制要求备注
    NOW(),
    NOW()
FROM users u
CROSS JOIN status_flow_templates sft
WHERE sft.is_default = true
    AND NOT EXISTS (
        SELECT 1 FROM user_status_preferences usp
        WHERE usp.user_id = u.id
    );

-- 提交事务
COMMIT;

-- 显示修复结果
SELECT '修复完成！' as message;

-- 验证配置
SELECT
    '默认模板' as check_item,
    CASE WHEN COUNT(*) > 0 THEN '✓ 存在' ELSE '✗ 不存在' END as status
FROM status_flow_templates WHERE is_default = true
UNION ALL
SELECT
    '状态流转规则数量' as check_item,
    COUNT(*)::text || ' 条' as status
FROM status_transitions st
JOIN status_flow_templates sft ON st.template_id = sft.id
WHERE sft.is_default = true AND st.is_enabled = true
UNION ALL
SELECT
    '用户偏好配置' as check_item,
    COUNT(*)::text || ' 个用户' as status
FROM user_status_preferences;

EOF

# 执行修复
if [ "$FORCE_MODE" = true ]; then
    info "执行强制修复..."
    docker exec -i jobview-db psql -U jobview -d jobview_db < /tmp/fix_status_flow.sql
else
    echo ""
    echo -e "${YELLOW}即将执行的修复操作：${NC}"
    echo "1. 确保存在默认状态流转模板"
    echo "2. 添加完整的状态流转规则"
    echo "3. 允许状态回退和跳过"
    echo "4. 为所有用户设置默认偏好"
    echo ""
    read -p "是否执行修复？[y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker exec -i jobview-db psql -U jobview -d jobview_db < /tmp/fix_status_flow.sql
    else
        error "修复已取消"
        exit 1
    fi
fi

success "修复完成"

# 3. 重启后端服务以应用更改
info "重启后端服务..."
docker-compose -f docker-compose.baota.yml restart backend || docker-compose restart backend
success "服务已重启"

echo ""
echo "========================================="
echo -e "${GREEN}修复完成！${NC}"
echo ""
echo "请执行以下操作验证："
echo "1. 刷新浏览器页面（强制刷新：Ctrl+F5）"
echo "2. 重新登录系统"
echo "3. 尝试拖拽状态"
echo ""
echo "如果问题仍然存在，请运行诊断脚本："
echo "  ./scripts/diagnose-status-flow.sh"
echo "========================================="
