# 🚀 JobView 生产环境部署指南

> 从本地到生产服务器的完整部署方案

## 📋 目录

- [🏗️ 部署架构](#部署架构)
- [⚙️ 环境准备](#环境准备)
- [🗄️ 数据库配置](#数据库配置)
- [🔧 路径配置](#路径配置)
- [📦 前端部署](#前端部署)
- [⚡ 后端部署](#后端部署)
- [🌐 Nginx配置](#nginx配置)
- [🔍 部署验证](#部署验证)
- [🔄 更新流程](#更新流程)

## 🏗️ 部署架构

### 生产环境架构图
```
用户浏览器
    ↓ HTTPS (443)
Nginx (jobview.bfsmlt.top)
    ├── / → 静态文件 (前端)
    └── /api/ → 反向代理到后端 (8010)
        ↓
Go后端服务 (localhost:8010)
    ↓
PostgreSQL数据库 (localhost:5432)
```

### 关键配置原则
- **前端**: 静态文件部署，支持SPA路由
- **后端**: 本地运行，通过反向代理访问
- **API路径**: 统一使用 `/api` 前缀，适配本地和线上环境
- **数据库**: 完整迁移，确保字段完整性

## ⚙️ 环境准备

### 服务器要求
- **操作系统**: Linux (CentOS/Ubuntu)
- **内存**: 最低2GB，推荐4GB
- **存储**: 最低20GB可用空间
- **网络**: 稳定的公网连接

### 软件依赖安装
```bash
# 1. 更新系统
yum update -y  # CentOS
# 或
apt update && apt upgrade -y  # Ubuntu

# 2. 安装Node.js 18+
curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
yum install -y nodejs  # CentOS
# 或
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt-get install -y nodejs  # Ubuntu

# 3. 安装Go 1.24+
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile

# 4. 安装PostgreSQL
yum install -y postgresql postgresql-server  # CentOS
# 或
apt install -y postgresql postgresql-contrib  # Ubuntu

# 5. 安装PM2 (进程管理)
npm install -g pm2

# 6. 验证安装
node --version
go version
psql --version
```

## 🗄️ 数据库配置

### 1. PostgreSQL初始化和配置
```bash
# 初始化数据库 (仅首次)
postgresql-setup initdb  # CentOS
# 或
sudo -u postgres initdb /var/lib/postgresql/data  # Ubuntu

# 启动PostgreSQL服务
systemctl start postgresql
systemctl enable postgresql

# 配置PostgreSQL
sudo -u postgres psql
```

### 2. 创建数据库和用户
```sql
-- 创建数据库
CREATE DATABASE jobview_prod;

-- 创建用户
CREATE USER jobview_user WITH PASSWORD 'your_secure_password_here';

-- 授权
GRANT ALL PRIVILEGES ON DATABASE jobview_prod TO jobview_user;
GRANT CREATE ON SCHEMA public TO jobview_user;
GRANT USAGE ON SCHEMA public TO jobview_user;

-- 退出
\q
```

### 3. 数据库迁移 (关键步骤)
```bash
# 创建迁移脚本目录
mkdir -p /www/wwwroot/jobview.bfsmlt.top/db-migration

# 创建完整的数据库结构脚本
cat > /www/wwwroot/jobview.bfsmlt.top/db-migration/001_initial_schema.sql << 'EOF'
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 投递记录表 (完整字段)
CREATE TABLE IF NOT EXISTS job_applications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    position_title VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT '已投递',
    application_date DATE NOT NULL DEFAULT CURRENT_DATE,
    salary_range VARCHAR(100),
    work_location VARCHAR(255),
    interview_time TIMESTAMP,
    notes TEXT,

    -- 提醒相关字段
    reminder_enabled BOOLEAN DEFAULT FALSE,
    reminder_time TIMESTAMP,
    follow_up_date DATE,

    -- 状态跟踪相关字段 (新增，防止迁移时缺失)
    last_status_change TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status_version INTEGER DEFAULT 1,
    status_history JSONB DEFAULT '{}',
    status_duration_stats JSONB DEFAULT '{}',

    -- 扩展字段
    company_type VARCHAR(50),
    position_type VARCHAR(50),
    application_source VARCHAR(100),
    hr_contact VARCHAR(255),
    job_description TEXT,
    requirements TEXT,
    benefits TEXT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 状态历史记录表
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

-- 提醒表
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

-- 用户偏好设置表
CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference_config JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);
EOF

# 创建索引优化脚本
cat > /www/wwwroot/jobview.bfsmlt.top/db-migration/002_create_indexes.sql << 'EOF'
-- 基础索引
CREATE INDEX IF NOT EXISTS idx_job_applications_user_id ON job_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_status ON job_applications(status);
CREATE INDEX IF NOT EXISTS idx_job_applications_date ON job_applications(application_date DESC);

-- 复合索引 (性能优化关键)
CREATE INDEX IF NOT EXISTS idx_job_applications_user_date ON job_applications(user_id, application_date DESC);
CREATE INDEX IF NOT EXISTS idx_job_applications_user_status ON job_applications(user_id, status);
CREATE INDEX IF NOT EXISTS idx_job_applications_user_company ON job_applications(user_id, company_name);

-- 覆盖索引 (避免回表查询)
CREATE INDEX IF NOT EXISTS idx_job_applications_status_stats ON job_applications(user_id, status) INCLUDE (id, company_name, position_title);

-- 部分索引 (提醒功能优化)
CREATE INDEX IF NOT EXISTS idx_job_applications_reminder ON job_applications(reminder_time)
WHERE reminder_enabled = TRUE AND reminder_time IS NOT NULL;

-- 全文搜索索引
CREATE INDEX IF NOT EXISTS idx_job_applications_search ON job_applications
USING gin(to_tsvector('simple', company_name || ' ' || position_title || ' ' || COALESCE(work_location, '')));

-- 状态历史索引
CREATE INDEX IF NOT EXISTS idx_status_history_application ON status_history(job_application_id, status_changed_at DESC);
CREATE INDEX IF NOT EXISTS idx_status_history_user ON status_history(user_id, status_changed_at DESC);

-- 提醒索引
CREATE INDEX IF NOT EXISTS idx_reminders_user_time ON reminders(user_id, reminder_time);
CREATE INDEX IF NOT EXISTS idx_reminders_active ON reminders(reminder_time) WHERE is_dismissed = FALSE;
EOF

# 创建默认数据脚本
cat > /www/wwwroot/jobview.bfsmlt.top/db-migration/003_default_data.sql << 'EOF'
-- 创建测试用户 (密码: TestPass123!)
INSERT INTO users (username, email, password_hash) VALUES
('testuser', 'test@jobview.com', '$2a$10$k8Y1THPD8eRKQVbFdoYFRu.9xhqHdF7YNJ3/dFjDxQSQx4lCpZmKO')
ON CONFLICT (username) DO NOTHING;

-- 插入状态配置数据
-- 这里可以根据需要添加初始配置数据
EOF

# 执行迁移
export PGPASSWORD='your_secure_password_here'

echo "开始数据库迁移..."
psql -h localhost -U jobview_user -d jobview_prod -f /www/wwwroot/jobview.bfsmlt.top/db-migration/001_initial_schema.sql
psql -h localhost -U jobview_user -d jobview_prod -f /www/wwwroot/jobview.bfsmlt.top/db-migration/002_create_indexes.sql
psql -h localhost -U jobview_user -d jobview_prod -f /www/wwwroot/jobview.bfsmlt.top/db-migration/003_default_data.sql

echo "迁移完成，验证表结构..."
psql -h localhost -U jobview_user -d jobview_prod -c "\dt"
psql -h localhost -U jobview_user -d jobview_prod -c "\d job_applications"
```

### 4. 数据库连接配置
```bash
# 配置PostgreSQL允许本地连接
sudo vi /var/lib/pgsql/data/pg_hba.conf
# 确保包含以下行:
# local   all             all                                     md5
# host    all             all             127.0.0.1/32            md5

# 重启PostgreSQL
systemctl restart postgresql
```

## 🔧 路径配置

### 1. 环境变量配置 (关键)
```bash
# 创建生产环境配置文件
cat > /www/wwwroot/jobview.bfsmlt.top/.env.production << 'EOF'
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=jobview_user
DB_PASSWORD=your_secure_password_here
DB_NAME=jobview_prod

# JWT配置 (必须32字符以上)
JWT_SECRET=your-super-secret-jwt-key-must-be-at-least-32-characters-long

# 环境配置
ENVIRONMENT=production
GIN_MODE=release

# 服务配置
PORT=8010
HOST=127.0.0.1

# 文件上传路径
UPLOAD_PATH=/www/wwwroot/jobview.bfsmlt.top/uploads

# 日志配置
LOG_LEVEL=info
LOG_PATH=/www/wwwroot/jobview.bfsmlt.top/logs
EOF

# 设置权限
chmod 600 /www/wwwroot/jobview.bfsmlt.top/.env.production
```

### 2. 前端环境配置
```bash
# 创建生产环境前端配置
cat > /www/wwwroot/jobview.bfsmlt.top/frontend/.env.production << 'EOF'
# API 基础路径 (关键配置)
VITE_API_BASE=/api

# 生产环境标识
NODE_ENV=production

# 构建优化
VITE_BUILD_TARGET=es2015
VITE_LEGACY=true
EOF
```

## 📦 前端部署

### 1. 代码上传和构建
```bash
# 创建项目目录
mkdir -p /www/wwwroot/jobview.bfsmlt.top
cd /www/wwwroot/jobview.bfsmlt.top

# 方式一: Git克隆 (推荐)
git clone https://github.com/your-repo/jobView.git .

# 方式二: 本地上传 (使用scp)
# 在本地执行:
# scp -r ./jobView/* root@your-server-ip:/www/wwwroot/jobview.bfsmlt.top/

# 设置权限
chown -R www:www /www/wwwroot/jobview.bfsmlt.top
chmod -R 755 /www/wwwroot/jobview.bfsmlt.top
```

### 2. 前端构建
```bash
cd /www/wwwroot/jobview.bfsmlt.top/frontend

# 安装依赖
npm ci --production=false

# 构建生产版本 (使用生产环境配置)
npm run build

# 验证构建结果
ls -la dist/
```

### 3. 关键文件检查
```bash
# 检查关键文件是否存在
ls -la /www/wwwroot/jobview.bfsmlt.top/
echo "前端构建文件:"
ls -la /www/wwwroot/jobview.bfsmlt.top/frontend/dist/
echo "检查index.html是否存在:"
cat /www/wwwroot/jobview.bfsmlt.top/frontend/dist/index.html | head -10
```

## ⚡ 后端部署

### 1. 后端编译
```bash
cd /www/wwwroot/jobview.bfsmlt.top/backend

# 下载Go依赖
go mod download
go mod verify

# 编译生产版本
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o jobview-server cmd/main.go

# 验证编译结果
ls -la jobview-server
file jobview-server
```

### 2. 后端服务配置
```bash
# 创建PM2配置文件
cat > /www/wwwroot/jobview.bfsmlt.top/ecosystem.config.js << 'EOF'
module.exports = {
  apps: [{
    name: 'jobview-backend',
    script: './backend/jobview-server',
    cwd: '/www/wwwroot/jobview.bfsmlt.top',
    instances: 1,
    autorestart: true,
    watch: false,
    max_memory_restart: '500M',
    env: {
      NODE_ENV: 'production'
    },
    env_file: '/www/wwwroot/jobview.bfsmlt.top/.env.production',
    log_file: '/www/wwwroot/jobview.bfsmlt.top/logs/backend.log',
    error_file: '/www/wwwroot/jobview.bfsmlt.top/logs/backend-error.log',
    out_file: '/www/wwwroot/jobview.bfsmlt.top/logs/backend-out.log',
    log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
  }]
};
EOF

# 创建日志目录
mkdir -p /www/wwwroot/jobview.bfsmlt.top/logs
mkdir -p /www/wwwroot/jobview.bfsmlt.top/uploads
chown -R www:www /www/wwwroot/jobview.bfsmlt.top/logs
chown -R www:www /www/wwwroot/jobview.bfsmlt.top/uploads
```

### 3. 启动后端服务
```bash
cd /www/wwwroot/jobview.bfsmlt.top

# 使用PM2启动
pm2 start ecosystem.config.js

# 保存PM2配置
pm2 save

# 设置开机自启
pm2 startup
# 按照提示执行输出的命令

# 检查服务状态
pm2 status
pm2 logs jobview-backend
```

## 🌐 Nginx配置

### 您的Nginx配置分析和优化
您提供的配置基本正确，但需要一些调整：

```nginx
server {
    listen 80;
    listen 443 ssl http2;
    server_name jobview.bfsmlt.top;

    # 关键修改：指向前端构建目录
    root /www/wwwroot/jobview.bfsmlt.top/frontend/dist;
    index index.html index.htm;

    # SSL配置 (保持不变)
    if ($server_port !~ 443){
        rewrite ^(/.*)$ https://$host$1 permanent;
    }
    ssl_certificate    /www/server/panel/vhost/cert/jobview.bfsmlt.top/fullchain.pem;
    ssl_certificate_key  /www/server/panel/vhost/cert/jobview.bfsmlt.top/privkey.pem;
    ssl_protocols TLSv1.1 TLSv1.2 TLSv1.3;
    ssl_ciphers EECDH+CHACHA20:EECDH+CHACHA20-draft:EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:EECDH+3DES:RSA+3DES:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    add_header Strict-Transport-Security "max-age=31536000";
    error_page 497  https://$host$request_uri;

    # 前端路由支持 (保持不变，很好)
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API反向代理 (保持不变，正确)
    location /api/ {
        proxy_pass http://127.0.0.1:8010;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 新增：超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # 新增：缓冲配置
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # 新增：文件上传支持
    location /api/upload {
        proxy_pass http://127.0.0.1:8010;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 文件上传配置
        client_max_body_size 10M;
        proxy_request_buffering off;
    }

    # 新增：静态资源优化
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header Vary Accept-Encoding;

        # 启用gzip压缩
        gzip on;
        gzip_vary on;
        gzip_min_length 1024;
        gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript image/svg+xml;
    }

    # 安全配置 (保持不变)
    location ~ ^/(\.user.ini|\.htaccess|\.git|\.env|\.svn|\.project|LICENSE|README.md) {
        return 404;
    }

    # SSL验证 (保持不变)
    location ~ \.well-known {
        allow all;
    }

    # 日志配置 (保持不变)
    access_log  /www/wwwlogs/jobview.bfsmlt.top.log;
    error_log  /www/wwwlogs/jobview.bfsmlt.top.error.log;
}
```

### 应用Nginx配置
```bash
# 1. 备份原配置
cp /www/server/panel/vhost/nginx/jobview.bfsmlt.top.conf /www/server/panel/vhost/nginx/jobview.bfsmlt.top.conf.backup

# 2. 更新配置 (在宝塔面板中修改，或者直接编辑文件)
# 主要修改：root 路径指向 /www/wwwroot/jobview.bfsmlt.top/frontend/dist

# 3. 测试配置
nginx -t

# 4. 重载配置
nginx -s reload
# 或者
systemctl reload nginx
```

## 🔍 部署验证

### 1. 数据库连接测试
```bash
# 测试数据库连接
export PGPASSWORD='your_secure_password_here'
psql -h localhost -U jobview_user -d jobview_prod -c "SELECT version();"
psql -h localhost -U jobview_user -d jobview_prod -c "SELECT count(*) FROM users;"
```

### 2. 后端服务测试
```bash
# 测试后端服务
curl -I http://127.0.0.1:8010/api/v1/health
curl http://127.0.0.1:8010/api/v1/health

# 检查服务状态
pm2 status
pm2 logs jobview-backend --lines 50
```

### 3. 前端访问测试
```bash
# 测试静态文件
curl -I https://jobview.bfsmlt.top/
curl -I https://jobview.bfsmlt.top/static/js/
```

### 4. 完整功能测试
```bash
# 测试API接口
curl -X POST https://jobview.bfsmlt.top/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"TestPass123!"}'

# 测试前端路由
curl -I https://jobview.bfsmlt.top/dashboard
curl -I https://jobview.bfsmlt.top/kanban
```

### 5. 性能检查
```bash
# 检查响应时间
curl -o /dev/null -s -w '%{time_total}\n' https://jobview.bfsmlt.top/api/v1/health

# 检查系统资源
top -p $(pgrep jobview-server)
free -h
df -h
```

## 🔄 更新流程

### 1. 前端更新
```bash
cd /www/wwwroot/jobview.bfsmlt.top/frontend

# 拉取最新代码
git pull origin main

# 重新构建
npm ci
npm run build

# 无需重启Nginx (静态文件自动更新)
```

### 2. 后端更新
```bash
cd /www/wwwroot/jobview.bfsmlt.top

# 拉取最新代码
git pull origin main

# 重新编译
cd backend
go build -o jobview-server cmd/main.go

# 重启服务
pm2 restart jobview-backend

# 检查状态
pm2 status
pm2 logs jobview-backend --lines 20
```

### 3. 数据库更新 (如有需要)
```bash
# 备份数据库
pg_dump -h localhost -U jobview_user jobview_prod > /backup/jobview_$(date +%Y%m%d_%H%M%S).sql

# 执行新的迁移脚本
psql -h localhost -U jobview_user -d jobview_prod -f new_migration.sql
```

### 4. 完整更新脚本
```bash
#!/bin/bash
# update-jobview.sh

set -e

echo "开始更新JobView..."

# 备份
echo "1. 创建备份..."
pg_dump -h localhost -U jobview_user jobview_prod > /backup/jobview_$(date +%Y%m%d_%H%M%S).sql

# 更新代码
echo "2. 更新代码..."
cd /www/wwwroot/jobview.bfsmlt.top
git pull origin main

# 更新前端
echo "3. 更新前端..."
cd frontend
npm ci
npm run build

# 更新后端
echo "4. 更新后端..."
cd ../backend
go build -o jobview-server cmd/main.go

# 重启服务
echo "5. 重启服务..."
pm2 restart jobview-backend

# 验证
echo "6. 验证更新..."
sleep 5
curl -f http://127.0.0.1:8010/api/v1/health || (echo "后端服务异常!" && exit 1)
curl -f https://jobview.bfsmlt.top/ || (echo "前端访问异常!" && exit 1)

echo "更新完成!"
pm2 status
```

## 🚨 故障排除

### 常见问题和解决方案

#### 1. 前端404错误
```bash
# 检查文件权限
ls -la /www/wwwroot/jobview.bfsmlt.top/frontend/dist/
chown -R www:www /www/wwwroot/jobview.bfsmlt.top/

# 检查Nginx配置
nginx -t
```

#### 2. API请求失败
```bash
# 检查后端服务
pm2 logs jobview-backend
netstat -tlnp | grep :8010

# 检查防火墙
firewall-cmd --list-ports
```

#### 3. 数据库连接失败
```bash
# 检查PostgreSQL状态
systemctl status postgresql
netstat -tlnp | grep :5432

# 测试连接
psql -h localhost -U jobview_user -d jobview_prod
```

#### 4. 路径配置问题
```bash
# 检查环境变量
cat /www/wwwroot/jobview.bfsmlt.top/.env.production

# 检查前端API配置
cat /www/wwwroot/jobview.bfsmlt.top/frontend/.env.production
```

---

**🚀 按照此指南，您可以成功将JobView部署到生产环境！**

> **注意事项**:
> 1. 请将所有 `your_secure_password_here` 替换为实际的安全密码
> 2. 确保防火墙开放80和443端口
> 3. 定期备份数据库和重要文件
> 4. 监控服务状态和系统资源使用情况