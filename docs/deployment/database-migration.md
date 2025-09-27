# 🗄️ 数据库迁移脚本

> 完整的数据库结构迁移和字段补全方案

## 📋 迁移概述

此脚本解决之前部署时出现的数据库字段缺失问题，确保所有表结构完整且包含最新功能的字段。

## 🚀 快速执行

```bash
# 下载并执行迁移脚本
cd /www/wwwroot/jobview.bfsmlt.top
wget https://raw.githubusercontent.com/your-repo/jobView/main/scripts/migrate-database.sh
chmod +x migrate-database.sh
./migrate-database.sh
```

## 📝 详细迁移步骤

### 1. 环境准备
```bash
#!/bin/bash
# migrate-database.sh

set -e

echo "🗄️ JobView 数据库迁移脚本"
echo "================================"

# 检查环境变量
if [ -z "$DB_PASSWORD" ]; then
    read -s -p "请输入数据库密码: " DB_PASSWORD
    echo
fi

export PGPASSWORD="$DB_PASSWORD"
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-jobview_user}
DB_NAME=${DB_NAME:-jobview_prod}

# 测试数据库连接
echo "测试数据库连接..."
if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ 数据库连接失败，请检查配置"
    exit 1
fi
echo "✅ 数据库连接成功"
```

### 2. 备份现有数据
```bash
# 创建备份
BACKUP_FILE="/backup/jobview_migration_$(date +%Y%m%d_%H%M%S).sql"
mkdir -p /backup

echo "📦 创建数据库备份..."
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME > $BACKUP_FILE
echo "✅ 备份完成: $BACKUP_FILE"

# 验证备份
if [ ! -s "$BACKUP_FILE" ]; then
    echo "❌ 备份文件为空，中止迁移"
    exit 1
fi
```

### 3. 检查现有表结构
```bash
echo "🔍 检查现有表结构..."

# 创建检查脚本
cat > check_tables.sql << 'EOF'
-- 检查现有表
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;

-- 检查users表结构
\d users

-- 检查job_applications表结构
\d job_applications

-- 检查是否存在扩展表
SELECT EXISTS (
    SELECT FROM information_schema.tables
    WHERE table_schema = 'public'
    AND table_name = 'status_history'
) as status_history_exists;

SELECT EXISTS (
    SELECT FROM information_schema.tables
    WHERE table_schema = 'public'
    AND table_name = 'reminders'
) as reminders_exists;

SELECT EXISTS (
    SELECT FROM information_schema.tables
    WHERE table_schema = 'public'
    AND table_name = 'user_preferences'
) as user_preferences_exists;
EOF

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f check_tables.sql
```

### 4. 完整表结构迁移
```bash
echo "🏗️ 开始表结构迁移..."

# 创建完整的表结构脚本
cat > migration_001_tables.sql << 'EOF'
-- ========================================
-- JobView 完整表结构迁移脚本
-- 版本: v2.5.0
-- 日期: 2025-01-21
-- ========================================

-- 开始事务
BEGIN;

-- 1. 用户表 (确保字段完整)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    full_name VARCHAR(100),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 检查并添加缺失的用户表字段
DO $$
BEGIN
    -- 检查avatar_url字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='users' AND column_name='avatar_url') THEN
        ALTER TABLE users ADD COLUMN avatar_url VARCHAR(500);
        RAISE NOTICE '✅ 添加users.avatar_url字段';
    END IF;

    -- 检查full_name字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='users' AND column_name='full_name') THEN
        ALTER TABLE users ADD COLUMN full_name VARCHAR(100);
        RAISE NOTICE '✅ 添加users.full_name字段';
    END IF;

    -- 检查phone字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='users' AND column_name='phone') THEN
        ALTER TABLE users ADD COLUMN phone VARCHAR(20);
        RAISE NOTICE '✅ 添加users.phone字段';
    END IF;
END
$$;

-- 2. 投递记录表 (完整字段，防止缺失)
CREATE TABLE IF NOT EXISTS job_applications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    position_title VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT '已投递',
    application_date DATE NOT NULL DEFAULT CURRENT_DATE,

    -- 基础信息字段
    salary_range VARCHAR(100),
    work_location VARCHAR(255),
    interview_time TIMESTAMP,
    notes TEXT,

    -- 提醒相关字段
    reminder_enabled BOOLEAN DEFAULT FALSE,
    reminder_time TIMESTAMP,
    follow_up_date DATE,

    -- 状态跟踪相关字段 (关键新增字段)
    last_status_change TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status_version INTEGER DEFAULT 1,
    status_history JSONB DEFAULT '{}',
    status_duration_stats JSONB DEFAULT '{}',

    -- 扩展信息字段
    company_type VARCHAR(50),
    position_type VARCHAR(50),
    application_source VARCHAR(100),
    hr_contact VARCHAR(255),
    job_description TEXT,
    requirements TEXT,
    benefits TEXT,

    -- 元数据
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 检查并添加缺失的投递记录表字段
DO $$
DECLARE
    missing_columns TEXT[] := ARRAY[
        'last_status_change', 'status_version', 'status_history', 'status_duration_stats',
        'company_type', 'position_type', 'application_source', 'hr_contact',
        'job_description', 'requirements', 'benefits'
    ];
    col_name TEXT;
BEGIN
    FOREACH col_name IN ARRAY missing_columns
    LOOP
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                       WHERE table_name='job_applications' AND column_name=col_name) THEN
            CASE col_name
                WHEN 'last_status_change' THEN
                    ALTER TABLE job_applications ADD COLUMN last_status_change TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
                WHEN 'status_version' THEN
                    ALTER TABLE job_applications ADD COLUMN status_version INTEGER DEFAULT 1;
                WHEN 'status_history' THEN
                    ALTER TABLE job_applications ADD COLUMN status_history JSONB DEFAULT '{}';
                WHEN 'status_duration_stats' THEN
                    ALTER TABLE job_applications ADD COLUMN status_duration_stats JSONB DEFAULT '{}';
                WHEN 'company_type' THEN
                    ALTER TABLE job_applications ADD COLUMN company_type VARCHAR(50);
                WHEN 'position_type' THEN
                    ALTER TABLE job_applications ADD COLUMN position_type VARCHAR(50);
                WHEN 'application_source' THEN
                    ALTER TABLE job_applications ADD COLUMN application_source VARCHAR(100);
                WHEN 'hr_contact' THEN
                    ALTER TABLE job_applications ADD COLUMN hr_contact VARCHAR(255);
                WHEN 'job_description' THEN
                    ALTER TABLE job_applications ADD COLUMN job_description TEXT;
                WHEN 'requirements' THEN
                    ALTER TABLE job_applications ADD COLUMN requirements TEXT;
                WHEN 'benefits' THEN
                    ALTER TABLE job_applications ADD COLUMN benefits TEXT;
            END CASE;
            RAISE NOTICE '✅ 添加job_applications.%字段', col_name;
        END IF;
    END LOOP;
END
$$;

-- 3. 状态历史记录表
CREATE TABLE IF NOT EXISTS status_history (
    id SERIAL PRIMARY KEY,
    job_application_id INTEGER NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    status_changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    duration_minutes INTEGER,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. 提醒表
CREATE TABLE IF NOT EXISTS reminders (
    id SERIAL PRIMARY KEY,
    application_id INTEGER NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('interview', 'follow_up')),
    reminder_time TIMESTAMP NOT NULL,
    interview_time TIMESTAMP,
    message TEXT,
    is_dismissed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. 用户偏好设置表
CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference_config JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- 提交事务
COMMIT;

-- 输出成功信息
\echo '✅ 表结构迁移完成'
EOF

# 执行表结构迁移
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migration_001_tables.sql
```

### 5. 索引优化迁移
```bash
echo "📊 创建性能优化索引..."

cat > migration_002_indexes.sql << 'EOF'
-- ========================================
-- JobView 索引优化脚本
-- 版本: v2.5.0
-- ========================================

BEGIN;

-- 基础索引
DROP INDEX IF EXISTS idx_job_applications_user_id;
CREATE INDEX idx_job_applications_user_id ON job_applications(user_id);

DROP INDEX IF EXISTS idx_job_applications_status;
CREATE INDEX idx_job_applications_status ON job_applications(status);

DROP INDEX IF EXISTS idx_job_applications_date;
CREATE INDEX idx_job_applications_date ON job_applications(application_date DESC);

-- 复合索引 (关键性能优化)
DROP INDEX IF EXISTS idx_job_applications_user_date;
CREATE INDEX idx_job_applications_user_date ON job_applications(user_id, application_date DESC);

DROP INDEX IF EXISTS idx_job_applications_user_status;
CREATE INDEX idx_job_applications_user_status ON job_applications(user_id, status);

DROP INDEX IF EXISTS idx_job_applications_user_company;
CREATE INDEX idx_job_applications_user_company ON job_applications(user_id, company_name);

-- 覆盖索引 (避免回表查询)
DROP INDEX IF EXISTS idx_job_applications_status_stats;
CREATE INDEX idx_job_applications_status_stats ON job_applications(user_id, status)
INCLUDE (id, company_name, position_title, application_date);

-- 部分索引 (提醒功能优化)
DROP INDEX IF EXISTS idx_job_applications_reminder;
CREATE INDEX idx_job_applications_reminder ON job_applications(reminder_time)
WHERE reminder_enabled = TRUE AND reminder_time IS NOT NULL;

-- 全文搜索索引
DROP INDEX IF EXISTS idx_job_applications_search;
CREATE INDEX idx_job_applications_search ON job_applications
USING gin(to_tsvector('simple',
    company_name || ' ' ||
    position_title || ' ' ||
    COALESCE(work_location, '') || ' ' ||
    COALESCE(notes, '')
));

-- 状态历史表索引
DROP INDEX IF EXISTS idx_status_history_application;
CREATE INDEX idx_status_history_application ON status_history(job_application_id, status_changed_at DESC);

DROP INDEX IF EXISTS idx_status_history_user;
CREATE INDEX idx_status_history_user ON status_history(user_id, status_changed_at DESC);

-- 提醒表索引
DROP INDEX IF EXISTS idx_reminders_user_time;
CREATE INDEX idx_reminders_user_time ON reminders(user_id, reminder_time);

DROP INDEX IF EXISTS idx_reminders_active;
CREATE INDEX idx_reminders_active ON reminders(reminder_time)
WHERE is_dismissed = FALSE;

-- 用户表索引
DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX idx_users_email ON users(email);

DROP INDEX IF EXISTS idx_users_username;
CREATE UNIQUE INDEX idx_users_username ON users(username);

COMMIT;

\echo '✅ 索引优化完成'
EOF

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migration_002_indexes.sql
```

### 6. 数据完整性检查和修复
```bash
echo "🔍 数据完整性检查..."

cat > migration_003_data_fix.sql << 'EOF'
-- ========================================
-- JobView 数据完整性修复脚本
-- ========================================

BEGIN;

-- 1. 修复缺失的状态版本号
UPDATE job_applications
SET status_version = 1
WHERE status_version IS NULL;

-- 2. 修复缺失的最后状态变更时间
UPDATE job_applications
SET last_status_change = updated_at
WHERE last_status_change IS NULL;

-- 3. 初始化空的状态历史
UPDATE job_applications
SET status_history = '{"history": [], "metadata": {"total_changes": 0, "current_status": "' || status || '", "last_changed": "' || COALESCE(last_status_change, created_at)::text || '", "total_duration_minutes": 0}}'
WHERE status_history = '{}' OR status_history IS NULL;

-- 4. 修复缺失的application_date
UPDATE job_applications
SET application_date = created_at::date
WHERE application_date IS NULL;

-- 5. 为现有用户创建默认偏好设置
INSERT INTO user_preferences (user_id, preference_config)
SELECT id, '{
    "dashboard_layout": "standard",
    "default_time_range": "month",
    "notification_enabled": true,
    "chart_style": "modern",
    "auto_refresh_interval": 300,
    "theme_preference": "light"
}'
FROM users
WHERE id NOT IN (SELECT user_id FROM user_preferences);

-- 6. 为现有记录创建状态历史
INSERT INTO status_history (job_application_id, user_id, new_status, status_changed_at)
SELECT id, user_id, status, COALESCE(last_status_change, created_at)
FROM job_applications
WHERE id NOT IN (
    SELECT DISTINCT job_application_id
    FROM status_history
);

COMMIT;

\echo '✅ 数据完整性修复完成'
EOF

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migration_003_data_fix.sql
```

### 7. 创建默认数据
```bash
echo "📝 创建默认数据..."

cat > migration_004_default_data.sql << 'EOF'
-- ========================================
-- JobView 默认数据脚本
-- ========================================

BEGIN;

-- 创建测试用户 (如果不存在)
-- 密码: TestPass123!
INSERT INTO users (username, email, password_hash, full_name) VALUES
('testuser', 'test@jobview.com', '$2a$10$k8Y1THPD8eRKQVbFdoYFRu.9xhqHdF7YNJ3/dFjDxQSQx4lCpZmKO', '测试用户')
ON CONFLICT (username) DO NOTHING;

-- 为测试用户创建偏好设置
INSERT INTO user_preferences (user_id, preference_config)
SELECT u.id, '{
    "dashboard_layout": "standard",
    "default_time_range": "month",
    "notification_enabled": true,
    "chart_style": "modern",
    "auto_refresh_interval": 300,
    "theme_preference": "light"
}'
FROM users u
WHERE u.username = 'testuser'
AND NOT EXISTS (
    SELECT 1 FROM user_preferences up WHERE up.user_id = u.id
);

COMMIT;

\echo '✅ 默认数据创建完成'
EOF

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migration_004_default_data.sql
```

### 8. 验证迁移结果
```bash
echo "🧪 验证迁移结果..."

cat > verify_migration.sql << 'EOF'
-- 验证表结构
SELECT
    table_name,
    (SELECT count(*) FROM information_schema.columns WHERE table_name = t.table_name) as column_count
FROM information_schema.tables t
WHERE table_schema = 'public'
ORDER BY table_name;

-- 验证关键字段
SELECT
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'job_applications'
AND column_name IN ('status_version', 'status_history', 'last_status_change')
ORDER BY column_name;

-- 验证数据完整性
SELECT
    'users' as table_name,
    count(*) as record_count
FROM users
UNION ALL
SELECT
    'job_applications',
    count(*)
FROM job_applications
UNION ALL
SELECT
    'status_history',
    count(*)
FROM status_history
UNION ALL
SELECT
    'reminders',
    count(*)
FROM reminders
UNION ALL
SELECT
    'user_preferences',
    count(*)
FROM user_preferences;

-- 验证索引
SELECT
    indexname,
    tablename
FROM pg_indexes
WHERE schemaname = 'public'
AND indexname LIKE 'idx_%'
ORDER BY tablename, indexname;

\echo '✅ 迁移验证完成'
EOF

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f verify_migration.sql
```

### 9. 清理临时文件
```bash
echo "🧹 清理临时文件..."

# 删除临时SQL文件
rm -f check_tables.sql
rm -f migration_001_tables.sql
rm -f migration_002_indexes.sql
rm -f migration_003_data_fix.sql
rm -f migration_004_default_data.sql
rm -f verify_migration.sql

echo "✅ 数据库迁移完成!"
echo "📊 迁移统计:"
echo "  - 备份文件: $BACKUP_FILE"
echo "  - 表结构: 已更新"
echo "  - 索引优化: 已完成"
echo "  - 数据修复: 已完成"
echo ""
echo "🔍 请验证应用功能是否正常"
```

## 📋 完整迁移脚本

```bash
#!/bin/bash
# complete-migration.sh

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🗄️ JobView 数据库完整迁移脚本${NC}"
echo "=================================="

# 1. 环境检查
echo -e "${YELLOW}1. 检查环境...${NC}"
if [ -z "$DB_PASSWORD" ]; then
    read -s -p "请输入数据库密码: " DB_PASSWORD
    echo
fi

export PGPASSWORD="$DB_PASSWORD"
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-jobview_user}
DB_NAME=${DB_NAME:-jobview_prod}

# 测试连接
if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}❌ 数据库连接失败${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 数据库连接成功${NC}"

# 2. 备份
echo -e "${YELLOW}2. 创建备份...${NC}"
BACKUP_FILE="/backup/jobview_migration_$(date +%Y%m%d_%H%M%S).sql"
mkdir -p /backup
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME > $BACKUP_FILE
echo -e "${GREEN}✅ 备份完成: $BACKUP_FILE${NC}"

# 3. 执行迁移
echo -e "${YELLOW}3. 执行表结构迁移...${NC}"
# (这里包含所有上面的迁移SQL代码)

echo -e "${YELLOW}4. 创建性能索引...${NC}"
# (索引创建代码)

echo -e "${YELLOW}5. 修复数据完整性...${NC}"
# (数据修复代码)

echo -e "${YELLOW}6. 创建默认数据...${NC}"
# (默认数据代码)

echo -e "${YELLOW}7. 验证迁移结果...${NC}"
# (验证代码)

echo -e "${GREEN}🎉 迁移完成!${NC}"
echo "备份文件: $BACKUP_FILE"
echo "请重启应用并验证功能"
```

---

**🗄️ 使用此脚本可以确保数据库结构完整，解决字段缺失问题！**

> **重要提醒**:
> 1. 迁移前务必备份数据库
> 2. 在低峰期执行迁移操作
> 3. 迁移后验证所有功能正常
> 4. 保留备份文件以备回滚