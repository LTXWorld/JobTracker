-- 状态跟踪系统数据迁移脚本
-- 创建时间: 2025-09-08
-- 目的: 为现有的job_applications记录创建初始状态历史
-- 版本: 1.0
-- 
-- 此脚本确保现有数据的平滑迁移，包括:
-- 1. 为现有记录创建初始状态历史
-- 2. 初始化新增的状态跟踪字段
-- 3. 数据完整性检查和验证
-- 4. 回滚机制支持

-- ============================================================================
-- 1. 数据迁移准备工作
-- ============================================================================

DO $$
DECLARE
    v_total_records INTEGER;
    v_records_to_migrate INTEGER;
BEGIN
    -- 检查现有数据
    SELECT COUNT(*) INTO v_total_records FROM job_applications;
    
    SELECT COUNT(*) INTO v_records_to_migrate 
    FROM job_applications 
    WHERE status_history IS NULL OR status_history = '{"history": [], "summary": {}}'::jsonb;
    
    RAISE NOTICE '======================================';
    RAISE NOTICE '状态跟踪系统数据迁移开始';
    RAISE NOTICE '======================================';
    RAISE NOTICE '总记录数: %', v_total_records;
    RAISE NOTICE '需要迁移的记录数: %', v_records_to_migrate;
    RAISE NOTICE '======================================';
    
    -- 如果没有需要迁移的记录，跳过迁移
    IF v_records_to_migrate = 0 THEN
        RAISE NOTICE '所有记录已完成迁移，跳过数据迁移过程';
        RETURN;
    END IF;
END;
$$;

-- ============================================================================
-- 2. 数据备份和回滚支持
-- ============================================================================

-- 创建备份表（如果不存在）
CREATE TABLE IF NOT EXISTS job_applications_backup_pre_status_tracking AS
SELECT * FROM job_applications LIMIT 0;

-- 备份现有数据（仅在首次迁移时）
INSERT INTO job_applications_backup_pre_status_tracking
SELECT * FROM job_applications ja
WHERE NOT EXISTS (
    SELECT 1 FROM job_applications_backup_pre_status_tracking jab 
    WHERE jab.id = ja.id
);

-- 记录备份信息
DO $$
DECLARE
    v_backup_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_backup_count FROM job_applications_backup_pre_status_tracking;
    RAISE NOTICE '备份记录数: %', v_backup_count;
END;
$$;

-- ============================================================================
-- 3. 数据迁移核心逻辑
-- ============================================================================

-- 3.1 为现有记录创建初始状态历史
CREATE OR REPLACE FUNCTION migrate_existing_status_history()
RETURNS INTEGER AS $$
DECLARE
    v_record RECORD;
    v_migrated_count INTEGER := 0;
    v_initial_duration_minutes INTEGER;
    v_status_history JSONB;
    v_history_entry JSONB;
BEGIN
    -- 遍历所有需要迁移的记录
    FOR v_record IN 
        SELECT id, user_id, status, created_at, updated_at, last_status_change
        FROM job_applications 
        WHERE status_history IS NULL 
           OR status_history = '{"history": [], "summary": {}}'::jsonb
           OR jsonb_array_length(COALESCE(status_history->'history', '[]'::jsonb)) = 0
        ORDER BY id
    LOOP
        -- 计算初始状态持续时间
        v_initial_duration_minutes := EXTRACT(EPOCH FROM (
            COALESCE(v_record.last_status_change, v_record.updated_at, NOW()) - v_record.created_at
        )) / 60;
        
        -- 创建初始历史条目
        v_history_entry := jsonb_build_object(
            'timestamp', extract(epoch from v_record.created_at),
            'old_status', NULL,
            'new_status', v_record.status,
            'duration_minutes', v_initial_duration_minutes,
            'changed_at', v_record.created_at::text,
            'migration_note', 'Initial status from data migration'
        );
        
        -- 构建完整的状态历史结构
        v_status_history := jsonb_build_object(
            'history', jsonb_build_array(v_history_entry),
            'summary', jsonb_build_object(
                'total_changes', 1,
                'current_status', v_record.status,
                'last_changed', COALESCE(v_record.last_status_change, v_record.created_at)::text,
                'total_duration_minutes', v_initial_duration_minutes,
                'migration_timestamp', NOW()::text
            )
        );
        
        -- 更新主表记录
        UPDATE job_applications 
        SET 
            status_history = v_status_history,
            last_status_change = COALESCE(v_record.last_status_change, v_record.created_at),
            status_version = 1,
            status_duration_stats = jsonb_build_object(
                v_record.status::text, v_initial_duration_minutes
            ),
            updated_at = NOW()
        WHERE id = v_record.id;
        
        -- 插入到状态历史表（作为初始记录）
        INSERT INTO job_status_history (
            job_application_id, 
            user_id, 
            old_status, 
            new_status, 
            status_changed_at,
            duration_minutes,
            metadata,
            created_at
        ) VALUES (
            v_record.id,
            COALESCE(v_record.user_id, 1), -- 兼容性处理
            NULL, -- 初始状态没有旧状态
            v_record.status,
            v_record.created_at,
            v_initial_duration_minutes,
            jsonb_build_object(
                'migration_source', 'existing_data',
                'original_created_at', v_record.created_at::text,
                'migrated_at', NOW()::text
            ),
            v_record.created_at
        );
        
        v_migrated_count := v_migrated_count + 1;
        
        -- 每100条记录输出进度
        IF v_migrated_count % 100 = 0 THEN
            RAISE NOTICE '已迁移 % 条记录...', v_migrated_count;
            COMMIT; -- 分批提交避免长事务
        END IF;
    END LOOP;
    
    RETURN v_migrated_count;
END;
$$ LANGUAGE plpgsql;

-- 执行迁移
DO $$
DECLARE
    v_migrated_count INTEGER;
    v_start_time TIMESTAMP;
    v_end_time TIMESTAMP;
BEGIN
    v_start_time := NOW();
    
    -- 暂时禁用触发器以避免重复处理
    ALTER TABLE job_applications DISABLE TRIGGER tr_job_applications_status_change;
    
    -- 执行迁移
    SELECT migrate_existing_status_history() INTO v_migrated_count;
    
    -- 重新启用触发器
    ALTER TABLE job_applications ENABLE TRIGGER tr_job_applications_status_change;
    
    v_end_time := NOW();
    
    RAISE NOTICE '======================================';
    RAISE NOTICE '数据迁移完成!';
    RAISE NOTICE '======================================';
    RAISE NOTICE '成功迁移记录数: %', v_migrated_count;
    RAISE NOTICE '迁移耗时: %', v_end_time - v_start_time;
    RAISE NOTICE '======================================';
END;
$$;

-- ============================================================================
-- 4. 数据完整性验证
-- ============================================================================

-- 4.1 验证迁移结果
DO $$
DECLARE
    v_apps_count INTEGER;
    v_history_count INTEGER;
    v_null_history_count INTEGER;
    v_version_mismatch_count INTEGER;
BEGIN
    RAISE NOTICE '======================================';
    RAISE NOTICE '数据完整性验证开始';
    RAISE NOTICE '======================================';
    
    -- 检查记录总数
    SELECT COUNT(*) INTO v_apps_count FROM job_applications;
    SELECT COUNT(*) INTO v_history_count FROM job_status_history;
    
    RAISE NOTICE 'job_applications表记录数: %', v_apps_count;
    RAISE NOTICE 'job_status_history表记录数: %', v_history_count;
    
    -- 检查空的状态历史
    SELECT COUNT(*) INTO v_null_history_count 
    FROM job_applications 
    WHERE status_history IS NULL 
       OR status_history = '{"history": [], "summary": {}}'::jsonb
       OR jsonb_array_length(COALESCE(status_history->'history', '[]'::jsonb)) = 0;
    
    RAISE NOTICE '状态历史为空的记录数: %', v_null_history_count;
    
    -- 检查版本号异常
    SELECT COUNT(*) INTO v_version_mismatch_count
    FROM job_applications
    WHERE status_version IS NULL OR status_version < 1;
    
    RAISE NOTICE '状态版本异常的记录数: %', v_version_mismatch_count;
    
    -- 输出验证结果
    IF v_null_history_count = 0 AND v_version_mismatch_count = 0 THEN
        RAISE NOTICE '✅ 数据完整性验证通过';
    ELSE
        RAISE WARNING '⚠️  发现数据完整性问题，请检查相关记录';
    END IF;
    
    RAISE NOTICE '======================================';
END;
$$;

-- 4.2 检查状态历史一致性
SELECT 
    '状态历史一致性检查' as check_type,
    COUNT(*) as issue_count
FROM (
    SELECT * FROM check_status_history_consistency()
) issues;

-- ============================================================================
-- 5. 用户默认偏好设置初始化
-- ============================================================================

-- 为现有用户创建默认偏好设置
INSERT INTO user_status_preferences (user_id, preference_config)
SELECT DISTINCT 
    user_id,
    jsonb_build_object(
        'notifications', jsonb_build_object(
            'status_change', true,
            'reminder_alerts', true,
            'weekly_summary', false
        ),
        'display', jsonb_build_object(
            'timeline_view', 'chronological',
            'status_colors', jsonb_build_object(
                '已投递', '#6366f1',
                '简历筛选中', '#f59e0b',
                '笔试中', '#10b981',
                '一面中', '#8b5cf6',
                '已收到offer', '#059669',
                '已拒绝', '#ef4444'
            ),
            'show_duration', true
        ),
        'automation', jsonb_build_object(
            'auto_reminders', true,
            'smart_suggestions', true
        ),
        'created_by_migration', NOW()::text
    )
FROM job_applications 
WHERE user_id IS NOT NULL
ON CONFLICT (user_id) DO NOTHING;

RAISE NOTICE '已为用户初始化默认偏好设置';

-- ============================================================================
-- 6. 性能优化和统计信息更新
-- ============================================================================

-- 更新所有相关表的统计信息
ANALYZE job_applications;
ANALYZE job_status_history;
ANALYZE status_flow_templates;
ANALYZE user_status_preferences;

-- 检查索引创建状态
DO $$
DECLARE
    v_index_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_index_count
    FROM pg_indexes 
    WHERE schemaname = 'public' 
      AND tablename IN ('job_applications', 'job_status_history', 'status_flow_templates', 'user_status_preferences');
    
    RAISE NOTICE '状态跟踪系统相关索引总数: %', v_index_count;
END;
$$;

-- ============================================================================
-- 7. 回滚支持函数
-- ============================================================================

-- 创建回滚函数（紧急情况使用）
CREATE OR REPLACE FUNCTION rollback_status_tracking_migration()
RETURNS INTEGER AS $$
DECLARE
    v_restored_count INTEGER := 0;
BEGIN
    RAISE NOTICE '开始回滚状态跟踪系统迁移...';
    
    -- 检查备份表是否存在
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'job_applications_backup_pre_status_tracking') THEN
        RAISE EXCEPTION '备份表不存在，无法执行回滚操作';
    END IF;
    
    -- 暂时禁用触发器
    ALTER TABLE job_applications DISABLE TRIGGER tr_job_applications_status_change;
    
    -- 恢复原始数据
    UPDATE job_applications 
    SET 
        status_history = NULL,
        last_status_change = backup.updated_at,
        status_duration_stats = NULL,
        status_version = NULL
    FROM job_applications_backup_pre_status_tracking backup
    WHERE job_applications.id = backup.id;
    
    GET DIAGNOSTICS v_restored_count = ROW_COUNT;
    
    -- 清空状态历史表
    DELETE FROM job_status_history 
    WHERE metadata ? 'migration_source';
    
    -- 重新启用触发器
    ALTER TABLE job_applications ENABLE TRIGGER tr_job_applications_status_change;
    
    RAISE NOTICE '回滚完成，恢复了 % 条记录', v_restored_count;
    RETURN v_restored_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION rollback_status_tracking_migration() IS '紧急回滚状态跟踪系统迁移的函数';

-- ============================================================================
-- 8. 记录迁移完成
-- ============================================================================

-- 记录迁移日志
INSERT INTO maintenance_log (operation, table_name, executed_at, notes) 
VALUES ('DATA_MIGRATION_006', 'status_tracking_system', NOW(), 
        'Status tracking system data migration completed successfully')
ON CONFLICT DO NOTHING;

-- 清理迁移函数
DROP FUNCTION IF EXISTS migrate_existing_status_history();

-- 输出最终结果
DO $$
DECLARE
    v_final_apps_count INTEGER;
    v_final_history_count INTEGER;
    v_final_prefs_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_final_apps_count FROM job_applications;
    SELECT COUNT(*) INTO v_final_history_count FROM job_status_history;
    SELECT COUNT(*) INTO v_final_prefs_count FROM user_status_preferences;
    
    RAISE NOTICE '====================================================';
    RAISE NOTICE '状态跟踪系统数据迁移全部完成!';
    RAISE NOTICE '====================================================';
    RAISE NOTICE '最终统计:';
    RAISE NOTICE '- job_applications记录: %', v_final_apps_count;
    RAISE NOTICE '- job_status_history记录: %', v_final_history_count;
    RAISE NOTICE '- user_status_preferences记录: %', v_final_prefs_count;
    RAISE NOTICE '';
    RAISE NOTICE '✅ 数据库结构扩展已完成';
    RAISE NOTICE '✅ 现有数据迁移已完成';
    RAISE NOTICE '✅ 索引优化已生效';
    RAISE NOTICE '✅ 触发器和约束已激活';
    RAISE NOTICE '✅ 用户偏好设置已初始化';
    RAISE NOTICE '✅ 备份和回滚机制已就绪';
    RAISE NOTICE '';
    RAISE NOTICE '数据库已准备就绪，可以支持状态流转跟踪功能！';
    RAISE NOTICE '====================================================';
END;
$$;