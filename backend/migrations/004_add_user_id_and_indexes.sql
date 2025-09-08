-- 添加用户ID字段和数据库查询优化索引
-- 创建时间: 2025-09-07
-- 目的: 实现用户权限隔离和数据库查询性能优化

-- 添加 user_id 字段
ALTER TABLE job_applications 
ADD COLUMN IF NOT EXISTS user_id INTEGER NOT NULL DEFAULT 1;

-- 修改 application_date 为 VARCHAR 类型以兼容现有代码
ALTER TABLE job_applications 
ALTER COLUMN application_date TYPE VARCHAR(10);

-- 修改 follow_up_date 为 VARCHAR 类型以兼容现有代码
ALTER TABLE job_applications 
ALTER COLUMN follow_up_date TYPE VARCHAR(10);

-- 创建高优先级索引 (P0 - 关键索引)

-- 1. 用户ID索引 - 核心权限过滤索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_user_id 
ON job_applications(user_id);

-- 2. 复合索引：用户+投递日期 - 支持 GetAll 查询的排序
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_user_date 
ON job_applications(user_id, application_date DESC);

-- 3. 复合索引：用户+状态 - 支持状态筛选和统计查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_user_status 
ON job_applications(user_id, status);

-- 创建中优先级索引 (P1 - 重要索引)

-- 4. 复合索引：用户+创建时间 - 支持备用排序
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_user_created 
ON job_applications(user_id, created_at DESC);

-- 5. 状态统计覆盖索引 - 优化 GetStatusStatistics 查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_status_stats 
ON job_applications(user_id, status) 
INCLUDE (id);

-- 创建低优先级索引 (P2 - 优化索引)

-- 6. 提醒时间部分索引 - 只为启用提醒的记录创建索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_reminder 
ON job_applications(user_id, reminder_time) 
WHERE reminder_enabled = TRUE AND reminder_time IS NOT NULL;

-- 7. 公司名称搜索索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_company_search 
ON job_applications(user_id, company_name) 
WHERE company_name IS NOT NULL;

-- 8. 面试时间索引 - 支持面试时间查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_applications_interview_user_time 
ON job_applications(user_id, interview_time) 
WHERE interview_time IS NOT NULL;

-- 移除不再需要的旧索引（如果存在）
DROP INDEX IF EXISTS idx_job_applications_application_date;
DROP INDEX IF EXISTS idx_job_applications_status;
DROP INDEX IF EXISTS idx_job_applications_company_name;
DROP INDEX IF EXISTS idx_job_applications_date;
DROP INDEX IF EXISTS idx_job_applications_company;

-- 添加字段注释
COMMENT ON COLUMN job_applications.user_id IS '用户ID，用于权限隔离';

-- 添加表注释
COMMENT ON TABLE job_applications IS '求职申请记录表，已优化索引以提升查询性能';

-- 创建索引使用情况查询视图（用于性能监控）
CREATE OR REPLACE VIEW index_usage_stats AS
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes 
WHERE schemaname = 'public' AND tablename = 'job_applications'
ORDER BY idx_scan DESC;

COMMENT ON VIEW index_usage_stats IS '索引使用情况统计视图，用于监控索引性能';