-- 创建求职投递记录表
CREATE TABLE IF NOT EXISTS job_applications (
    id SERIAL PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    position_title VARCHAR(255) NOT NULL,
    application_date VARCHAR(10) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT '已投递',
    job_description TEXT,
    salary_range VARCHAR(100),
    work_location VARCHAR(255),
    contact_info VARCHAR(500),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_job_applications_application_date ON job_applications(application_date);
CREATE INDEX IF NOT EXISTS idx_job_applications_status ON job_applications(status);
CREATE INDEX IF NOT EXISTS idx_job_applications_company_name ON job_applications(company_name);