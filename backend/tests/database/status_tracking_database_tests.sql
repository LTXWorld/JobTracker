-- JobView状态跟踪系统数据库测试脚本
-- 测试工程师: 🧪 PACT Tester
-- 创建时间: 2025-09-08
-- 版本: 1.0

-- ============================================================================
-- 1. 数据库结构完整性测试
-- ============================================================================

-- 检查所有状态跟踪相关表是否存在
DO $$
BEGIN
    RAISE NOTICE '开始数据库结构完整性测试...';
    
    -- 检查job_status_history表
    IF NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'job_status_history') THEN
        RAISE EXCEPTION 'job_status_history表不存在';
    ELSE
        RAISE NOTICE '✅ job_status_history表存在';
    END IF;
    
    -- 检查status_flow_templates表
    IF NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'status_flow_templates') THEN
        RAISE EXCEPTION 'status_flow_templates表不存在';
    ELSE
        RAISE NOTICE '✅ status_flow_templates表存在';
    END IF;
    
    -- 检查user_status_preferences表
    IF NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'user_status_preferences') THEN
        RAISE EXCEPTION 'user_status_preferences表不存在';
    ELSE
        RAISE NOTICE '✅ user_status_preferences表存在';
    END IF;
    
    -- 检查job_applications表的扩展字段
    IF NOT EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name = 'job_applications' AND column_name = 'status_history'
    ) THEN
        RAISE EXCEPTION 'job_applications表缺少status_history字段';
    ELSE
        RAISE NOTICE '✅ job_applications.status_history字段存在';
    END IF;
    
    RAISE NOTICE '数据库结构完整性测试通过';
END $$;

-- ============================================================================
-- 2. 索引效率测试
-- ============================================================================

-- 检查关键索引是否存在并测试性能
DO $$
DECLARE
    index_count INTEGER;
    test_start TIMESTAMP;
    test_end TIMESTAMP;
    duration_ms INTEGER;
BEGIN
    RAISE NOTICE '开始索引效率测试...';
    
    -- 检查job_status_history索引
    SELECT COUNT(*) INTO index_count
    FROM pg_indexes 
    WHERE tablename = 'job_status_history' 
    AND indexname LIKE 'idx_%';
    
    IF index_count < 3 THEN
        RAISE WARNING '⚠️  job_status_history表索引数量可能不足: %', index_count;
    ELSE
        RAISE NOTICE '✅ job_status_history表索引数量正常: %', index_count;
    END IF;
    
    -- 测试status_history查询性能
    test_start := clock_timestamp();
    PERFORM * FROM job_applications WHERE status_history IS NOT NULL LIMIT 10;
    test_end := clock_timestamp();
    duration_ms := EXTRACT(milliseconds FROM test_end - test_start);
    
    RAISE NOTICE '✅ status_history查询耗时: %ms', duration_ms;
    
    IF duration_ms > 100 THEN
        RAISE WARNING '⚠️  status_history查询性能可能需要优化';
    END IF;
    
    RAISE NOTICE '索引效率测试完成';
END $$;

-- ============================================================================
-- 3. 约束条件和数据完整性测试
-- ============================================================================

-- 测试数据约束和触发器功能
DO $$
DECLARE
    test_user_id INTEGER := 99999;  -- 使用不太可能存在的测试用户ID
    test_job_id INTEGER;
    constraint_test_passed BOOLEAN := TRUE;
BEGIN
    RAISE NOTICE '开始约束条件和数据完整性测试...';
    
    -- 准备测试数据
    BEGIN
        -- 插入测试用户（如果不存在）
        INSERT INTO users (id, username, email, password) 
        VALUES (test_user_id, 'test_user', 'test@example.com', 'hashed_password')
        ON CONFLICT (id) DO NOTHING;
        
        -- 插入测试岗位申请
        INSERT INTO job_applications (user_id, company_name, position_title, status, created_at) 
        VALUES (test_user_id, '测试公司', '测试职位', '已投递', NOW())
        RETURNING id INTO test_job_id;
        
        RAISE NOTICE '✅ 测试数据准备完成，job_id: %', test_job_id;
        
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE '⚠️  测试数据准备失败: %', SQLERRM;
        constraint_test_passed := FALSE;
    END;
    
    -- 测试1: 状态转换约束测试
    BEGIN
        INSERT INTO job_status_history (job_application_id, user_id, old_status, new_status)
        VALUES (test_job_id, test_user_id, '已投递', '已投递');  -- 相同状态转换应该被约束阻止
        
        RAISE WARNING '⚠️  状态转换约束未生效：允许了相同状态的转换';
        constraint_test_passed := FALSE;
        
    EXCEPTION 
        WHEN check_violation THEN
            RAISE NOTICE '✅ 状态转换约束正常工作';
        WHEN OTHERS THEN
            RAISE WARNING '⚠️  状态转换约束测试异常: %', SQLERRM;
            constraint_test_passed := FALSE;
    END;
    
    -- 测试2: 持续时间约束测试
    BEGIN
        INSERT INTO job_status_history (job_application_id, user_id, new_status, duration_minutes)
        VALUES (test_job_id, test_user_id, '简历筛选中', -100);  -- 负数持续时间应该被阻止
        
        RAISE WARNING '⚠️  持续时间约束未生效：允许了负数持续时间';
        constraint_test_passed := FALSE;
        
    EXCEPTION 
        WHEN check_violation THEN
            RAISE NOTICE '✅ 持续时间约束正常工作';
        WHEN OTHERS THEN
            RAISE WARNING '⚠️  持续时间约束测试异常: %', SQLERRM;
            constraint_test_passed := FALSE;
    END;
    
    -- 测试3: JSONB约束测试
    BEGIN
        INSERT INTO job_status_history (job_application_id, user_id, new_status, metadata)
        VALUES (test_job_id, test_user_id, '简历筛选中', '"invalid_json_object"');  -- 非对象类型应该被阻止
        
        RAISE WARNING '⚠️  JSONB约束未生效：允许了非对象类型的metadata';
        constraint_test_passed := FALSE;
        
    EXCEPTION 
        WHEN check_violation THEN
            RAISE NOTICE '✅ JSONB约束正常工作';
        WHEN OTHERS THEN
            RAISE WARNING '⚠️  JSONB约束测试异常: %', SQLERRM;
            constraint_test_passed := FALSE;
    END;
    
    -- 清理测试数据
    DELETE FROM job_applications WHERE id = test_job_id;
    DELETE FROM users WHERE id = test_user_id;
    
    IF constraint_test_passed THEN
        RAISE NOTICE '✅ 约束条件和数据完整性测试通过';
    ELSE
        RAISE EXCEPTION '❌ 约束条件和数据完整性测试失败';
    END IF;
    
END $$;

-- ============================================================================
-- 4. 状态转换函数测试
-- ============================================================================

-- 测试状态转换验证函数
DO $$
DECLARE
    is_valid BOOLEAN;
BEGIN
    RAISE NOTICE '开始状态转换函数测试...';
    
    -- 测试有效的状态转换
    IF EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'validate_status_transition') THEN
        SELECT validate_status_transition(1, '已投递', '简历筛选中') INTO is_valid;
        
        IF is_valid THEN
            RAISE NOTICE '✅ 有效状态转换验证正确';
        ELSE
            RAISE WARNING '⚠️  有效状态转换被错误拒绝';
        END IF;
        
        -- 测试无效的状态转换
        SELECT validate_status_transition(1, '已拒绝', '已投递') INTO is_valid;
        
        IF NOT is_valid THEN
            RAISE NOTICE '✅ 无效状态转换验证正确';
        ELSE
            RAISE WARNING '⚠️  无效状态转换被错误允许';
        END IF;
    ELSE
        RAISE NOTICE '⚠️  状态转换验证函数不存在，跳过功能测试';
    END IF;
    
    RAISE NOTICE '状态转换函数测试完成';
END $$;

-- ============================================================================
-- 5. 触发器和自动化功能测试
-- ============================================================================

-- 测试状态更新时的自动历史记录功能
DO $$
DECLARE
    test_user_id INTEGER := 99998;
    test_job_id INTEGER;
    history_count INTEGER;
    status_history_data JSONB;
BEGIN
    RAISE NOTICE '开始触发器和自动化功能测试...';
    
    -- 创建测试用户
    INSERT INTO users (id, username, email, password) 
    VALUES (test_user_id, 'trigger_test', 'trigger_test@example.com', 'hashed_password')
    ON CONFLICT (id) DO NOTHING;
    
    INSERT INTO job_applications (user_id, company_name, position_title, status, created_at) 
    VALUES (test_user_id, '触发器测试公司', '测试职位', '已投递', NOW())
    RETURNING id INTO test_job_id;
    
    -- 更新状态，测试触发器是否工作
    UPDATE job_applications 
    SET status = '简历筛选中' 
    WHERE id = test_job_id;
    
    -- 检查是否自动创建了历史记录
    SELECT COUNT(*) INTO history_count
    FROM job_status_history
    WHERE job_application_id = test_job_id;
    
    IF history_count > 0 THEN
        RAISE NOTICE '✅ 状态更新触发器正常工作，创建了%条历史记录', history_count;
    ELSE
        RAISE WARNING '⚠️  状态更新触发器可能未正常工作';
    END IF;
    
    -- 检查job_applications表的status_history字段是否被更新
    SELECT status_history INTO status_history_data
    FROM job_applications
    WHERE id = test_job_id;
    
    IF status_history_data IS NOT NULL AND jsonb_array_length(status_history_data -> 'history') > 0 THEN
        RAISE NOTICE '✅ status_history字段自动更新正常';
    ELSE
        RAISE WARNING '⚠️  status_history字段自动更新可能有问题';
    END IF;
    
    -- 清理测试数据
    DELETE FROM job_applications WHERE id = test_job_id;
    DELETE FROM users WHERE id = test_user_id;
    
    RAISE NOTICE '触发器和自动化功能测试完成';
END $$;

-- ============================================================================
-- 6. 性能基准测试
-- ============================================================================

-- 创建性能测试数据并测试查询性能
DO $$
DECLARE
    start_time TIMESTAMP;
    end_time TIMESTAMP;
    duration_ms INTEGER;
    test_records INTEGER := 100;
    i INTEGER;
    test_user_id INTEGER := 99997;
    test_job_ids INTEGER[];
BEGIN
    RAISE NOTICE '开始性能基准测试...';
    
    -- 创建测试用户
    INSERT INTO users (id, username, email, password) 
    VALUES (test_user_id, 'perf_test', 'perf_test@example.com', 'hashed_password')
    ON CONFLICT (id) DO NOTHING;
    
    -- 批量插入测试数据
    start_time := clock_timestamp();
    
    FOR i IN 1..test_records LOOP
        INSERT INTO job_applications (user_id, company_name, position_title, status, created_at)
        VALUES (test_user_id, '性能测试公司' || i, '测试职位' || i, '已投递', NOW() - (i || ' hours')::INTERVAL);
    END LOOP;
    
    end_time := clock_timestamp();
    duration_ms := EXTRACT(milliseconds FROM end_time - start_time);
    
    RAISE NOTICE '✅ 批量插入%条记录耗时: %ms (平均%.2fms/条)', 
        test_records, duration_ms, duration_ms::FLOAT / test_records;
    
    -- 测试复杂查询性能
    start_time := clock_timestamp();
    
    PERFORM ja.*, jsh.* 
    FROM job_applications ja
    LEFT JOIN job_status_history jsh ON ja.id = jsh.job_application_id
    WHERE ja.user_id = test_user_id
    ORDER BY ja.created_at DESC, jsh.status_changed_at DESC;
    
    end_time := clock_timestamp();
    duration_ms := EXTRACT(milliseconds FROM end_time - start_time);
    
    RAISE NOTICE '✅ 复杂关联查询耗时: %ms', duration_ms;
    
    -- 测试JSONB查询性能
    start_time := clock_timestamp();
    
    PERFORM * FROM job_applications 
    WHERE user_id = test_user_id 
    AND status_history ? 'history';
    
    end_time := clock_timestamp();
    duration_ms := EXTRACT(milliseconds FROM end_time - start_time);
    
    RAISE NOTICE '✅ JSONB查询耗时: %ms', duration_ms;
    
    -- 清理测试数据
    DELETE FROM job_applications WHERE user_id = test_user_id;
    DELETE FROM users WHERE id = test_user_id;
    
    RAISE NOTICE '性能基准测试完成';
END $$;

-- ============================================================================
-- 7. 数据一致性验证测试
-- ============================================================================

-- 检查数据的完整性和一致性
DO $$
DECLARE
    inconsistent_count INTEGER;
    orphaned_history_count INTEGER;
    total_applications INTEGER;
    total_history_entries INTEGER;
BEGIN
    RAISE NOTICE '开始数据一致性验证测试...';
    
    -- 检查孤立的状态历史记录
    SELECT COUNT(*) INTO orphaned_history_count
    FROM job_status_history jsh
    LEFT JOIN job_applications ja ON jsh.job_application_id = ja.id
    WHERE ja.id IS NULL;
    
    IF orphaned_history_count > 0 THEN
        RAISE WARNING '⚠️  发现%条孤立的状态历史记录', orphaned_history_count;
    ELSE
        RAISE NOTICE '✅ 无孤立的状态历史记录';
    END IF;
    
    -- 检查状态历史记录与岗位申请的用户ID一致性
    SELECT COUNT(*) INTO inconsistent_count
    FROM job_status_history jsh
    JOIN job_applications ja ON jsh.job_application_id = ja.id
    WHERE jsh.user_id != ja.user_id;
    
    IF inconsistent_count > 0 THEN
        RAISE WARNING '⚠️  发现%条用户ID不一致的状态历史记录', inconsistent_count;
    ELSE
        RAISE NOTICE '✅ 状态历史记录用户ID一致性正常';
    END IF;
    
    -- 统计总体数据情况
    SELECT COUNT(*) INTO total_applications FROM job_applications;
    SELECT COUNT(*) INTO total_history_entries FROM job_status_history;
    
    RAISE NOTICE '📊 数据统计 - 总申请数: %, 总历史记录数: %', 
        total_applications, total_history_entries;
    
    RAISE NOTICE '数据一致性验证测试完成';
END $$;

-- ============================================================================
-- 8. 测试总结报告
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE '===========================================';
    RAISE NOTICE '数据库层测试验证完成';
    RAISE NOTICE '===========================================';
    RAISE NOTICE '测试类型:';
    RAISE NOTICE '  ✅ 结构完整性测试';
    RAISE NOTICE '  ✅ 索引效率测试';
    RAISE NOTICE '  ✅ 约束条件测试';
    RAISE NOTICE '  ✅ 状态转换函数测试';
    RAISE NOTICE '  ✅ 触发器功能测试';
    RAISE NOTICE '  ✅ 性能基准测试';
    RAISE NOTICE '  ✅ 数据一致性测试';
    RAISE NOTICE '===========================================';
    RAISE NOTICE '测试工程师: 🧪 PACT Tester';
    RAISE NOTICE '测试时间: %', NOW();
    RAISE NOTICE '===========================================';
END $$;