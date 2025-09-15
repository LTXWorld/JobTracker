#!/bin/bash
# diagnose-status-flow.sh - 诊断状态流转问题

echo "========================================="
echo "   JobView 状态流转诊断"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 1. 检查数据库中的状态流转配置
echo -e "${YELLOW}1. 检查数据库状态流转配置${NC}"
echo "----------------------------------------"

docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT
    sft.name as template_name,
    sft.is_default,
    sft.created_at,
    sft.updated_at
FROM status_flow_templates sft
ORDER BY sft.is_default DESC, sft.created_at;
"

echo ""
echo -e "${YELLOW}2. 检查状态流转规则${NC}"
echo "----------------------------------------"

docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT
    st.from_status,
    st.to_status,
    st.allow_skip,
    st.is_enabled,
    sft.name as template_name
FROM status_transitions st
JOIN status_flow_templates sft ON st.template_id = sft.id
WHERE sft.is_default = true
ORDER BY
    CASE st.from_status
        WHEN '未投递' THEN 1
        WHEN '已投递' THEN 2
        WHEN '简历筛选中' THEN 3
        WHEN '笔试中' THEN 4
        WHEN '一面中' THEN 5
        WHEN '二面中' THEN 6
        WHEN '三面中' THEN 7
        WHEN 'HR面中' THEN 8
        WHEN '已通过' THEN 9
        ELSE 10
    END,
    st.to_status;
"

echo ""
echo -e "${YELLOW}3. 检查用户状态偏好设置${NC}"
echo "----------------------------------------"

docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT
    u.username,
    usp.template_id,
    usp.allow_backward,
    usp.allow_skip,
    usp.require_note,
    sft.name as template_name
FROM user_status_preferences usp
JOIN users u ON usp.user_id = u.id
LEFT JOIN status_flow_templates sft ON usp.template_id = sft.id
LIMIT 5;
"

echo ""
echo -e "${YELLOW}4. 统计各状态的投递记录数${NC}"
echo "----------------------------------------"

docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT
    status,
    COUNT(*) as count
FROM job_applications
GROUP BY status
ORDER BY count DESC;
"

echo ""
echo -e "${YELLOW}5. 检查最近的状态变更历史${NC}"
echo "----------------------------------------"

docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT
    ja.company,
    ja.position,
    sth.from_status,
    sth.to_status,
    sth.changed_at,
    sth.change_reason
FROM status_tracking_history sth
JOIN job_applications ja ON sth.application_id = ja.id
ORDER BY sth.changed_at DESC
LIMIT 10;
"

echo ""
echo -e "${YELLOW}6. 检查数据库迁移状态${NC}"
echo "----------------------------------------"

docker exec jobview-db psql -U jobview -d jobview_db -c "
SELECT
    version,
    applied_at
FROM schema_migrations
ORDER BY version DESC
LIMIT 5;
"

echo ""
echo "========================================="
echo -e "${GREEN}诊断完成！${NC}"
echo ""
echo "请检查以上输出，特别注意："
echo "1. 是否有默认模板 (is_default = true)"
echo "2. 状态流转规则是否完整"
echo "3. allow_backward 是否为 false（禁止回退）"
echo "4. 从'已投递'到'简历筛选中'的规则是否存在且启用"
echo "========================================="