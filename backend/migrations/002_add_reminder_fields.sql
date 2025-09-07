-- 添加面试时间和提醒相关字段
-- 创建时间: 2025-09-06

-- 添加新字段到job_applications表
ALTER TABLE job_applications 
ADD COLUMN IF NOT EXISTS interview_time TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS reminder_time TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS reminder_enabled BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS follow_up_date DATE,
ADD COLUMN IF NOT EXISTS hr_name VARCHAR(100),
ADD COLUMN IF NOT EXISTS hr_phone VARCHAR(50),
ADD COLUMN IF NOT EXISTS hr_email VARCHAR(100),
ADD COLUMN IF NOT EXISTS interview_location VARCHAR(200),
ADD COLUMN IF NOT EXISTS interview_type VARCHAR(50); -- 现场/视频/电话

-- 添加索引以优化查询
CREATE INDEX IF NOT EXISTS idx_job_applications_interview_time ON job_applications(interview_time);
CREATE INDEX IF NOT EXISTS idx_job_applications_reminder_time ON job_applications(reminder_time);
CREATE INDEX IF NOT EXISTS idx_job_applications_reminder_enabled ON job_applications(reminder_enabled);

-- 添加注释
COMMENT ON COLUMN job_applications.interview_time IS '面试时间';
COMMENT ON COLUMN job_applications.reminder_time IS '提醒时间';
COMMENT ON COLUMN job_applications.reminder_enabled IS '是否启用提醒';
COMMENT ON COLUMN job_applications.follow_up_date IS '跟进日期';
COMMENT ON COLUMN job_applications.hr_name IS 'HR姓名';
COMMENT ON COLUMN job_applications.hr_phone IS 'HR电话';
COMMENT ON COLUMN job_applications.hr_email IS 'HR邮箱';
COMMENT ON COLUMN job_applications.interview_location IS '面试地点';
COMMENT ON COLUMN job_applications.interview_type IS '面试类型';