-- 秋招岗位投递记录系统数据库初始化脚本
-- 创建时间: 2025-09-05

-- 创建投递状态枚举类型
CREATE TYPE application_status AS ENUM (
    '已投递',
    '简历筛选中',
    '笔试中',
    '笔试通过',
    '一面中',
    '一面通过',
    '二面中', 
    '二面通过',
    '三面中',
    '三面通过',
    'HR面中',
    'HR面通过',
    '待发offer',
    '已拒绝',
    '已收到offer',
    '已接受offer',
    '流程结束'
);

-- 创建投递记录表
CREATE TABLE job_applications (
    id SERIAL PRIMARY KEY,
    company_name VARCHAR(100) NOT NULL, -- 公司名称
    position_title VARCHAR(100) NOT NULL, -- 职位名称  
    application_date DATE NOT NULL DEFAULT CURRENT_DATE, -- 投递日期
    status application_status NOT NULL DEFAULT '已投递', -- 投递状态
    job_description TEXT, -- 岗位描述
    salary_range VARCHAR(50), -- 薪资范围
    work_location VARCHAR(100), -- 工作地点
    contact_info VARCHAR(200), -- 联系方式
    notes TEXT, -- 备注信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP -- 更新时间
);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建触发器，自动更新updated_at字段
CREATE TRIGGER update_job_applications_updated_at 
    BEFORE UPDATE ON job_applications 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 创建索引优化查询性能
CREATE INDEX idx_job_applications_company ON job_applications(company_name);
CREATE INDEX idx_job_applications_status ON job_applications(status);
CREATE INDEX idx_job_applications_date ON job_applications(application_date DESC);

-- 插入测试数据
INSERT INTO job_applications (company_name, position_title, application_date, status, salary_range, work_location, notes) VALUES
('阿里巴巴', '后端开发工程师', '2024-09-01', '已投递', '20-35K', '杭州', '通过官网投递'),
('腾讯', '前端开发工程师', '2024-09-02', '笔试中', '18-30K', '深圳', '朋友内推'),
('字节跳动', 'Go开发工程师', '2024-09-03', '一面中', '25-40K', '北京', 'BOSS直聘沟通');