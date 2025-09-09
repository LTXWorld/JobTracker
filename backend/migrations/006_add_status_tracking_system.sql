-- 状态流转跟踪系统数据库扩展
-- 创建时间: 2025-09-08  
-- 目的: 为JobView系统添加完整的状态历史跟踪功能
-- 版本: 1.0
-- 
-- 此迁移脚本基于架构设计文档实现以下功能:
-- 1. 状态历史跟踪表 (job_status_history)
-- 2. 扩展job_applications表以支持状态流转
-- 3. 状态流转配置表 (status_flow_templates)
-- 4. 用户偏好设置表 (user_status_preferences) 
-- 5. 高性能索引优化
-- 6. 数据完整性约束和触发器

-- ============================================================================
-- 1. 核心状态历史跟踪表
-- ============================================================================

-- 创建状态历史记录表
CREATE TABLE IF NOT EXISTS job_status_history (
    id BIGSERIAL PRIMARY KEY,
    job_application_id INTEGER NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL, -- 冗余存储便于快速查询
    
    -- 状态信息
    old_status application_status,
    new_status application_status NOT NULL,
    status_changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 持续时间信息 (以分钟为单位)
    duration_minutes INTEGER, -- 在前一个状态停留的时长
    
    -- 元数据存储 (JSONB格式)
    metadata JSONB DEFAULT '{}', -- 支持存储自定义字段、注释、触发原因等
    
    -- 审计信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 约束
    CONSTRAINT chk_status_change_valid CHECK (old_status IS DISTINCT FROM new_status OR old_status IS NULL),
    CONSTRAINT chk_duration_positive CHECK (duration_minutes IS NULL OR duration_minutes >= 0),
    CONSTRAINT chk_metadata_is_object CHECK (jsonb_typeof(metadata) = 'object')
);

-- 添加表注释
COMMENT ON TABLE job_status_history IS '岗位申请状态历史记录表，记录所有状态变更及元数据';
COMMENT ON COLUMN job_status_history.job_application_id IS '关联的岗位申请ID';
COMMENT ON COLUMN job_status_history.user_id IS '用户ID，用于权限隔离和快速查询';
COMMENT ON COLUMN job_status_history.old_status IS '变更前的状态，初始状态为NULL';
COMMENT ON COLUMN job_status_history.new_status IS '变更后的状态';
COMMENT ON COLUMN job_status_history.duration_minutes IS '在前一个状态的停留时长（分钟）';
COMMENT ON COLUMN job_status_history.metadata IS '状态变更元数据，支持存储注释、原因等信息';

-- ============================================================================
-- 2. 扩展job_applications表
-- ============================================================================

-- 添加状态历史相关字段
ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS status_history JSONB DEFAULT '{"history": [], "summary": {}}';
ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS last_status_change TIMESTAMP WITH TIME ZONE DEFAULT NOW();
ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS status_duration_stats JSONB DEFAULT '{}';
ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS status_version INTEGER DEFAULT 1; -- 乐观锁版本控制

-- 添加字段注释
COMMENT ON COLUMN job_applications.status_history IS '状态历史JSON存储，包含历史记录和汇总信息';
COMMENT ON COLUMN job_applications.last_status_change IS '最后一次状态变更时间';
COMMENT ON COLUMN job_applications.status_duration_stats IS '状态停留时长统计信息';
COMMENT ON COLUMN job_applications.status_version IS '状态版本号，用于乐观锁并发控制';

-- ============================================================================
-- 3. 状态流转配置表
-- ============================================================================

-- 创建状态流转模板表
CREATE TABLE IF NOT EXISTS status_flow_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- 流转配置 (JSONB格式)
    flow_config JSONB NOT NULL DEFAULT '{"transitions": {}, "rules": {}}',
    
    -- 配置属性
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_by INTEGER, -- 可选外键，暂不强制引用
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 约束
    CONSTRAINT uk_status_flow_name UNIQUE(name),
    CONSTRAINT chk_flow_config_is_object CHECK (jsonb_typeof(flow_config) = 'object')
);

-- 添加表注释
COMMENT ON TABLE status_flow_templates IS '状态流转模板配置表';
COMMENT ON COLUMN status_flow_templates.flow_config IS '流转规则配置，定义允许的状态转换和业务规则';

-- 创建默认的状态流转模板
INSERT INTO status_flow_templates (name, description, flow_config, is_default, is_active) VALUES (
    'default_flow',
    '默认求职申请状态流转模板',
    '{
        "transitions": {
            "已投递": ["简历筛选中", "简历筛选未通过", "已拒绝"],
            "简历筛选中": ["笔试中", "简历筛选未通过"],
            "简历筛选未通过": ["流程结束"],
            "笔试中": ["笔试通过", "笔试未通过"],
            "笔试通过": ["一面中"],
            "笔试未通过": ["流程结束"],
            "一面中": ["一面通过", "一面未通过"],
            "一面通过": ["二面中", "三面中", "HR面中"],
            "一面未通过": ["流程结束"],
            "二面中": ["二面通过", "二面未通过"],
            "二面通过": ["三面中", "HR面中"],
            "二面未通过": ["流程结束"],
            "三面中": ["三面通过", "三面未通过"],
            "三面通过": ["HR面中"],
            "三面未通过": ["流程结束"],
            "HR面中": ["HR面通过", "HR面未通过"],
            "HR面通过": ["待发offer"],
            "HR面未通过": ["流程结束"],
            "待发offer": ["已收到offer", "已拒绝"],
            "已收到offer": ["已接受offer", "已拒绝"],
            "已接受offer": ["流程结束"],
            "已拒绝": ["流程结束"],
            "流程结束": []
        },
        "rules": {
            "auto_transitions": {
                "笔试通过": "一面中",
                "一面通过": "二面中",
                "二面通过": "三面中",
                "三面通过": "HR面中",
                "HR面通过": "待发offer"
            },
            "require_confirmation": ["已拒绝", "流程结束"],
            "time_limits": {
                "简历筛选中": 7,
                "笔试中": 3,
                "一面中": 1,
                "二面中": 1,
                "三面中": 1,
                "HR面中": 1
            }
        }
    }',
    true,
    true
) ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- 4. 用户状态偏好设置表
-- ============================================================================

-- 创建用户状态偏好设置表
CREATE TABLE IF NOT EXISTS user_status_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    
    -- 偏好配置 (JSONB格式)
    preference_config JSONB NOT NULL DEFAULT '{}',
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 约束
    CONSTRAINT uk_user_preferences_user_id UNIQUE(user_id),
    CONSTRAINT chk_preference_config_is_object CHECK (jsonb_typeof(preference_config) = 'object')
);

-- 添加表注释
COMMENT ON TABLE user_status_preferences IS '用户状态流转偏好设置表';
COMMENT ON COLUMN user_status_preferences.preference_config IS '用户偏好配置，包含通知设置、显示选项等';

-- ============================================================================
-- 5. 高性能索引优化
-- ============================================================================

-- 5.1 job_status_history表索引
-- 主要查询索引：按用户和岗位查询历史
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_status_history_user_job 
ON job_status_history(user_id, job_application_id, status_changed_at DESC);

-- 时间范围查询索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_status_history_time_range 
ON job_status_history(user_id, status_changed_at DESC);

-- 状态统计查询索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_status_history_status_stats 
ON job_status_history(user_id, new_status, status_changed_at DESC);

-- JSONB元数据查询索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_status_history_metadata 
ON job_status_history USING GIN(metadata);

-- 5.2 扩展的job_applications表索引
-- 状态历史JSONB索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_status_history 
ON job_applications USING GIN(status_history);

-- 最后状态变更时间索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_last_status_change 
ON job_applications(user_id, last_status_change DESC);

-- 复合状态查询索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_status_with_history 
ON job_applications(user_id, status, last_status_change DESC) 
INCLUDE (status_history, status_duration_stats);

-- 状态版本控制索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_version 
ON job_applications(user_id, status_version);

-- 5.3 配置表索引
-- 状态流转模板查询索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_status_flow_templates_active 
ON status_flow_templates(is_active, is_default);

-- 流转配置JSONB索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_status_flow_templates_config 
ON status_flow_templates USING GIN(flow_config);

-- 5.4 偏好设置表索引
-- 用户偏好配置JSONB索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_preferences_config 
ON user_status_preferences USING GIN(preference_config);

-- ============================================================================
-- 6. 数据完整性约束和业务规则
-- ============================================================================

-- 6.1 状态转换验证函数
CREATE OR REPLACE FUNCTION validate_status_transition(
    p_user_id INTEGER,
    p_old_status application_status,
    p_new_status application_status,
    p_flow_template_id INTEGER DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    v_allowed_transitions JSONB;
    v_flow_config JSONB;
BEGIN
    -- 如果是初始状态设置，允许任意状态
    IF p_old_status IS NULL THEN
        RETURN TRUE;
    END IF;
    
    -- 相同状态不允许转换
    IF p_old_status = p_new_status THEN
        RETURN FALSE;
    END IF;
    
    -- 获取流转配置
    SELECT flow_config INTO v_flow_config
    FROM status_flow_templates 
    WHERE (p_flow_template_id IS NOT NULL AND id = p_flow_template_id) 
       OR (p_flow_template_id IS NULL AND is_default = TRUE)
    LIMIT 1;
    
    -- 如果没有找到配置，默认允许转换
    IF v_flow_config IS NULL THEN
        RETURN TRUE;
    END IF;
    
    -- 检查转换规则
    v_allowed_transitions := v_flow_config->'transitions'->p_old_status::text;
    
    -- 如果没有定义转换规则，默认允许
    IF v_allowed_transitions IS NULL THEN
        RETURN TRUE;
    END IF;
    
    -- 检查新状态是否在允许的转换列表中
    RETURN v_allowed_transitions ? p_new_status::text;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_status_transition(INTEGER, application_status, application_status, INTEGER) 
IS '验证状态转换是否合法的业务规则函数';

-- 6.2 状态变更触发器函数
CREATE OR REPLACE FUNCTION trigger_job_status_change() 
RETURNS TRIGGER AS $$
DECLARE
    v_old_status application_status;
    v_duration_minutes INTEGER;
    v_status_history JSONB;
    v_history_entry JSONB;
BEGIN
    -- 获取旧状态
    v_old_status := OLD.status;
    
    -- 如果状态没有变更，跳过处理
    IF NEW.status = OLD.status THEN
        RETURN NEW;
    END IF;
    
    -- 验证状态转换合法性
    IF NOT validate_status_transition(NEW.user_id, v_old_status, NEW.status) THEN
        RAISE EXCEPTION '不允许的状态转换: % -> %', v_old_status, NEW.status;
    END IF;
    
    -- 计算在旧状态的停留时长
    IF OLD.last_status_change IS NOT NULL THEN
        v_duration_minutes := EXTRACT(EPOCH FROM (NOW() - OLD.last_status_change)) / 60;
    ELSE
        v_duration_minutes := EXTRACT(EPOCH FROM (NOW() - OLD.created_at)) / 60;
    END IF;
    
    -- 更新状态变更时间和版本号
    NEW.last_status_change := NOW();
    NEW.status_version := OLD.status_version + 1;
    
    -- 插入状态历史记录
    INSERT INTO job_status_history (
        job_application_id, 
        user_id, 
        old_status, 
        new_status, 
        duration_minutes,
        metadata
    ) VALUES (
        NEW.id, 
        NEW.user_id, 
        v_old_status, 
        NEW.status, 
        v_duration_minutes,
        COALESCE(NEW.status_history->'current_metadata', '{}')
    );
    
    -- 构建历史条目
    v_history_entry := jsonb_build_object(
        'timestamp', extract(epoch from NOW()),
        'old_status', v_old_status,
        'new_status', NEW.status,
        'duration_minutes', v_duration_minutes,
        'changed_at', NOW()::text
    );
    
    -- 更新status_history字段
    v_status_history := COALESCE(NEW.status_history, '{"history": [], "summary": {}}'::jsonb);
    v_status_history := jsonb_set(
        v_status_history, 
        '{history}', 
        (v_status_history->'history') || v_history_entry
    );
    
    -- 更新汇总信息
    v_status_history := jsonb_set(
        v_status_history,
        '{summary}',
        jsonb_build_object(
            'total_changes', jsonb_array_length(v_status_history->'history'),
            'current_status', NEW.status,
            'last_changed', NOW()::text,
            'total_duration_minutes', COALESCE((v_status_history->'summary'->>'total_duration_minutes')::INTEGER, 0) + v_duration_minutes
        )
    );
    
    NEW.status_history := v_status_history;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION trigger_job_status_change() IS '状态变更触发器函数，自动记录状态历史';

-- 创建状态变更触发器
DROP TRIGGER IF EXISTS tr_job_applications_status_change ON job_applications;
CREATE TRIGGER tr_job_applications_status_change
    BEFORE UPDATE ON job_applications
    FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status)
    EXECUTE FUNCTION trigger_job_status_change();

-- ============================================================================
-- 7. 数据查询优化视图
-- ============================================================================

-- 7.1 状态历史汇总视图
CREATE OR REPLACE VIEW job_status_summary AS
SELECT 
    ja.id,
    ja.user_id,
    ja.company_name,
    ja.position_title,
    ja.status as current_status,
    ja.last_status_change,
    ja.status_version,
    
    -- 从status_history JSONB提取的统计信息
    (ja.status_history->'summary'->>'total_changes')::INTEGER as total_status_changes,
    (ja.status_history->'summary'->>'total_duration_minutes')::INTEGER as total_process_duration_minutes,
    
    -- 当前状态停留时长
    EXTRACT(EPOCH FROM (NOW() - ja.last_status_change)) / 60 as current_status_duration_minutes,
    
    -- 历史记录条数
    (SELECT COUNT(*) FROM job_status_history jsh WHERE jsh.job_application_id = ja.id) as history_count,
    
    -- 最近状态变更
    (SELECT jsh.old_status || ' -> ' || jsh.new_status 
     FROM job_status_history jsh 
     WHERE jsh.job_application_id = ja.id 
     ORDER BY jsh.status_changed_at DESC LIMIT 1) as last_transition

FROM job_applications ja
WHERE ja.user_id IS NOT NULL;

COMMENT ON VIEW job_status_summary IS '岗位申请状态汇总视图，提供状态历史统计信息';

-- 7.2 用户状态分析视图
CREATE OR REPLACE VIEW user_status_analytics AS
SELECT 
    user_id,
    -- 基础统计
    COUNT(*) as total_applications,
    COUNT(DISTINCT status) as unique_statuses,
    AVG(status_version) as avg_status_changes_per_application,
    
    -- 状态分布
    COUNT(CASE WHEN status IN ('已投递', '简历筛选中') THEN 1 END) as early_stage_count,
    COUNT(CASE WHEN status IN ('笔试中', '笔试通过', '一面中', '一面通过', '二面中', '二面通过', '三面中', '三面通过', 'HR面中', 'HR面通过') THEN 1 END) as interview_stage_count,
    COUNT(CASE WHEN status IN ('待发offer', '已收到offer', '已接受offer') THEN 1 END) as offer_stage_count,
    COUNT(CASE WHEN status LIKE '%未通过' OR status = '已拒绝' THEN 1 END) as failed_count,
    COUNT(CASE WHEN status = '流程结束' THEN 1 END) as completed_count,
    
    -- 时间统计
    AVG(EXTRACT(EPOCH FROM (NOW() - last_status_change)) / 60) as avg_current_status_duration_minutes,
    MAX(last_status_change) as most_recent_status_change,
    MIN(last_status_change) as oldest_status_change,
    
    -- 成功率分析
    ROUND(
        COUNT(CASE WHEN status IN ('已收到offer', '已接受offer') THEN 1 END) * 100.0 / 
        NULLIF(COUNT(*), 0), 2
    ) as success_rate_percentage

FROM job_applications
WHERE user_id IS NOT NULL
GROUP BY user_id;

COMMENT ON VIEW user_status_analytics IS '用户状态分析视图，提供用户级别的统计分析';

-- ============================================================================
-- 8. 便民查询函数
-- ============================================================================

-- 8.1 获取岗位状态历史
CREATE OR REPLACE FUNCTION get_job_status_history(
    p_user_id INTEGER,
    p_job_application_id INTEGER,
    p_limit INTEGER DEFAULT 50
) RETURNS TABLE (
    id BIGINT,
    old_status application_status,
    new_status application_status,
    status_changed_at TIMESTAMP WITH TIME ZONE,
    duration_minutes INTEGER,
    metadata JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        jsh.id,
        jsh.old_status,
        jsh.new_status,
        jsh.status_changed_at,
        jsh.duration_minutes,
        jsh.metadata
    FROM job_status_history jsh
    INNER JOIN job_applications ja ON jsh.job_application_id = ja.id
    WHERE ja.user_id = p_user_id 
      AND jsh.job_application_id = p_job_application_id
    ORDER BY jsh.status_changed_at DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_job_status_history(INTEGER, INTEGER, INTEGER) 
IS '获取指定岗位申请的状态历史记录';

-- 8.2 状态转换时长分析
CREATE OR REPLACE FUNCTION analyze_status_durations(p_user_id INTEGER)
RETURNS TABLE (
    status application_status,
    avg_duration_minutes NUMERIC,
    min_duration_minutes INTEGER,
    max_duration_minutes INTEGER,
    total_occurrences BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        jsh.old_status as status,
        AVG(jsh.duration_minutes)::NUMERIC(10,2) as avg_duration_minutes,
        MIN(jsh.duration_minutes) as min_duration_minutes,
        MAX(jsh.duration_minutes) as max_duration_minutes,
        COUNT(*) as total_occurrences
    FROM job_status_history jsh
    WHERE jsh.user_id = p_user_id 
      AND jsh.old_status IS NOT NULL
      AND jsh.duration_minutes IS NOT NULL
    GROUP BY jsh.old_status
    ORDER BY avg_duration_minutes DESC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION analyze_status_durations(INTEGER) IS '分析用户在各个状态的停留时长统计';

-- ============================================================================
-- 9. 数据维护和清理
-- ============================================================================

-- 9.1 历史数据清理函数
CREATE OR REPLACE FUNCTION cleanup_old_status_history(
    p_retention_days INTEGER DEFAULT 365
) RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    -- 删除超过保留期的历史记录
    DELETE FROM job_status_history 
    WHERE status_changed_at < (NOW() - INTERVAL '1 day' * p_retention_days);
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    
    -- 记录清理日志
    INSERT INTO maintenance_log (operation, table_name, executed_at, notes) 
    VALUES ('CLEANUP_STATUS_HISTORY', 'job_status_history', NOW(), 
            format('Deleted %s records older than %s days', v_deleted_count, p_retention_days))
    ON CONFLICT DO NOTHING;
    
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_status_history(INTEGER) IS '清理超过保留期的状态历史数据';

-- 9.2 状态历史一致性检查
CREATE OR REPLACE FUNCTION check_status_history_consistency()
RETURNS TABLE (
    job_application_id INTEGER,
    issue_type TEXT,
    issue_description TEXT
) AS $$
BEGIN
    -- 检查状态历史记录与主表状态不一致的情况
    RETURN QUERY
    SELECT 
        ja.id as job_application_id,
        'STATUS_MISMATCH' as issue_type,
        format('Main table status: %s, Latest history status: %s', 
               ja.status, jsh_latest.new_status) as issue_description
    FROM job_applications ja
    INNER JOIN (
        SELECT DISTINCT ON (job_application_id) 
            job_application_id, new_status 
        FROM job_status_history 
        ORDER BY job_application_id, status_changed_at DESC
    ) jsh_latest ON ja.id = jsh_latest.job_application_id
    WHERE ja.status != jsh_latest.new_status;
    
    -- 检查孤立的历史记录
    RETURN QUERY
    SELECT 
        jsh.job_application_id,
        'ORPHANED_HISTORY' as issue_type,
        'Status history record exists but job application is missing' as issue_description
    FROM job_status_history jsh
    LEFT JOIN job_applications ja ON jsh.job_application_id = ja.id
    WHERE ja.id IS NULL;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_status_history_consistency() IS '检查状态历史数据一致性';

-- ============================================================================
-- 10. 性能监控和统计
-- ============================================================================

-- 10.1 扩展索引使用统计视图
CREATE OR REPLACE VIEW status_tracking_index_stats AS
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    CASE 
        WHEN idx_scan = 0 THEN 0
        ELSE ROUND(idx_tup_fetch::NUMERIC / GREATEST(idx_tup_read, 1) * 100, 2)
    END as selectivity_percentage
FROM pg_stat_user_indexes 
WHERE schemaname = 'public' 
  AND tablename IN ('job_applications', 'job_status_history', 'status_flow_templates', 'user_status_preferences')
ORDER BY tablename, idx_scan DESC;

COMMENT ON VIEW status_tracking_index_stats IS '状态跟踪系统索引使用统计视图';

-- 10.2 表大小和性能统计
CREATE OR REPLACE FUNCTION get_status_tracking_table_stats()
RETURNS TABLE (
    table_name TEXT,
    row_count BIGINT,
    table_size TEXT,
    index_size TEXT,
    total_size TEXT,
    avg_row_size INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.table_name::TEXT,
        (SELECT n_tup_ins + n_tup_upd FROM pg_stat_user_tables WHERE relname = t.table_name) as row_count,
        pg_size_pretty(pg_relation_size(t.table_name)) as table_size,
        pg_size_pretty(pg_indexes_size(t.table_name)) as index_size,
        pg_size_pretty(pg_total_relation_size(t.table_name)) as total_size,
        CASE 
            WHEN pg_relation_size(t.table_name) = 0 THEN 0
            ELSE (pg_relation_size(t.table_name) / GREATEST((SELECT n_tup_ins + n_tup_upd FROM pg_stat_user_tables WHERE relname = t.table_name), 1))::INTEGER
        END as avg_row_size
    FROM (VALUES 
        ('job_applications'),
        ('job_status_history'),
        ('status_flow_templates'),
        ('user_status_preferences')
    ) AS t(table_name);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_status_tracking_table_stats() IS '获取状态跟踪系统相关表的大小和性能统计';

-- ============================================================================
-- 11. 记录迁移完成
-- ============================================================================

-- 记录此次迁移的完成
INSERT INTO maintenance_log (operation, table_name, executed_at, notes) 
VALUES ('MIGRATION_006', 'status_tracking_system', NOW(), 'Status tracking system database extension completed successfully')
ON CONFLICT DO NOTHING;

-- 更新表统计信息
ANALYZE job_applications;
ANALYZE job_status_history;
ANALYZE status_flow_templates;  
ANALYZE user_status_preferences;

-- 输出完成信息
DO $$ 
BEGIN 
    RAISE NOTICE '==================================================';
    RAISE NOTICE 'Status Tracking System Database Extension Completed!';
    RAISE NOTICE '==================================================';
    RAISE NOTICE 'Features Added:';
    RAISE NOTICE '- ✅ job_status_history table with JSONB metadata support';
    RAISE NOTICE '- ✅ Extended job_applications table with status tracking fields';
    RAISE NOTICE '- ✅ status_flow_templates table with transition rules';
    RAISE NOTICE '- ✅ user_status_preferences table for user settings';
    RAISE NOTICE '- ✅ High-performance indexes optimized for status queries';
    RAISE NOTICE '- ✅ Status transition validation and triggers';
    RAISE NOTICE '- ✅ Comprehensive views and utility functions';
    RAISE NOTICE '- ✅ Data maintenance and monitoring tools';
    RAISE NOTICE '';
    RAISE NOTICE 'Database Schema Version: 006 - Status Tracking System';
    RAISE NOTICE 'Compatible with existing optimized indexes (84-89%% performance boost)';
    RAISE NOTICE '==================================================';
    RAISE NOTICE 'Next Steps:';
    RAISE NOTICE '1. Run data migration script for existing records';
    RAISE NOTICE '2. Test status transition validation rules';
    RAISE NOTICE '3. Configure user preferences and flow templates';
    RAISE NOTICE '4. Monitor index usage with status_tracking_index_stats view';
    RAISE NOTICE '==================================================';
END $$;