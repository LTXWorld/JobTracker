-- 添加企业属性字段：company_attribute，取值限定为 '央国企' 或 '私企'
-- 兼容已有数据：允许为空，便于老数据后续手动补充

ALTER TABLE job_applications
ADD COLUMN IF NOT EXISTS company_attribute VARCHAR(10);

-- 添加检查约束（若不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'job_applications' AND constraint_name = 'chk_company_attribute_valid'
    ) THEN
        ALTER TABLE job_applications
        ADD CONSTRAINT chk_company_attribute_valid
        CHECK (company_attribute IS NULL OR company_attribute IN ('央国企', '私企'));
    END IF;
END $$;

COMMENT ON COLUMN job_applications.company_attribute IS '企业属性：央国企/私企（可为空，老数据兼容）';

