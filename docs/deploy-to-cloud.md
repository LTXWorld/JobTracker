# JobView 云服务器部署指南（Production）

本文档说明如何将本项目部署到一台 Ubuntu 22.04 云服务器。提供推荐方案（Nginx + Systemd）与可选方案（Docker Compose）。

---

## 0. 准备工作

- 云服务器（公网 IP）、可选域名 your-domain.com
- 开放端口：22/80/443
- 基础工具：`sudo apt update && sudo apt install -y curl git ufw`

环境变量文件（服务器上）：`/opt/jobview/backend/.env`

```
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=jobview
DB_PASSWORD=change_me_strong
DB_NAME=jobview_db
DB_SSLMODE=disable

SERVER_PORT=8010
ENVIRONMENT=production

JWT_SECRET=replace_with_very_strong_256bit_secret
JWT_ACCESS_DURATION=24h
JWT_REFRESH_DURATION=720h
```

注意：生产需将后端 CORS 白名单改为包含你的域名（`backend/cmd/main.go` 中的 `CORSMiddleware`）。

前端建议以同源方式访问后端：将 `frontend/src/api/request.ts` 的 `baseURL` 改为 `/api`，由 Nginx 反代到后端。

---

## 1) 推荐方案：Nginx + Systemd + PostgreSQL

### 1.1 安装 Nginx 与 Certbot（HTTPS）

```
sudo apt install -y nginx
sudo snap install core; sudo snap refresh core
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
```

### 1.2 准备 PostgreSQL（任选）

- 容器方式：

```
sudo apt install -y docker.io docker-compose-plugin
sudo docker run -d --name pg-jobview \
  -e POSTGRES_DB=jobview_db \
  -e POSTGRES_USER=jobview \
  -e POSTGRES_PASSWORD=change_me_strong \
  -v /opt/jobview/pgdata:/var/lib/postgresql/data \
  -p 5432:5432 postgres:15-alpine
```

- 托管数据库：将连接信息填入 `.env`。

### 1.3 后端部署（Systemd）

```
sudo mkdir -p /opt/jobview/backend
cd /opt/jobview/backend

# 构建二进制（可在本机或服务器）
sudo apt install -y golang-go
cd /path/to/repo/backend
GOOS=linux GOARCH=amd64 go build -o jobview-backend ./cmd

# 拷贝产物与 .env 到服务器
cp jobview-backend /opt/jobview/backend/
cp /path/to/.env /opt/jobview/backend/.env
```

Systemd：`/etc/systemd/system/jobview-backend.service`

```
[Unit]
Description=JobView Backend
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/jobview/backend
EnvironmentFile=/opt/jobview/backend/.env
ExecStart=/opt/jobview/backend/jobview-backend
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

启动：

```
sudo systemctl daemon-reload
sudo systemctl enable --now jobview-backend
sudo systemctl status jobview-backend
```

### 1.4 前端构建与部署（静态）

```
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

cd /path/to/repo/frontend
sed -i "s#baseURL: 'http://localhost:8010'#baseURL: '/api'#" src/api/request.ts
npm ci && npm run build

sudo mkdir -p /var/www/jobview
sudo rsync -a dist/ /var/www/jobview/
```

### 1.5 Nginx 站点（同源反代）

`/etc/nginx/sites-available/jobview.conf`

```
server {
  listen 80;
  server_name your-domain.com;

  root /var/www/jobview;
  index index.html;

  location / { try_files $uri $uri/ /index.html; }

  location /api/ {
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_pass http://127.0.0.1:8010/api/;
  }

  location /static/ { proxy_pass http://127.0.0.1:8010/static/; }
}
```

启用与 HTTPS：

```
sudo ln -s /etc/nginx/sites-available/jobview.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
sudo certbot --nginx -d your-domain.com
```

访问：
- 前端 `https://your-domain.com`
- API `https://your-domain.com/api/...`
- 附件 `https://your-domain.com/static/...`

---

## 2) 可选：Docker Compose 一体化

示例结构：

- 构建后端镜像的 `infra/docker/Dockerfile.backend`
- Nginx 站点 `infra/docker/nginx.conf`
- 根目录 `docker-compose.yml` 将 db/backend/frontend 拉起

流程：

```
# 先构建前端静态
cd frontend && npm ci && npm run build

# 在项目根目录
docker compose up -d --build
```

如需 HTTPS，可在宿主机层以 Nginx/Caddy 做 443 终端并反代到容器 80。

---

## 3) 运维与排障

- 健康检查：`GET /health` 返回 200
- 日志：`journalctl -u jobview-backend -f`、`/var/log/nginx/*.log`
- 备份：`pg_dump` 定时写入 `/opt/jobview/backup/`（建议 cron）
- 常见问题：
  - CORS 报错：确认后端 CORS 白名单包含生产域名，或使用同源反代 `/api`
  - 刷新白屏：Nginx 确保 `try_files $uri $uri/ /index.html;`
  - 文件访问：确保 `/static/` 反代到后端

完成以上步骤，即可在云服务器上稳定对外服务。

