# 🐳 Docker容器化自动部署指南

> 基于Docker + GitHub Actions的完整CI/CD解决方案

## 📋 目录

- [🎯 架构概览](#架构概览)
- [🐳 Docker配置](#docker配置)
- [⚙️ GitHub Actions CI/CD](#github-actions-cicd)
- [🚀 部署流程](#部署流程)
- [🔧 服务器配置](#服务器配置)
- [📊 监控和维护](#监控和维护)

## 🎯 架构概览

### 🏗️ 容器化架构
```
GitHub Repository (main分支)
    ↓ (Push触发)
GitHub Actions
    ├── 构建前端镜像
    ├── 构建后端镜像
    └── 推送到服务器
        ↓
服务器 Docker Environment
├── nginx-proxy (反向代理)
├── jobview-frontend (Vue 3)
├── jobview-backend (Go)
├── postgres (数据库)
└── redis (缓存,可选)
```

### 🔄 自动化流程
1. **代码推送** → GitHub main分支
2. **GitHub Actions** → 自动构建Docker镜像
3. **部署到服务器** → Docker Compose更新服务
4. **健康检查** → 验证部署成功
5. **通知** → 部署结果通知

## 🐳 Docker配置

### 1. 前端Dockerfile
```dockerfile
# frontend/Dockerfile
FROM node:18-alpine AS builder

WORKDIR /app

# 复制package文件
COPY package*.json ./
RUN npm ci --only=production

# 复制源代码
COPY . .

# 构建生产版本
RUN npm run build

# 生产环境
FROM nginx:alpine

# 复制构建结果
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制Nginx配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

### 2. 前端Nginx配置
```nginx
# frontend/nginx.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # 前端路由支持 (SPA)
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types
        text/plain
        text/css
        application/json
        application/javascript
        text/xml
        application/xml
        application/xml+rss
        text/javascript
        image/svg+xml;
}
```

### 3. 后端Dockerfile
```dockerfile
# backend/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装必要工具
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o jobview-server cmd/main.go

# 生产环境
FROM alpine:latest

# 安装CA证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 复制二进制文件
COPY --from=builder /app/jobview-server .

# 创建必要目录
RUN mkdir -p logs uploads

# 暴露端口
EXPOSE 8010

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8010/api/v1/health || exit 1

CMD ["./jobview-server"]
```

### 4. Docker Compose配置
```yaml
# docker-compose.yml
version: '3.8'

services:
  # PostgreSQL数据库
  postgres:
    image: postgres:15-alpine
    container_name: jobview-postgres
    environment:
      POSTGRES_DB: ${DB_NAME:-jobview_prod}
      POSTGRES_USER: ${DB_USER:-jobview_user}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    networks:
      - jobview-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-jobview_user} -d ${DB_NAME:-jobview_prod}"]
      interval: 30s
      timeout: 10s
      retries: 5

  # Redis缓存 (可选)
  redis:
    image: redis:7-alpine
    container_name: jobview-redis
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    networks:
      - jobview-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 5

  # 后端服务
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: jobview-backend
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER:-jobview_user}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME:-jobview_prod}
      - JWT_SECRET=${JWT_SECRET}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - ENVIRONMENT=production
      - GIN_MODE=release
      - PORT=8010
      - HOST=0.0.0.0
    volumes:
      - backend_logs:/root/logs
      - backend_uploads:/root/uploads
    networks:
      - jobview-network
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8010/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 5

  # 前端服务
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: jobview-frontend
    networks:
      - jobview-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost/"]
      interval: 30s
      timeout: 10s
      retries: 5

  # Nginx反向代理
  nginx:
    image: nginx:alpine
    container_name: jobview-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/conf.d:/etc/nginx/conf.d
      - /www/server/panel/vhost/cert/jobview.bfsmlt.top:/etc/nginx/ssl
    networks:
      - jobview-network
    depends_on:
      - frontend
      - backend
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 5

volumes:
  postgres_data:
  redis_data:
  backend_logs:
  backend_uploads:

networks:
  jobview-network:
    driver: bridge
```

### 5. 环境变量配置
```bash
# .env.production
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=jobview_user
DB_PASSWORD=your_secure_database_password_here
DB_NAME=jobview_prod

# JWT配置
JWT_SECRET=your-super-secret-jwt-key-must-be-at-least-32-characters-long

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# 应用配置
ENVIRONMENT=production
GIN_MODE=release

# 前端配置
VITE_API_BASE=/api
```

### 6. Nginx主配置
```nginx
# nginx/nginx.conf
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # 日志格式
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;

    # 基础配置
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 10M;

    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        application/json
        application/javascript
        text/xml
        application/xml
        application/xml+rss
        text/javascript
        image/svg+xml;

    # 包含虚拟主机配置
    include /etc/nginx/conf.d/*.conf;
}
```

### 7. 虚拟主机配置
```nginx
# nginx/conf.d/jobview.conf
server {
    listen 80;
    server_name jobview.bfsmlt.top;

    # 重定向到HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name jobview.bfsmlt.top;

    # SSL配置
    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # 安全头
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    # API代理到后端
    location /api/ {
        proxy_pass http://backend:8010;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # 缓冲配置
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # 文件上传
    location /api/upload {
        proxy_pass http://backend:8010;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        client_max_body_size 10M;
        proxy_request_buffering off;
    }

    # 前端应用
    location / {
        proxy_pass http://frontend:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 健康检查端点
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
```

### 8. 数据库初始化脚本
```sql
-- scripts/init-db.sql
-- 数据库初始化和优化配置

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 优化PostgreSQL配置
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

-- 重新加载配置
SELECT pg_reload_conf();

-- 输出初始化信息
\echo 'JobView数据库初始化完成'
```

---

**🐳 Docker容器化配置完成！接下来创建GitHub Actions CI/CD流程。**