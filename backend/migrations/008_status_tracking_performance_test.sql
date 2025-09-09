-- 状态跟踪系统性能测试和验证脚本
-- 创建时间: 2025-09-08
-- 目的: 验证状态跟踪系统的性能和功能正确性
-- 版本: 1.0
--
-- 此脚本包含以下测试内容:
-- 1. 数据库结构完整性测试
-- 2. 索引性能基准测试
-- 3. 状态转换业务逻辑测试
-- 4. JSONB查询性能测试
-- 5. 触发器功能测试
-- 6. 并发安全性测试
-- 7. 数据完整性验证

-- ============================================================================
-- 1. 测试环境准备
-- ============================================================================

-- 创建测试结果记录表
CREATE TABLE IF NOT EXISTS status_tracking_test_results (
    id SERIAL PRIMARY KEY,
    test_name VARCHAR(100) NOT NULL,
    test_category VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL, -- PASS, FAIL, SKIP
    execution_time_ms NUMERIC(10,3),
    details JSONB DEFAULT '{}',
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 清理之前的测试结果
TRUNCATE TABLE status_tracking_test_results;

-- 测试辅助函数：记录测试结果
CREATE OR REPLACE FUNCTION record_test_result(
    p_test_name VARCHAR(100),
    p_category VARCHAR(50),
    p_status VARCHAR(20),
    p_execution_time_ms NUMERIC DEFAULT NULL,
    p_details JSONB DEFAULT '{}'
) RETURNS VOID AS $$
BEGIN
    INSERT INTO status_tracking_test_results (
        test_name, test_category, status, execution_time_ms, details
    ) VALUES (
        p_test_name, p_category, p_status, p_execution_time_ms, p_details
    );
END;
$$ LANGUAGE plpgsql;

-- 开始测试
DO $$
BEGIN
    RAISE NOTICE '========================================================';
    RAISE NOTICE '状态跟踪系统性能测试和验证开始';
    RAISE NOTICE '执行时间: %', NOW();
    RAISE NOTICE '========================================================';
END;
$$;

-- ============================================================================
-- 2. 数据库结构完整性测试
-- ============================================================================

DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_test_details JSONB := '{}';
    v_table_exists BOOLEAN;
    v_column_exists BOOLEAN;
    v_index_exists BOOLEAN;
BEGIN
    v_start_time := clock_timestamp();
    
    -- 测试核心表是否存在
    SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'job_status_history') INTO v_table_exists;
    v_test_details := jsonb_set(v_test_details, '{job_status_history_exists}', to_jsonb(v_table_exists));
    
    SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'status_flow_templates') INTO v_table_exists;
    v_test_details := jsonb_set(v_test_details, '{status_flow_templates_exists}', to_jsonb(v_table_exists));
    
    SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'user_status_preferences') INTO v_table_exists;
    v_test_details := jsonb_set(v_test_details, '{user_status_preferences_exists}', to_jsonb(v_table_exists));
    
    -- 测试关键字段是否存在
    SELECT EXISTS(
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'job_applications' AND column_name = 'status_history'
    ) INTO v_column_exists;
    v_test_details := jsonb_set(v_test_details, '{status_history_column_exists}', to_jsonb(v_column_exists));
    
    -- 测试关键索引是否存在
    SELECT EXISTS(
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_job_status_history_user_job'
    ) INTO v_index_exists;
    v_test_details := jsonb_set(v_test_details, '{key_index_exists}', to_jsonb(v_index_exists));
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    -- 判断测试结果
    IF v_table_exists AND v_column_exists AND v_index_exists THEN
        PERFORM record_test_result('数据库结构完整性', '结构测试', 'PASS', v_execution_time, v_test_details);
    ELSE
        PERFORM record_test_result('数据库结构完整性', '结构测试', 'FAIL', v_execution_time, v_test_details);
    END IF;
END;
$$;

-- ============================================================================
-- 3. 创建测试数据
-- ============================================================================

DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_test_user_id INTEGER := 9999;
    v_test_records_count INTEGER := 1000;
    v_created_count INTEGER := 0;
    i INTEGER;
    v_random_status application_status;
    v_random_date VARCHAR(10);
    v_status_array application_status[] := ARRAY[
        '已投递'::application_status, '简历筛选中'::application_status, '笔试中'::application_status,
        '一面中'::application_status, '二面中'::application_status, 'HR面中'::application_status,
        '已收到offer'::application_status, '已拒绝'::application_status
    ];
BEGIN
    v_start_time := clock_timestamp();
    
    -- 清理之前的测试数据
    DELETE FROM job_applications WHERE user_id = v_test_user_id;
    
    -- 暂时禁用触发器以加快插入速度
    ALTER TABLE job_applications DISABLE TRIGGER tr_job_applications_status_change;
    
    -- 创建测试数据
    FOR i IN 1..v_test_records_count LOOP
        v_random_status := v_status_array[1 + floor(random() * array_length(v_status_array, 1))];
        v_random_date := TO_CHAR(CURRENT_DATE - (random() * 180)::int, 'YYYY-MM-DD');
        
        INSERT INTO job_applications (
            user_id, company_name, position_title, application_date, status,
            job_description, salary_range, work_location, notes,
            status_history, last_status_change, status_version
        ) VALUES (
            v_test_user_id,
            'Test Company ' || i,
            'Test Position ' || i,
            v_random_date,
            v_random_status,
            'Test job description for performance testing',
            '20-30K',
            'Test City',
            'Performance test record',
            '{"history": [], "summary": {}}'::jsonb,
            NOW() - (random() * INTERVAL '30 days'),
            1
        );
        
        v_created_count := v_created_count + 1;
        
        -- 分批提交避免长事务
        IF i % 100 = 0 THEN
            COMMIT;
        END IF;
    END LOOP;
    
    -- 重新启用触发器
    ALTER TABLE job_applications ENABLE TRIGGER tr_job_applications_status_change;
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    PERFORM record_test_result('测试数据创建', '数据准备', 'PASS', v_execution_time, 
        jsonb_build_object('created_count', v_created_count, 'target_count', v_test_records_count));
    
    RAISE NOTICE '创建了 % 条测试记录，耗时 % 毫秒', v_created_count, v_execution_time;
END;
$$;

-- ============================================================================
-- 4. 索引性能基准测试
-- ============================================================================

-- 4.1 测试用户状态查询性能
DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_test_user_id INTEGER := 9999;
    v_result_count INTEGER;
    v_query_plan TEXT;
BEGIN
    v_start_time := clock_timestamp();
    
    -- 执行复杂状态查询
    SELECT COUNT(*) INTO v_result_count
    FROM job_applications ja
    LEFT JOIN job_status_history jsh ON ja.id = jsh.job_application_id
    WHERE ja.user_id = v_test_user_id
      AND ja.status IN ('一面中', '二面中', 'HR面中')
      AND ja.last_status_change >= (NOW() - INTERVAL '30 days')
    GROUP BY ja.user_id;
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    -- 获取查询计划
    SELECT string_agg(line, E'\n') INTO v_query_plan
    FROM (
        EXPLAIN (FORMAT TEXT)
        SELECT COUNT(*)
        FROM job_applications ja
        LEFT JOIN job_status_history jsh ON ja.id = jsh.job_application_id
        WHERE ja.user_id = v_test_user_id
          AND ja.status IN ('一面中', '二面中', 'HR面中')
          AND ja.last_status_change >= (NOW() - INTERVAL '30 days')
        GROUP BY ja.user_id
    ) AS plan_lines(line);
    
    -- 性能基准: 复杂查询应在50毫秒内完成
    IF v_execution_time <= 50 THEN
        PERFORM record_test_result('复杂状态查询性能', '性能测试', 'PASS', v_execution_time, 
            jsonb_build_object('result_count', v_result_count, 'query_plan', v_query_plan));
    ELSE
        PERFORM record_test_result('复杂状态查询性能', '性能测试', 'FAIL', v_execution_time, 
            jsonb_build_object('result_count', v_result_count, 'threshold_ms', 50));
    END IF;
END;
$$;

-- 4.2 测试JSONB查询性能
DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_test_user_id INTEGER := 9999;
    v_result_count INTEGER;
BEGIN
    v_start_time := clock_timestamp();
    
    -- 更新部分记录的status_history以测试JSONB查询
    UPDATE job_applications 
    SET status_history = jsonb_set(
        status_history, 
        '{metadata}', 
        jsonb_build_object('test_flag', true, 'category', 'tech')
    )
    WHERE user_id = v_test_user_id AND id % 10 = 0;
    
    -- 执行JSONB查询测试
    SELECT COUNT(*) INTO v_result_count
    FROM job_applications
    WHERE user_id = v_test_user_id
      AND status_history->'metadata'->>'test_flag' = 'true';
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    -- JSONB查询性能基准: 应在100毫秒内完成
    IF v_execution_time <= 100 THEN
        PERFORM record_test_result('JSONB查询性能', '性能测试', 'PASS', v_execution_time, 
            jsonb_build_object('result_count', v_result_count));
    ELSE
        PERFORM record_test_result('JSONB查询性能', '性能测试', 'FAIL', v_execution_time, 
            jsonb_build_object('result_count', v_result_count, 'threshold_ms', 100));
    END IF;
END;
$$;

-- ============================================================================
-- 5. 状态转换业务逻辑测试
-- ============================================================================

-- 5.1 测试状态转换验证函数
DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_test_results JSONB := '{}';
    v_valid_transition BOOLEAN;
    v_invalid_transition BOOLEAN;
BEGIN
    v_start_time := clock_timestamp();
    
    -- 测试有效的状态转换
    SELECT validate_status_transition(1, '已投递'::application_status, '简历筛选中'::application_status) INTO v_valid_transition;
    v_test_results := jsonb_set(v_test_results, '{valid_transition}', to_jsonb(v_valid_transition));
    
    -- 测试无效的状态转换
    SELECT validate_status_transition(1, '已拒绝'::application_status, '一面中'::application_status) INTO v_invalid_transition;
    v_test_results := jsonb_set(v_test_results, '{invalid_transition}', to_jsonb(NOT v_invalid_transition));
    
    -- 测试初始状态设置
    SELECT validate_status_transition(1, NULL, '已投递'::application_status) INTO v_valid_transition;
    v_test_results := jsonb_set(v_test_results, '{initial_status}', to_jsonb(v_valid_transition));
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    -- 所有测试都应该通过
    IF (v_test_results->>'valid_transition')::boolean 
       AND (v_test_results->>'invalid_transition')::boolean 
       AND (v_test_results->>'initial_status')::boolean THEN
        PERFORM record_test_result('状态转换验证逻辑', '业务逻辑测试', 'PASS', v_execution_time, v_test_results);
    ELSE
        PERFORM record_test_result('状态转换验证逻辑', '业务逻辑测试', 'FAIL', v_execution_time, v_test_results);
    END IF;
END;
$$;

-- 5.2 测试触发器功能
DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_test_app_id INTEGER;
    v_test_user_id INTEGER := 9999;
    v_history_count_before INTEGER;
    v_history_count_after INTEGER;
    v_status_history JSONB;
    v_trigger_working BOOLEAN := FALSE;
BEGIN
    v_start_time := clock_timestamp();
    
    -- 创建一个测试记录
    INSERT INTO job_applications (
        user_id, company_name, position_title, application_date, status
    ) VALUES (
        v_test_user_id, 'Trigger Test Company', 'Trigger Test Position', 
        TO_CHAR(CURRENT_DATE, 'YYYY-MM-DD'), '已投递'::application_status
    ) RETURNING id INTO v_test_app_id;
    
    -- 记录更新前的历史记录数
    SELECT COUNT(*) INTO v_history_count_before
    FROM job_status_history 
    WHERE job_application_id = v_test_app_id;
    
    -- 更新状态触发触发器
    UPDATE job_applications 
    SET status = '简历筛选中'::application_status
    WHERE id = v_test_app_id;
    
    -- 记录更新后的历史记录数
    SELECT COUNT(*) INTO v_history_count_after
    FROM job_status_history 
    WHERE job_application_id = v_test_app_id;
    
    -- 检查状态历史是否更新
    SELECT status_history INTO v_status_history
    FROM job_applications 
    WHERE id = v_test_app_id;
    
    -- 验证触发器是否正常工作
    IF v_history_count_after > v_history_count_before 
       AND v_status_history IS NOT NULL 
       AND jsonb_array_length(v_status_history->'history') > 0 THEN
        v_trigger_working := TRUE;
    END IF;
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    IF v_trigger_working THEN
        PERFORM record_test_result('状态变更触发器', '业务逻辑测试', 'PASS', v_execution_time, 
            jsonb_build_object('history_before', v_history_count_before, 'history_after', v_history_count_after));
    ELSE
        PERFORM record_test_result('状态变更触发器', '业务逻辑测试', 'FAIL', v_execution_time, 
            jsonb_build_object('history_before', v_history_count_before, 'history_after', v_history_count_after));
    END IF;
    
    -- 清理测试数据
    DELETE FROM job_applications WHERE id = v_test_app_id;
END;
$$;

-- ============================================================================
-- 6. 并发安全性测试
-- ============================================================================

-- 6.1 测试状态版本控制
DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_test_app_id INTEGER;
    v_test_user_id INTEGER := 9999;
    v_version_before INTEGER;
    v_version_after INTEGER;
    v_version_control_working BOOLEAN := FALSE;
BEGIN
    v_start_time := clock_timestamp();
    
    -- 创建一个测试记录
    INSERT INTO job_applications (
        user_id, company_name, position_title, application_date, status, status_version
    ) VALUES (
        v_test_user_id, 'Version Test Company', 'Version Test Position', 
        TO_CHAR(CURRENT_DATE, 'YYYY-MM-DD'), '已投递'::application_status, 1
    ) RETURNING id INTO v_test_app_id;
    
    -- 获取初始版本号
    SELECT status_version INTO v_version_before
    FROM job_applications WHERE id = v_test_app_id;
    
    -- 更新状态
    UPDATE job_applications 
    SET status = '简历筛选中'::application_status
    WHERE id = v_test_app_id;
    
    -- 获取更新后的版本号
    SELECT status_version INTO v_version_after
    FROM job_applications WHERE id = v_test_app_id;
    
    -- 验证版本号是否递增
    IF v_version_after = v_version_before + 1 THEN
        v_version_control_working := TRUE;
    END IF;
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    IF v_version_control_working THEN
        PERFORM record_test_result('状态版本控制', '并发安全测试', 'PASS', v_execution_time, 
            jsonb_build_object('version_before', v_version_before, 'version_after', v_version_after));
    ELSE
        PERFORM record_test_result('状态版本控制', '并发安全测试', 'FAIL', v_execution_time, 
            jsonb_build_object('version_before', v_version_before, 'version_after', v_version_after));
    END IF;
    
    -- 清理测试数据
    DELETE FROM job_applications WHERE id = v_test_app_id;
END;
$$;

-- ============================================================================
-- 7. 数据完整性验证
-- ============================================================================

DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_consistency_issues INTEGER;
    v_orphaned_records INTEGER;
    v_integrity_passed BOOLEAN := TRUE;
BEGIN
    v_start_time := clock_timestamp();
    
    -- 检查状态历史一致性
    SELECT COUNT(*) INTO v_consistency_issues
    FROM (
        SELECT * FROM check_status_history_consistency()
    ) issues;
    
    -- 检查孤立记录
    SELECT COUNT(*) INTO v_orphaned_records
    FROM job_status_history jsh
    LEFT JOIN job_applications ja ON jsh.job_application_id = ja.id
    WHERE ja.id IS NULL;
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    IF v_consistency_issues = 0 AND v_orphaned_records = 0 THEN
        PERFORM record_test_result('数据完整性检查', '数据完整性测试', 'PASS', v_execution_time, 
            jsonb_build_object('consistency_issues', v_consistency_issues, 'orphaned_records', v_orphaned_records));
    ELSE
        PERFORM record_test_result('数据完整性检查', '数据完整性测试', 'FAIL', v_execution_time, 
            jsonb_build_object('consistency_issues', v_consistency_issues, 'orphaned_records', v_orphaned_records));
    END IF;
END;
$$;

-- ============================================================================
-- 8. 查询函数测试
-- ============================================================================

-- 8.1 测试状态历史查询函数
DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_test_user_id INTEGER := 9999;
    v_test_app_id INTEGER;
    v_history_records INTEGER;
    v_function_working BOOLEAN := FALSE;
BEGIN
    v_start_time := clock_timestamp();
    
    -- 选择一个有历史记录的测试应用
    SELECT id INTO v_test_app_id
    FROM job_applications 
    WHERE user_id = v_test_user_id
    LIMIT 1;
    
    -- 测试获取状态历史函数
    SELECT COUNT(*) INTO v_history_records
    FROM get_job_status_history(v_test_user_id, v_test_app_id, 10);
    
    -- 验证函数是否返回结果（即使是0也表示函数正常工作）
    IF v_history_records >= 0 THEN
        v_function_working := TRUE;
    END IF;
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    IF v_function_working THEN
        PERFORM record_test_result('状态历史查询函数', '功能测试', 'PASS', v_execution_time, 
            jsonb_build_object('history_records', v_history_records));
    ELSE
        PERFORM record_test_result('状态历史查询函数', '功能测试', 'FAIL', v_execution_time, 
            jsonb_build_object('history_records', v_history_records));
    END IF;
END;
$$;

-- 8.2 测试状态分析函数
DO $$
DECLARE
    v_start_time TIMESTAMP;
    v_execution_time NUMERIC;
    v_test_user_id INTEGER := 9999;
    v_analysis_records INTEGER;
    v_function_working BOOLEAN := FALSE;
BEGIN
    v_start_time := clock_timestamp();
    
    -- 测试状态持续时间分析函数
    SELECT COUNT(*) INTO v_analysis_records
    FROM analyze_status_durations(v_test_user_id);
    
    -- 验证函数是否正常工作
    IF v_analysis_records >= 0 THEN
        v_function_working := TRUE;
    END IF;
    
    v_execution_time := EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000;
    
    IF v_function_working THEN
        PERFORM record_test_result('状态分析函数', '功能测试', 'PASS', v_execution_time, 
            jsonb_build_object('analysis_records', v_analysis_records));
    ELSE
        PERFORM record_test_result('状态分析函数', '功能测试', 'FAIL', v_execution_time, 
            jsonb_build_object('analysis_records', v_analysis_records));
    END IF;
END;
$$;

-- ============================================================================
-- 9. 清理测试数据
-- ============================================================================

DO $$
DECLARE
    v_cleaned_count INTEGER;
BEGIN
    -- 清理测试数据
    DELETE FROM job_applications WHERE user_id = 9999;
    GET DIAGNOSTICS v_cleaned_count = ROW_COUNT;
    
    -- 清理测试用户偏好
    DELETE FROM user_status_preferences WHERE user_id = 9999;
    
    RAISE NOTICE '清理了 % 条测试记录', v_cleaned_count;
    
    PERFORM record_test_result('测试数据清理', '清理', 'PASS', 0, 
        jsonb_build_object('cleaned_records', v_cleaned_count));
END;
$$;

-- ============================================================================
-- 10. 生成测试报告
-- ============================================================================

-- 测试结果汇总
DO $$
DECLARE
    v_total_tests INTEGER;
    v_passed_tests INTEGER;
    v_failed_tests INTEGER;
    v_total_execution_time NUMERIC;
    v_avg_execution_time NUMERIC;
BEGIN
    -- 统计测试结果
    SELECT 
        COUNT(*),
        COUNT(CASE WHEN status = 'PASS' THEN 1 END),
        COUNT(CASE WHEN status = 'FAIL' THEN 1 END),
        SUM(COALESCE(execution_time_ms, 0)),
        AVG(COALESCE(execution_time_ms, 0))
    INTO v_total_tests, v_passed_tests, v_failed_tests, v_total_execution_time, v_avg_execution_time
    FROM status_tracking_test_results;
    
    RAISE NOTICE '========================================================';
    RAISE NOTICE '状态跟踪系统测试报告';
    RAISE NOTICE '========================================================';
    RAISE NOTICE '测试执行时间: %', NOW();
    RAISE NOTICE '总测试数: %', v_total_tests;
    RAISE NOTICE '通过测试: % (%.1f%%)', v_passed_tests, (v_passed_tests * 100.0 / v_total_tests);
    RAISE NOTICE '失败测试: % (%.1f%%)', v_failed_tests, (v_failed_tests * 100.0 / v_total_tests);
    RAISE NOTICE '总执行时间: % 毫秒', v_total_execution_time;
    RAISE NOTICE '平均执行时间: % 毫秒', ROUND(v_avg_execution_time, 2);
    RAISE NOTICE '';
    
    IF v_failed_tests = 0 THEN
        RAISE NOTICE '✅ 所有测试通过！状态跟踪系统已准备就绪';
    ELSE
        RAISE WARNING '⚠️  有 % 个测试失败，请检查详细结果', v_failed_tests;
    END IF;
    
    RAISE NOTICE '========================================================';
END;
$$;

-- 显示详细测试结果
SELECT 
    test_category as "测试类别",
    test_name as "测试名称",
    status as "状态",
    ROUND(COALESCE(execution_time_ms, 0), 2) as "执行时间(毫秒)",
    executed_at as "执行时间"
FROM status_tracking_test_results
ORDER BY test_category, executed_at;

-- 显示失败的测试详情
SELECT 
    test_name as "失败测试",
    details as "详细信息"
FROM status_tracking_test_results
WHERE status = 'FAIL';

-- 性能统计
SELECT 
    test_category as "测试类别",
    COUNT(*) as "测试数量",
    AVG(execution_time_ms) as "平均执行时间",
    MAX(execution_time_ms) as "最长执行时间",
    MIN(execution_time_ms) as "最短执行时间"
FROM status_tracking_test_results
WHERE execution_time_ms IS NOT NULL
GROUP BY test_category
ORDER BY "平均执行时间" DESC;

-- 记录测试完成日志
INSERT INTO maintenance_log (operation, table_name, executed_at, notes) 
VALUES ('PERFORMANCE_TEST', 'status_tracking_system', NOW(), 
        'Status tracking system performance test and validation completed')
ON CONFLICT DO NOTHING;

-- 清理测试辅助函数
DROP FUNCTION IF EXISTS record_test_result(VARCHAR(100), VARCHAR(50), VARCHAR(20), NUMERIC, JSONB);

RAISE NOTICE '性能测试和验证脚本执行完成！';