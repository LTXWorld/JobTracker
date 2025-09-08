-- 数据库查询优化 - 索引优化
-- 创建时间: 2025-09-07
-- 优化目标: 提升查询性能 60-80%

-- 1. 用户ID索引 (关键索引 - 支持所有用户相关查询)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_user_id 
ON job_applications(user_id);

-- 2. 复合索引：用户+投递日期 (支持GetAll排序查询优化)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_user_date 
ON job_applications(user_id, application_date DESC);

-- 3. 复合索引：用户+状态 (支持状态筛选和统计查询)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_user_status 
ON job_applications(user_id, status);

-- 4. 复合索引：用户+创建时间 (支持备用排序)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_user_created 
ON job_applications(user_id, created_at DESC);

-- 5. 状态统计优化索引 (包含索引减少回表)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_status_stats 
ON job_applications(user_id, status) 
INCLUDE (id);

-- 6. 提醒时间部分索引 (只为启用提醒的记录创建索引)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_reminder 
ON job_applications(reminder_time) 
WHERE reminder_enabled = TRUE AND reminder_time IS NOT NULL;

-- 7. 公司名称搜索索引 (支持公司维度查询)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_company_search 
ON job_applications(user_id, company_name) 
WHERE company_name IS NOT NULL;

-- 验证索引创建状态
-- SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'job_applications' ORDER BY indexname;