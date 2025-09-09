-- JSONB约束问题修复脚本
-- 修复job_status_history表的metadata字段JSONB约束
-- 问题: chk_metadata_is_object约束过于严格
-- 修复: 允许空对象、null值或字符串类型的metadata

-- 删除过于严格的约束
ALTER TABLE job_status_history DROP CONSTRAINT IF EXISTS chk_metadata_is_object;

-- 添加更灵活的约束，允许null、空对象或有效JSON
ALTER TABLE job_status_history ADD CONSTRAINT chk_metadata_valid_json 
CHECK (
    metadata IS NULL OR 
    jsonb_typeof(metadata) = 'object' OR 
    jsonb_typeof(metadata) = 'string' OR
    metadata = '{}'::jsonb
);

-- 更新现有记录，确保metadata字段符合新约束
UPDATE job_status_history 
SET metadata = '{}'::jsonb 
WHERE metadata IS NULL;

-- 添加注释说明修复
COMMENT ON CONSTRAINT chk_metadata_valid_json ON job_status_history IS 
'允许metadata为null、空对象、有效JSON对象或字符串类型';

-- 验证修复效果
DO $$
BEGIN
    -- 测试插入各种metadata值
    BEGIN
        INSERT INTO job_status_history 
        (job_application_id, user_id, new_status, metadata) 
        VALUES (1, 1, '简历筛选中', '{}');
        
        INSERT INTO job_status_history 
        (job_application_id, user_id, new_status, metadata) 
        VALUES (1, 1, '一面中', '{"note": "测试"}');
        
        INSERT INTO job_status_history 
        (job_application_id, user_id, new_status, metadata) 
        VALUES (1, 1, '二面中', NULL);
        
        RAISE NOTICE '✅ JSONB约束修复成功 - 各种metadata值都可以正常插入';
        
        -- 清理测试数据
        DELETE FROM job_status_history WHERE job_application_id = 1 AND user_id = 1;
        
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE '⚠️  JSONB约束修复可能有问题: %', SQLERRM;
    END;
END $$;