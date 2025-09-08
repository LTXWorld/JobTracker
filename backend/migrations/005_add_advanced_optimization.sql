-- 高级数据库优化：额外索引和监控工具
-- 创建时间: 2025-09-07
-- 目的: 添加高级优化索引、监控工具和性能统计功能

-- ============================================================================
-- 1. 高级索引优化
-- ============================================================================

-- 分页查询优化索引 - 支持高效分页和计数查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_pagination 
ON job_applications(user_id, application_date DESC, id DESC)
WHERE user_id IS NOT NULL;

-- 多维度筛选索引 - 支持复杂查询条件
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_multi_filter 
ON job_applications(user_id, status, application_date DESC)
WHERE user_id IS NOT NULL AND status IS NOT NULL;

-- 文本搜索索引 - 支持公司名称和职位标题的全文搜索
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_text_search 
ON job_applications USING GIN (to_tsvector('simple', 
    COALESCE(company_name, '') || ' ' || COALESCE(position_title, '')
))
WHERE user_id IS NOT NULL;

-- 时间范围查询索引 - 支持时间范围筛选
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_date_range 
ON job_applications(user_id, application_date)
WHERE user_id IS NOT NULL AND application_date IS NOT NULL;

-- ============================================================================
-- 2. 部分索引优化 - 针对特定业务场景
-- ============================================================================

-- 活跃记录索引 - 只索引最近3个月的记录
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_recent_active 
ON job_applications(user_id, application_date DESC, status)
WHERE user_id IS NOT NULL 
  AND application_date >= (CURRENT_DATE - INTERVAL '3 months')::VARCHAR;

-- 待处理提醒索引 - 只索引需要提醒的记录
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_pending_reminders 
ON job_applications(user_id, reminder_time)
WHERE reminder_enabled = TRUE 
  AND reminder_time IS NOT NULL 
  AND reminder_time > NOW();

-- ============================================================================
-- 3. 数据库统计和监控视图
-- ============================================================================

-- 用户数据统计视图
CREATE OR REPLACE VIEW user_data_summary AS
SELECT 
    user_id,
    COUNT(*) as total_applications,
    COUNT(CASE WHEN status IN ('已投递', '简历筛选中', '笔试中', '一面中', '二面中', '三面中', 'HR面中') THEN 1 END) as in_progress_count,
    COUNT(CASE WHEN status IN ('笔试通过', '一面通过', '二面通过', '三面通过', 'HR面通过', '待发offer', '已收到offer', '已接受offer', '流程结束') THEN 1 END) as success_count,
    COUNT(CASE WHEN status IN ('简历筛选未通过', '笔试未通过', '一面未通过', '二面未通过', '三面未通过', 'HR面未通过', '已拒绝') THEN 1 END) as failed_count,
    MIN(application_date) as first_application_date,
    MAX(application_date) as latest_application_date,
    COUNT(CASE WHEN reminder_enabled = TRUE THEN 1 END) as reminder_count,
    COUNT(CASE WHEN interview_time IS NOT NULL THEN 1 END) as interview_count
FROM job_applications
WHERE user_id IS NOT NULL
GROUP BY user_id;

COMMENT ON VIEW user_data_summary IS '用户投递数据汇总统计视图';

-- 应用状态分布视图
CREATE OR REPLACE VIEW application_status_distribution AS
SELECT 
    user_id,
    status,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (PARTITION BY user_id), 2) as percentage
FROM job_applications
WHERE user_id IS NOT NULL
GROUP BY user_id, status
ORDER BY user_id, count DESC;

COMMENT ON VIEW application_status_distribution IS '应用状态分布统计视图';

-- 月度投递趋势视图
CREATE OR REPLACE VIEW monthly_application_trend AS
SELECT 
    user_id,
    DATE_TRUNC('month', application_date::DATE) as month,
    COUNT(*) as applications_count,
    COUNT(CASE WHEN status NOT IN ('简历筛选未通过', '笔试未通过', '一面未通过', '二面未通过', '三面未通过', 'HR面未通过', '已拒绝') THEN 1 END) as non_failed_count
FROM job_applications
WHERE user_id IS NOT NULL 
  AND application_date IS NOT NULL
  AND application_date ~ '^\d{4}-\d{2}-\d{2}$'  -- 验证日期格式
GROUP BY user_id, DATE_TRUNC('month', application_date::DATE)
ORDER BY user_id, month DESC;

COMMENT ON VIEW monthly_application_trend IS '月度投递趋势统计视图';

-- ============================================================================
-- 4. 性能监控函数
-- ============================================================================

-- 索引使用率统计函数
CREATE OR REPLACE FUNCTION get_index_usage_stats()
RETURNS TABLE (
    table_name TEXT,
    index_name TEXT,
    index_size TEXT,
    scans BIGINT,
    tuples_read BIGINT,
    tuples_fetched BIGINT,
    usage_ratio NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        schemaname::TEXT || '.' || tablename::TEXT as table_name,
        indexrelname::TEXT as index_name,
        pg_size_pretty(pg_relation_size(indexrelid))::TEXT as index_size,
        idx_scan as scans,
        idx_tup_read as tuples_read,
        idx_tup_fetch as tuples_fetched,
        CASE 
            WHEN idx_scan = 0 THEN 0
            ELSE ROUND(idx_tup_fetch::NUMERIC / GREATEST(idx_tup_read, 1) * 100, 2)
        END as usage_ratio
    FROM pg_stat_user_indexes 
    WHERE schemaname = 'public' 
      AND tablename = 'job_applications'
    ORDER BY idx_scan DESC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_index_usage_stats() IS '获取索引使用率统计信息';

-- 表统计信息函数
CREATE OR REPLACE FUNCTION get_table_stats()
RETURNS TABLE (
    table_name TEXT,
    table_size TEXT,
    index_size TEXT,
    total_size TEXT,
    row_count BIGINT,
    seq_scan BIGINT,
    seq_tup_read BIGINT,
    idx_scan BIGINT,
    idx_tup_fetch BIGINT,
    n_tup_ins BIGINT,
    n_tup_upd BIGINT,
    n_tup_del BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        schemaname::TEXT || '.' || tablename::TEXT as table_name,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))::TEXT as table_size,
        pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename))::TEXT as index_size,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) + pg_indexes_size(schemaname||'.'||tablename))::TEXT as total_size,
        n_tup_ins + n_tup_upd + n_tup_del as row_count,
        seq_scan,
        seq_tup_read,
        idx_scan,
        idx_tup_fetch,
        n_tup_ins,
        n_tup_upd,
        n_tup_del
    FROM pg_stat_user_tables 
    WHERE schemaname = 'public' 
      AND tablename = 'job_applications';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_table_stats() IS '获取表统计信息';

-- 查询性能分析函数
CREATE OR REPLACE FUNCTION analyze_query_performance(query_text TEXT)
RETURNS TABLE (
    plan_text TEXT
) AS $$
BEGIN
    RETURN QUERY EXECUTE 'EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) ' || query_text;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION analyze_query_performance(TEXT) IS '分析查询性能的辅助函数';

-- ============================================================================
-- 5. 数据完整性约束优化
-- ============================================================================

-- 添加更强的数据完整性约束
ALTER TABLE job_applications 
ADD CONSTRAINT chk_application_date_format 
CHECK (application_date ~ '^\d{4}-\d{2}-\d{2}$');

ALTER TABLE job_applications 
ADD CONSTRAINT chk_follow_up_date_format 
CHECK (follow_up_date IS NULL OR follow_up_date ~ '^\d{4}-\d{2}-\d{2}$');

-- 添加业务逻辑约束
ALTER TABLE job_applications 
ADD CONSTRAINT chk_reminder_logic 
CHECK (
    (reminder_enabled = FALSE) OR 
    (reminder_enabled = TRUE AND reminder_time IS NOT NULL)
);

-- ============================================================================
-- 6. 定期维护任务
-- ============================================================================

-- 创建定期统计更新函数（需要定期调用）
CREATE OR REPLACE FUNCTION update_table_statistics()
RETURNS VOID AS $$
BEGIN
    -- 更新表统计信息
    ANALYZE job_applications;
    
    -- 记录维护日志
    INSERT INTO maintenance_log (operation, table_name, executed_at) 
    VALUES ('ANALYZE', 'job_applications', NOW())
    ON CONFLICT DO NOTHING;
END;
$$ LANGUAGE plpgsql;

-- 创建维护日志表
CREATE TABLE IF NOT EXISTS maintenance_log (
    id SERIAL PRIMARY KEY,
    operation VARCHAR(50) NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    executed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    notes TEXT,
    UNIQUE(operation, table_name, DATE(executed_at))
);

COMMENT ON TABLE maintenance_log IS '数据库维护操作日志表';

-- ============================================================================
-- 7. 性能测试数据生成器（仅开发环境使用）
-- ============================================================================

-- 性能测试数据生成函数（警告：仅用于测试环境）
CREATE OR REPLACE FUNCTION generate_test_data(
    target_user_id INTEGER,
    record_count INTEGER DEFAULT 1000
)
RETURNS VOID AS $$
DECLARE
    i INTEGER;
    random_status TEXT;
    random_date VARCHAR(10);
    status_array TEXT[] := ARRAY[
        '已投递', '简历筛选中', '简历筛选未通过', '笔试中', '笔试通过', '笔试未通过',
        '一面中', '一面通过', '一面未通过', '二面中', '二面通过', '二面未通过',
        '三面中', '三面通过', '三面未通过', 'HR面中', 'HR面通过', 'HR面未通过',
        '待发offer', '已拒绝', '已收到offer', '已接受offer', '流程结束'
    ];
BEGIN
    -- 警告：仅用于测试环境
    IF current_setting('server_version_num')::int >= 120000 THEN
        RAISE NOTICE 'WARNING: This function should only be used in test environments!';
    END IF;

    FOR i IN 1..record_count LOOP
        random_status := status_array[1 + floor(random() * array_length(status_array, 1))];
        random_date := TO_CHAR(CURRENT_DATE - (random() * 365)::int, 'YYYY-MM-DD');
        
        INSERT INTO job_applications (
            user_id, company_name, position_title, application_date, status,
            job_description, salary_range, work_location, notes
        ) VALUES (
            target_user_id,
            'Test Company ' || i,
            'Test Position ' || i,
            random_date,
            random_status::application_status,
            'Test job description for position ' || i,
            (10000 + random() * 40000)::int || '-' || (15000 + random() * 50000)::int,
            'Test City ' || (i % 10 + 1),
            'Test notes for application ' || i
        );
        
        -- 每1000条记录提交一次
        IF i % 1000 = 0 THEN
            COMMIT;
        END IF;
    END LOOP;
    
    -- 更新统计信息
    ANALYZE job_applications;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_test_data(INTEGER, INTEGER) IS '性能测试数据生成器（仅测试环境使用）';

-- ============================================================================
-- 8. 清理脚本
-- ============================================================================

-- 清理未使用索引的脚本（定期运行建议）
CREATE OR REPLACE FUNCTION cleanup_unused_indexes()
RETURNS TABLE(index_name TEXT, table_name TEXT, action TEXT) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        indexrelname::TEXT as index_name,
        tablename::TEXT as table_name,
        'DROP INDEX ' || indexrelname || ';' as action
    FROM pg_stat_user_indexes 
    WHERE schemaname = 'public' 
      AND tablename = 'job_applications'
      AND idx_scan = 0
      AND indexrelname NOT LIKE '%_pkey';  -- 保留主键索引
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_unused_indexes() IS '生成清理未使用索引的建议脚本';

-- ============================================================================
-- 9. 总结和文档
-- ============================================================================

-- 创建优化总结视图
CREATE OR REPLACE VIEW optimization_summary AS
SELECT 
    'Indexes Created' as category,
    COUNT(*) as count,
    string_agg(indexname, ', ') as details
FROM pg_indexes 
WHERE tablename = 'job_applications' AND schemaname = 'public'
UNION ALL
SELECT 
    'Table Size' as category,
    1 as count,
    pg_size_pretty(pg_total_relation_size('job_applications')) as details
UNION ALL
SELECT 
    'Index Size' as category,
    1 as count,
    pg_size_pretty(pg_indexes_size('job_applications')) as details;

COMMENT ON VIEW optimization_summary IS '数据库优化总结视图';

-- 记录此次迁移的完成
INSERT INTO maintenance_log (operation, table_name, executed_at, notes) 
VALUES ('MIGRATION_005', 'job_applications', NOW(), 'Advanced optimization: indexes, monitoring, and performance tools added')
ON CONFLICT DO NOTHING;

-- 输出完成信息
DO $$ 
BEGIN 
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'Advanced Database Optimization Completed!';
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'Added Features:';
    RAISE NOTICE '- Advanced pagination and multi-filter indexes';
    RAISE NOTICE '- Full-text search capabilities';
    RAISE NOTICE '- Performance monitoring views and functions';
    RAISE NOTICE '- Data integrity constraints';
    RAISE NOTICE '- Maintenance and cleanup tools';
    RAISE NOTICE '';
    RAISE NOTICE 'Next Steps:';
    RAISE NOTICE '1. Run ANALYZE on the table to update statistics';
    RAISE NOTICE '2. Monitor index usage with get_index_usage_stats()';
    RAISE NOTICE '3. Use the monitoring views for performance insights';
    RAISE NOTICE '==============================================';
END $$;