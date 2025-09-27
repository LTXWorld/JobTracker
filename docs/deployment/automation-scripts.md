# 🛠️ 自动化部署脚本集合

> 完整的服务器配置和部署自动化脚本

## 📋 脚本概览

- [🏗️ 服务器初始化脚本](#服务器初始化脚本)
- [🐳 Docker环境安装脚本](#docker环境安装脚本)
- [🚀 一键部署脚本](#一键部署脚本)
- [🗄️ 数据库管理脚本](#数据库管理脚本)
- [📊 监控和维护脚本](#监控和维护脚本)

## 🏗️ 服务器初始化脚本

### 1. 系统初始化脚本
```bash
#!/bin/bash
# scripts/init-server.sh
# 服务器初始化和基础环境配置

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🏗️ JobView 服务器初始化脚本${NC}"
echo "=================================="

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}请使用root用户运行此脚本${NC}"
    exit 1
fi

# 1. 更新系统
echo -e "${YELLOW}1. 更新系统...${NC}"
if command -v yum &> /dev/null; then
    yum update -y
    yum install -y curl wget git unzip
elif command -v apt &> /dev/null; then
    apt update && apt upgrade -y
    apt install -y curl wget git unzip
else
    echo -e "${RED}不支持的操作系统${NC}"
    exit 1
fi

# 2. 配置防火墙
echo -e "${YELLOW}2. 配置防火墙...${NC}"
if command -v firewall-cmd &> /dev/null; then
    systemctl start firewalld
    systemctl enable firewalld
    firewall-cmd --permanent --add-port=80/tcp
    firewall-cmd --permanent --add-port=443/tcp
    firewall-cmd --permanent --add-port=22/tcp
    firewall-cmd --reload
elif command -v ufw &> /dev/null; then
    ufw allow 22
    ufw allow 80
    ufw allow 443
    ufw --force enable
fi

# 3. 创建项目目录
echo -e "${YELLOW}3. 创建项目目录...${NC}"
PROJECT_DIR="/www/wwwroot/jobview.bfsmlt.top"
mkdir -p $PROJECT_DIR
mkdir -p $PROJECT_DIR/{logs,backup,scripts,nginx/conf.d}
mkdir -p /backup/jobview

# 4. 配置时区
echo -e "${YELLOW}4. 配置时区...${NC}"
timedatectl set-timezone Asia/Shanghai

# 5. 配置SSH安全
echo -e "${YELLOW}5. 配置SSH安全...${NC}"
cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup
sed -i 's/#PermitRootLogin yes/PermitRootLogin yes/' /etc/ssh/sshd_config
sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config
systemctl restart sshd

# 6. 安装必要工具
echo -e "${YELLOW}6. 安装必要工具...${NC}"
if command -v yum &> /dev/null; then
    yum install -y htop tree vim net-tools
elif command -v apt &> /dev/null; then
    apt install -y htop tree vim net-tools
fi

echo -e "${GREEN}✅ 服务器初始化完成！${NC}"
echo "项目目录: $PROJECT_DIR"
echo "备份目录: /backup/jobview"
```

## 🐳 Docker环境安装脚本

### 2. Docker安装脚本
```bash
#!/bin/bash
# scripts/install-docker.sh
# Docker和Docker Compose安装脚本

set -e

echo -e "${BLUE}🐳 Docker环境安装脚本${NC}"
echo "=================================="

# 1. 卸载旧版本Docker
echo -e "${YELLOW}1. 清理旧版本Docker...${NC}"
if command -v yum &> /dev/null; then
    yum remove -y docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-engine
elif command -v apt &> /dev/null; then
    apt remove -y docker docker-engine docker.io containerd runc
fi

# 2. 安装Docker
echo -e "${YELLOW}2. 安装Docker...${NC}"
curl -fsSL https://get.docker.com | bash

# 3. 启动Docker服务
echo -e "${YELLOW}3. 启动Docker服务...${NC}"
systemctl start docker
systemctl enable docker

# 4. 配置Docker用户组
echo -e "${YELLOW}4. 配置Docker用户组...${NC}"
groupadd docker 2>/dev/null || true
usermod -aG docker root

# 5. 配置Docker镜像加速器 (可选，中国大陆使用)
echo -e "${YELLOW}5. 配置Docker镜像加速器...${NC}"
mkdir -p /etc/docker
cat > /etc/docker/daemon.json << 'EOF'
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  },
  "storage-driver": "overlay2"
}
EOF

# 6. 重启Docker服务
systemctl restart docker

# 7. 安装Docker Compose
echo -e "${YELLOW}6. 安装Docker Compose...${NC}"
DOCKER_COMPOSE_VERSION="v2.24.0"
curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose

# 8. 验证安装
echo -e "${YELLOW}7. 验证安装...${NC}"
docker --version
docker-compose --version

# 9. 测试Docker
echo -e "${YELLOW}8. 测试Docker...${NC}"
docker run --rm hello-world

echo -e "${GREEN}✅ Docker环境安装完成！${NC}"
echo "Docker版本: $(docker --version)"
echo "Docker Compose版本: $(docker-compose --version)"
```

## 🚀 一键部署脚本

### 3. 一键部署主脚本
```bash
#!/bin/bash
# scripts/deploy.sh
# JobView一键部署脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 默认配置
DOMAIN="jobview.bfsmlt.top"
PROJECT_DIR="/www/wwwroot/jobview.bfsmlt.top"
ENVIRONMENT="production"
SKIP_MIGRATION=false
SKIP_BACKUP=false

# 帮助信息
show_help() {
    cat << EOF
JobView 一键部署脚本

用法: $0 [选项]

选项:
    --domain DOMAIN        部署域名 (默认: jobview.bfsmlt.top)
    --env ENV             环境 (development|production, 默认: production)
    --db-password PWD     数据库密码
    --jwt-secret SECRET   JWT密钥
    --skip-migration      跳过数据库迁移
    --skip-backup         跳过备份
    --help               显示此帮助信息

示例:
    $0 --domain jobview.bfsmlt.top --env production
    $0 --skip-migration --skip-backup

EOF
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --domain)
            DOMAIN="$2"
            shift 2
            ;;
        --env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --db-password)
            DB_PASSWORD="$2"
            shift 2
            ;;
        --jwt-secret)
            JWT_SECRET="$2"
            shift 2
            ;;
        --skip-migration)
            SKIP_MIGRATION=true
            shift
            ;;
        --skip-backup)
            SKIP_BACKUP=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}未知选项: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

echo -e "${BLUE}🚀 JobView 一键部署脚本${NC}"
echo "=================================="
echo "域名: $DOMAIN"
echo "环境: $ENVIRONMENT"
echo "项目目录: $PROJECT_DIR"

# 1. 环境检查
echo -e "${YELLOW}1. 检查部署环境...${NC}"

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker未安装，正在安装...${NC}"
    curl -fsSL https://get.docker.com | bash
    systemctl start docker
    systemctl enable docker
fi

# 检查Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Docker Compose未安装，正在安装...${NC}"
    DOCKER_COMPOSE_VERSION="v2.24.0"
    curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

# 检查Git
if ! command -v git &> /dev/null; then
    echo -e "${RED}Git未安装，正在安装...${NC}"
    if command -v yum &> /dev/null; then
        yum install -y git
    elif command -v apt &> /dev/null; then
        apt install -y git
    fi
fi

# 2. 创建项目目录
echo -e "${YELLOW}2. 准备项目目录...${NC}"
mkdir -p $PROJECT_DIR
cd $PROJECT_DIR

# 3. 克隆或更新代码
echo -e "${YELLOW}3. 获取最新代码...${NC}"
if [ -d ".git" ]; then
    echo "更新现有代码..."
    git pull origin main
else
    echo "克隆代码仓库..."
    # 注意：这里需要替换为实际的仓库地址
    git clone https://github.com/your-username/jobView.git .
fi

# 4. 配置环境变量
echo -e "${YELLOW}4. 配置环境变量...${NC}"

# 生成随机密码和密钥（如果未提供）
if [ -z "$DB_PASSWORD" ]; then
    DB_PASSWORD=$(openssl rand -base64 32)
    echo "生成的数据库密码: $DB_PASSWORD"
fi

if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(openssl rand -base64 48)
    echo "生成的JWT密钥: $JWT_SECRET"
fi

# 创建环境变量文件
cat > .env.production << EOF
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=jobview_user
DB_PASSWORD=$DB_PASSWORD
DB_NAME=jobview_prod

# JWT配置
JWT_SECRET=$JWT_SECRET

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# 应用配置
ENVIRONMENT=$ENVIRONMENT
GIN_MODE=release
PORT=8010
HOST=0.0.0.0

# 前端配置
VITE_API_BASE=/api

# 域名配置
DOMAIN=$DOMAIN
EOF

echo -e "${GREEN}✅ 环境变量配置完成${NC}"

# 5. 备份现有数据（如果存在）
if [ "$SKIP_BACKUP" != true ] && docker ps -q -f name=jobview-postgres &> /dev/null; then
    echo -e "${YELLOW}5. 备份现有数据...${NC}"
    BACKUP_FILE="/backup/jobview/backup_$(date +%Y%m%d_%H%M%S).sql"
    mkdir -p /backup/jobview
    docker-compose exec -T postgres pg_dump -U jobview_user jobview_prod > $BACKUP_FILE
    echo -e "${GREEN}✅ 备份完成: $BACKUP_FILE${NC}"
else
    echo -e "${YELLOW}5. 跳过数据备份${NC}"
fi

# 6. 停止现有服务
echo -e "${YELLOW}6. 停止现有服务...${NC}"
docker-compose down --remove-orphans || true

# 7. 构建和启动服务
echo -e "${YELLOW}7. 构建和启动服务...${NC}"
docker-compose -f docker-compose.yml --env-file .env.production up -d --build

# 8. 等待服务启动
echo -e "${YELLOW}8. 等待服务启动...${NC}"
sleep 30

# 9. 数据库迁移
if [ "$SKIP_MIGRATION" != true ]; then
    echo -e "${YELLOW}9. 执行数据库迁移...${NC}"
    ./scripts/migrate-database.sh
else
    echo -e "${YELLOW}9. 跳过数据库迁移${NC}"
fi

# 10. 健康检查
echo -e "${YELLOW}10. 健康检查...${NC}"
sleep 10

# 检查后端健康
if curl -f http://localhost:8010/api/v1/health &> /dev/null; then
    echo -e "${GREEN}✅ 后端服务健康${NC}"
else
    echo -e "${RED}❌ 后端服务异常${NC}"
    docker-compose logs backend
    exit 1
fi

# 检查前端访问
if curl -f http://localhost/ &> /dev/null; then
    echo -e "${GREEN}✅ 前端服务正常${NC}"
else
    echo -e "${RED}❌ 前端服务异常${NC}"
    docker-compose logs frontend
    exit 1
fi

# 11. 显示部署信息
echo -e "${GREEN}🎉 部署完成！${NC}"
echo "=================================="
echo "访问地址: https://$DOMAIN"
echo "API地址: https://$DOMAIN/api/v1/health"
echo "项目目录: $PROJECT_DIR"
echo ""
echo "服务状态:"
docker-compose ps
echo ""
echo "查看日志: docker-compose logs -f"
echo "重启服务: docker-compose restart"
echo "停止服务: docker-compose down"
```

### 4. 数据库迁移脚本
```bash
#!/bin/bash
# scripts/migrate-database.sh
# 数据库迁移脚本

set -e

echo -e "${BLUE}🗄️ 数据库迁移脚本${NC}"
echo "=========================="

# 加载环境变量
if [ -f .env.production ]; then
    source .env.production
fi

# 等待数据库服务就绪
echo "等待数据库服务启动..."
until docker-compose exec postgres pg_isready -U $DB_USER -d $DB_NAME; do
    echo "等待PostgreSQL启动..."
    sleep 5
done

echo -e "${GREEN}✅ 数据库连接成功${NC}"

# 执行迁移SQL
echo "执行数据库结构迁移..."
docker-compose exec -T postgres psql -U $DB_USER -d $DB_NAME << 'EOF'
-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 用户表
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

-- 投递记录表
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
    reminder_enabled BOOLEAN DEFAULT FALSE,
    reminder_time TIMESTAMP,
    follow_up_date DATE,
    last_status_change TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status_version INTEGER DEFAULT 1,
    status_history JSONB DEFAULT '{}',
    status_duration_stats JSONB DEFAULT '{}',
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

-- 状态历史表
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

-- 用户偏好表
CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference_config JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_job_applications_user_id ON job_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_user_date ON job_applications(user_id, application_date DESC);
CREATE INDEX IF NOT EXISTS idx_job_applications_user_status ON job_applications(user_id, status);
CREATE INDEX IF NOT EXISTS idx_job_applications_search ON job_applications USING gin(to_tsvector('simple', company_name || ' ' || position_title || ' ' || COALESCE(work_location, '')));

-- 创建默认用户
INSERT INTO users (username, email, password_hash, full_name) VALUES
('testuser', 'test@jobview.com', '$2a$10$k8Y1THPD8eRKQVbFdoYFRu.9xhqHdF7YNJ3/dFjDxQSQx4lCpZmKO', '测试用户')
ON CONFLICT (username) DO NOTHING;

\echo '✅ 数据库迁移完成'
EOF

echo -e "${GREEN}✅ 数据库迁移完成${NC}"
```

### 5. 服务管理脚本
```bash
#!/bin/bash
# scripts/manage-service.sh
# 服务管理脚本

set -e

PROJECT_DIR="/www/wwwroot/jobview.bfsmlt.top"

show_help() {
    cat << EOF
JobView 服务管理脚本

用法: $0 <命令> [选项]

命令:
    start           启动所有服务
    stop            停止所有服务
    restart         重启所有服务
    status          查看服务状态
    logs [服务名]    查看日志
    backup          备份数据库
    update          更新服务（拉取最新代码并重启）
    clean           清理未使用的Docker资源

示例:
    $0 start
    $0 logs backend
    $0 backup
    $0 update

EOF
}

# 切换到项目目录
cd $PROJECT_DIR

case "$1" in
    start)
        echo "🚀 启动JobView服务..."
        docker-compose up -d
        echo "✅ 服务启动完成"
        docker-compose ps
        ;;
    stop)
        echo "⏹️ 停止JobView服务..."
        docker-compose down
        echo "✅ 服务停止完成"
        ;;
    restart)
        echo "🔄 重启JobView服务..."
        docker-compose restart
        echo "✅ 服务重启完成"
        docker-compose ps
        ;;
    status)
        echo "📊 JobView服务状态:"
        docker-compose ps
        echo ""
        echo "系统资源使用:"
        docker stats --no-stream
        ;;
    logs)
        if [ -n "$2" ]; then
            echo "📋 查看${2}服务日志:"
            docker-compose logs -f --tail=100 $2
        else
            echo "📋 查看所有服务日志:"
            docker-compose logs -f --tail=50
        fi
        ;;
    backup)
        echo "💾 备份数据库..."
        BACKUP_FILE="/backup/jobview/manual_backup_$(date +%Y%m%d_%H%M%S).sql"
        mkdir -p /backup/jobview
        docker-compose exec -T postgres pg_dump -U jobview_user jobview_prod > $BACKUP_FILE
        echo "✅ 备份完成: $BACKUP_FILE"
        ;;
    update)
        echo "🔄 更新JobView..."
        # 备份数据
        echo "1. 备份数据..."
        BACKUP_FILE="/backup/jobview/update_backup_$(date +%Y%m%d_%H%M%S).sql"
        mkdir -p /backup/jobview
        docker-compose exec -T postgres pg_dump -U jobview_user jobview_prod > $BACKUP_FILE

        # 拉取最新代码
        echo "2. 拉取最新代码..."
        git pull origin main

        # 重新构建并重启
        echo "3. 重新构建并重启服务..."
        docker-compose down
        docker-compose up -d --build

        # 健康检查
        echo "4. 健康检查..."
        sleep 30
        if curl -f http://localhost:8010/api/v1/health &> /dev/null; then
            echo "✅ 更新完成，服务正常"
        else
            echo "❌ 更新失败，请检查日志"
            docker-compose logs
        fi
        ;;
    clean)
        echo "🧹 清理Docker资源..."
        docker system prune -f
        docker volume prune -f
        echo "✅ 清理完成"
        ;;
    *)
        show_help
        exit 1
        ;;
esac
```

## 📊 监控和维护脚本

### 6. 监控脚本
```bash
#!/bin/bash
# scripts/monitor.sh
# 系统监控脚本

set -e

PROJECT_DIR="/www/wwwroot/jobview.bfsmlt.top"
LOG_FILE="/var/log/jobview-monitor.log"

# 记录日志
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a $LOG_FILE
}

# 检查服务健康状况
check_health() {
    cd $PROJECT_DIR

    log "开始健康检查..."

    # 检查Docker服务
    if ! systemctl is-active --quiet docker; then
        log "❌ Docker服务未运行"
        systemctl start docker
        sleep 10
    fi

    # 检查容器状态
    CONTAINERS=(postgres redis backend frontend nginx)
    for container in "${CONTAINERS[@]}"; do
        if ! docker-compose ps | grep "jobview-$container" | grep -q "Up"; then
            log "❌ 容器 $container 未运行，尝试重启..."
            docker-compose restart $container
            sleep 10
        else
            log "✅ 容器 $container 运行正常"
        fi
    done

    # 检查API健康
    if curl -f http://localhost:8010/api/v1/health &> /dev/null; then
        log "✅ API服务健康"
    else
        log "❌ API服务异常，重启后端服务..."
        docker-compose restart backend
    fi

    # 检查前端访问
    if curl -f http://localhost/ &> /dev/null; then
        log "✅ 前端服务正常"
    else
        log "❌ 前端服务异常，重启前端服务..."
        docker-compose restart frontend nginx
    fi

    # 检查磁盘空间
    DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ $DISK_USAGE -gt 80 ]; then
        log "⚠️ 磁盘使用率过高: ${DISK_USAGE}%"
        # 清理Docker资源
        docker system prune -f
        # 清理旧备份（保留最近7天）
        find /backup/jobview -name "*.sql" -mtime +7 -delete
    fi

    # 检查内存使用
    MEMORY_USAGE=$(free | awk 'NR==2{printf "%.2f", $3*100/$2}')
    if (( $(echo "$MEMORY_USAGE > 80" | bc -l) )); then
        log "⚠️ 内存使用率过高: ${MEMORY_USAGE}%"
    fi

    log "健康检查完成"
}

# 自动备份
auto_backup() {
    cd $PROJECT_DIR

    log "开始自动备份..."

    BACKUP_DIR="/backup/jobview"
    mkdir -p $BACKUP_DIR

    # 创建备份
    BACKUP_FILE="$BACKUP_DIR/auto_backup_$(date +%Y%m%d_%H%M%S).sql"
    if docker-compose exec -T postgres pg_dump -U jobview_user jobview_prod > $BACKUP_FILE; then
        log "✅ 自动备份完成: $BACKUP_FILE"

        # 压缩备份文件
        gzip $BACKUP_FILE

        # 删除超过30天的备份
        find $BACKUP_DIR -name "auto_backup_*.sql.gz" -mtime +30 -delete

    else
        log "❌ 自动备份失败"
    fi
}

# 性能报告
performance_report() {
    cd $PROJECT_DIR

    log "生成性能报告..."

    # 系统资源
    log "=== 系统资源使用 ==="
    log "CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}')"
    log "内存: $(free | grep Mem | awk '{printf "使用: %.1f%% (%s/%s)", $3/$2 * 100.0, $3, $2}')"
    log "磁盘: $(df -h / | awk 'NR==2 {print $5 " (" $3 "/" $2 ")"}')"

    # Docker容器状态
    log "=== Docker容器状态 ==="
    docker-compose ps 2>&1 | while read line; do log "$line"; done

    # API响应时间
    RESPONSE_TIME=$(curl -o /dev/null -s -w '%{time_total}\n' http://localhost:8010/api/v1/health)
    log "API响应时间: ${RESPONSE_TIME}秒"
}

# 主函数
main() {
    case "$1" in
        health)
            check_health
            ;;
        backup)
            auto_backup
            ;;
        report)
            performance_report
            ;;
        all)
            check_health
            auto_backup
            performance_report
            ;;
        *)
            echo "用法: $0 {health|backup|report|all}"
            echo "  health  - 健康检查"
            echo "  backup  - 自动备份"
            echo "  report  - 性能报告"
            echo "  all     - 执行所有检查"
            exit 1
            ;;
    esac
}

main "$@"
```

### 7. 定时任务配置
```bash
#!/bin/bash
# scripts/setup-cron.sh
# 配置定时任务

echo "配置JobView定时任务..."

# 创建cron任务文件
cat > /etc/cron.d/jobview << 'EOF'
# JobView 定时任务
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

# 每15分钟进行健康检查
*/15 * * * * root /www/wwwroot/jobview.bfsmlt.top/scripts/monitor.sh health

# 每天凌晨2点自动备份数据库
0 2 * * * root /www/wwwroot/jobview.bfsmlt.top/scripts/monitor.sh backup

# 每周日凌晨3点生成性能报告
0 3 * * 0 root /www/wwwroot/jobview.bfsmlt.top/scripts/monitor.sh report

# 每天凌晨4点清理Docker资源
0 4 * * * root docker system prune -f
EOF

# 重启cron服务
systemctl restart crond 2>/dev/null || systemctl restart cron 2>/dev/null

echo "✅ 定时任务配置完成"
echo "查看任务: crontab -l"
echo "查看日志: tail -f /var/log/jobview-monitor.log"
```

---

**🛠️ 自动化部署脚本集合创建完成！现在更新主要的部署文档。**