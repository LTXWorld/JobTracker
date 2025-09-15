# JobView 生产环境重置与重新部署指南

本指南面向“线上尚无真实用户数据”的场景，提供一套可直接执行的“推倒重来”流程：停止服务 → 清空数据库/静态资源 → 重新部署后端与前端 → 验证上线。CI/CD 由 `.github/workflows/deploy.yml` 负责构建与分发。

> 重要说明
>
> - 请在维护窗口执行，操作前确认线上确无需要保留的数据。
> - 以下命令均以 Linux 服务器、systemd 管理后端服务、PostgreSQL 运行在 Docker 容器中为前提。
> - 请替换文中 `<变量>` 占位符为你的实际值。

---

## 一、前置条件与术语

- 服务器已安装：Docker、Docker Compose（可选）、systemd、Nginx。
- 后端 systemd 服务名：`jobview-backend`
- 前端目录（Nginx 根）：`/www/wwwroot/jobview.bfsmlt.top`
- API 反向代理：`/api/ -> http://127.0.0.1:8010`
- PostgreSQL 以容器运行：
  - 若参考仓库 docker-compose.yml：容器名 `jobView`、宿主端口 5433、数据库名 `jobView_db`、用户 `ltx`
  - 若按 docs/deploy-to-cloud.md 部署：容器名 `pg-jobview`、宿主端口 5432、数据库名 `jobview_db`、用户 `jobview`、数据目录 `/opt/jobview/pgdata`
- GitHub Actions 已配置 secrets（详见第六节）。

---

## 二、停止后端服务

```bash
sudo systemctl stop jobview-backend
sudo systemctl status jobview-backend --no-pager -l || true
```

---

## 三、清空数据库（Docker 容器 PostgreSQL）

说明：使用“丢弃并重建 public schema”的方式，安全、快速，无需删除卷文件。

1) 找到容器名或 ID（示例容器名为 `jobView`）
```bash
docker ps | grep -i postgres || docker ps | grep -i jobview
```

2) 进入容器内 psql（替换数据库用户/库名）
```bash
# 方案 A：docker-compose 基线（仓库默认）
docker exec -it jobView psql -U ltx -d jobView_db

# 方案 B：deploy-to-cloud 基线
docker exec -it pg-jobview psql -U jobview -d jobview_db
```

3) 在 psql 中执行（清空所有表、函数、触发器等）
```sql
BEGIN;
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO ltx;
GRANT ALL ON SCHEMA public TO public;
COMMIT;
```

4) 退出 psql：`\q`

可选（彻底重建数据库）：
```bash
# 方案 A：docker-compose 基线（仓库默认）
docker exec -it jobView psql -U ltx -d postgres -c "DROP DATABASE jobView_db WITH (FORCE);"
docker exec -it jobView psql -U ltx -d postgres -c "CREATE DATABASE jobView_db ENCODING 'UTF8';"

# 方案 B：deploy-to-cloud 基线
docker exec -it pg-jobview psql -U jobview -d postgres -c "DROP DATABASE jobview_db WITH (FORCE);"
docker exec -it pg-jobview psql -U jobview -d postgres -c "CREATE DATABASE jobview_db ENCODING 'UTF8';"
```

---

### 三-b、完全删除并重建 PostgreSQL 容器（替代方案）

如果你希望“容器 + 数据卷”一起清空后重新启动一个全新的数据库实例，按此节操作（高危，数据不可恢复）。以下给出两种基线方案。

1) 记录数据挂载信息（在删除容器前执行）
```bash
# 方案 A：docker-compose 基线
docker inspect jobView --format '{{ json .Mounts }}' | jq '.'

# 方案 B：deploy-to-cloud 基线
docker inspect pg-jobview --format '{{ json .Mounts }}' | jq '.'

# 记下类型与宿主机路径：
# - 绑定挂载："Source": "/data/postgres" 或 "/opt/jobview/pgdata"
# - 命名卷：   "Name":   "jobview_pgdata"（示例）
```

2) 停止并删除容器
```bash
# 方案 A：docker-compose 基线
docker stop jobView || true && docker rm jobView || true

# 方案 B：deploy-to-cloud 基线
docker stop pg-jobview || true && docker rm pg-jobview || true
```

3) 删除持久化数据
```bash
# 若为绑定挂载（推荐），删除宿主机目录（危险操作，确保路径正确）
# 仓库 compose 示例路径
sudo rm -rf /data/postgres
# deploy-to-cloud 示例路径
sudo rm -rf /opt/jobview/pgdata

# 若为命名卷，删除卷
# 先列出卷
docker volume ls | grep -i pg
# 删除相应卷（示例）
docker volume rm jobview_pgdata || true
```

4) 以全新数据目录重建容器（绑定挂载示例）
```bash
# 方案 A：docker-compose 基线（与仓库配置一致）
sudo mkdir -p /data/postgres && sudo chown -R $USER:$USER /data/postgres
docker run -d \
  --name jobView \
  --restart unless-stopped \
  -e POSTGRES_DB=jobView_db \
  -e POSTGRES_USER=ltx \
  -e POSTGRES_PASSWORD='<强密码>' \
  -e PGDATA=/var/lib/postgresql/data/pgdata \
  -p 5433:5432 \
  -v /data/postgres:/var/lib/postgresql/data \
  postgres:15-alpine

# 方案 B：deploy-to-cloud 基线（参考 docs/deploy-to-cloud.md）
sudo mkdir -p /opt/jobview/pgdata && sudo chown -R $USER:$USER /opt/jobview/pgdata
docker run -d \
  --name pg-jobview \
  --restart unless-stopped \
  -e POSTGRES_DB=jobview_db \
  -e POSTGRES_USER=jobview \
  -e POSTGRES_PASSWORD='<强密码>' \
  -v /opt/jobview/pgdata:/var/lib/postgresql/data \
  -p 5432:5432 postgres:15-alpine

# 等待容器就绪（可观察日志）
docker logs -f jobView | sed -n '/database system is ready to accept connections/,$p' || true
docker logs -f pg-jobview | sed -n '/database system is ready to accept connections/,$p' || true
```

5) 连接测试
```bash
docker exec -it jobView psql -U ltx -d jobView_db -c "SELECT version();" || true
docker exec -it pg-jobview psql -U jobview -d jobview_db -c "SELECT version();" || true
```

6) 后端连接参数核对
```bash
# /etc/jobview-backend.env 中 DB_* 参数需与新容器一致

# 方案 A：docker-compose 基线（宿主机端口 5433）
DB_HOST=127.0.0.1
DB_PORT=5433
DB_USER=ltx
DB_PASSWORD=<强密码>
DB_NAME=jobView_db

# 方案 B：deploy-to-cloud 基线（宿主机端口 5432）
# DB_HOST=127.0.0.1
# DB_PORT=5432
# DB_USER=jobview
# DB_PASSWORD=<强密码>
# DB_NAME=jobview_db
```

> 说明：若你更倾向使用 docker-compose 管理，也可在 compose 目录执行：
> - `docker compose down -v` 彻底移除容器和卷
> - 调整 `docker-compose.yml` 的挂载与端口
> - `docker compose up -d` 重新创建并启动

### 三-c、基于仓库 docker-compose.yml 的全量重建（推荐参考本仓库配置）

本仓库根目录提供了 `docker-compose.yml` 示例（镜像版本、环境变量、健康检查已配置）。线上若采用 Docker 运行 PostgreSQL，建议以此为基线重建，注意把“本地开发路径”改为服务器目录。

1) 将仓库的 compose 文件拷贝到服务器（若已在服务器仓库目录可跳过）
```bash
scp docker-compose.yml <user>@<host>:/opt/jobview/
ssh <user>@<host> 'cd /opt/jobview && ls -l docker-compose.yml'
```

2) 打开并修正与服务器匹配的项（默认文件片段如下）
```yaml
services:
  postgres:
    image: postgres:15-alpine
    container_name: jobView
    restart: unless-stopped
    environment:
      POSTGRES_DB: jobView_db
      POSTGRES_USER: ltx
      POSTGRES_PASSWORD: iutaol123
      PGDATA: /var/lib/postgresql/data/pgdata
    ports:
      - "5433:5432"   # 左侧为宿主机端口
    volumes:
      - /Users/lutao/docker/postgres-data:/var/lib/postgresql/data
```

将 `volumes` 左侧的宿主目录改为服务器实际目录（例如 `/data/postgres`）：
```bash
sudo mkdir -p /data/postgres && sudo chown -R $USER:$USER /data/postgres
```

3) 以 compose 方式“完全重建”数据库容器与卷（危险操作）
```bash
cd /opt/jobview
# 停止并删除容器与卷
docker compose down -v

# 如使用绑定挂载，清空宿主机目录（危险操作，确认路径无误）
sudo rm -rf /data/postgres/*

# 重新启动容器
docker compose up -d

# 等待健康检查通过（compose 中已配置 healthcheck）
docker compose ps
docker logs -f jobView | sed -n '/database system is ready to accept connections/,$p'
```

4) 后端连接参数与 compose 对齐
- 若后端通过“宿主机端口”连接：`DB_HOST=127.0.0.1`，`DB_PORT=5433`（对应 compose 的 `5433:5432`）
- 若后端也在 Docker 网络内与 DB 通信：`DB_HOST=jobView`，`DB_PORT=5432`

参考本仓库 `backend/.env`（开发环境示例为 5433），生产建议写入 `/etc/jobview-backend.env`：
```bash
DB_HOST=127.0.0.1
DB_PORT=5433   # 与 docker-compose.yml 中宿主机端口一致
DB_USER=ltx
DB_PASSWORD=iutaol123   # 生产务必修改为强密码
DB_NAME=jobView_db
```

5) 连接测试
```bash
docker exec -it jobView psql -U ltx -d jobView_db -c "SELECT version();"
```

> 与“第三节 a/b”相比，使用 compose 的优势是：配置集中、健康检查一致、端口/卷策略明确，且与仓库示例保持对齐，后续维护成本更低。

## 四、清理线上静态资源与后端产物

1) 前端静态资源（谨慎删除，保留证书校验目录 `.well-known`）
```bash
sudo mkdir -p /www/wwwroot/jobview.bfsmlt.top
# 备份（可选）
sudo tar -C /www/wwwroot -czf /root/jobview-frontend-$(date +%F-%H%M).tgz jobview.bfsmlt.top || true
# 清空（保留 .well-known）
sudo find /www/wwwroot/jobview.bfsmlt.top -mindepth 1 -maxdepth 1 ! -name '.well-known' -exec rm -rf {} +
```

2) 后端二进制目录（以 `/opt/jobview` 为例，按你机器实际路径修改）
```bash
sudo mkdir -p /opt/jobview
# 备份（可选）
[ -f /opt/jobview/jobview-backend ] && sudo cp /opt/jobview/jobview-backend /root/jobview-backend.bak-$(date +%F-%H%M) || true
# 清理旧文件（仅在确认无用时执行）
sudo rm -f /opt/jobview/jobview-backend
```

---

## 五、后端服务与环境文件检查（首次/重置后建议核对）

1) systemd 单元文件（参考）`/etc/systemd/system/jobview-backend.service`
```ini
[Unit]
Description=JobView Backend Service
After=network.target

[Service]
Type=simple
EnvironmentFile=/etc/jobview-backend.env
WorkingDirectory=/opt/jobview
ExecStart=/opt/jobview/jobview-backend
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

2) 环境变量文件（参考）`/etc/jobview-backend.env`
```bash
# 服务器
SERVER_PORT=8010
ENVIRONMENT=production

# 数据库
DB_HOST=127.0.0.1
DB_PORT=5433        # 若通过宿主机端口连接（与 docker-compose.yml 的 5433:5432 对齐）；
# DB_HOST=jobView    # 若后端在 Docker 网络内直连容器
# DB_PORT=5432
DB_USER=ltx
DB_PASSWORD=<你的密码>
DB_NAME=jobView_db

# JWT、文件上传等其余配置（按需补充）
JWT_SECRET=<安全的随机字符串>
```

3) 使配置生效（仅当改动了 unit/env 文件）
```bash
sudo systemctl daemon-reload
```

> 后端程序启动后会自动执行数据库迁移与默认流转模板补齐逻辑，无需手动建表。

---

## 六、CI/CD（GitHub Actions）部署

工作流文件：`.github/workflows/deploy.yml`

需要配置的 GitHub Secrets：
- `SERVER_HOST`：服务器 IP 或域名
- `SERVER_USER`：SSH 用户（具备 sudo 权限）
- `SERVER_SSH_KEY`：该用户的私钥（PEM 格式）
- `REMOTE_PATH_BACKEND`：后端二进制部署目录（例：`/opt/jobview`）
- `REMOTE_PATH_FRONTEND`：前端静态文件目录（例：`/www/wwwroot/jobview.bfsmlt.top`）

触发方式：
- 推送到 `main` 分支（包含 backend/ 或 frontend/ 改动），或
- 手动触发：GitHub → Actions → Deploy JobView Project → Run workflow

工作流主要步骤：
- 构建后端（Go 交叉编译 Linux amd64 输出 `backend/jobview-backend`）
- 构建前端（`frontend/dist/`）
- 通过 SSH 停止远端后端服务
- 通过 SCP 将后端二进制与前端静态文件上传到远端目标目录
- 通过 SSH 重启后端服务

部署完成后，远端应存在：
- `/opt/jobview/jobview-backend`（可执行）
- `/www/wwwroot/jobview.bfsmlt.top/*`（前端静态资源）

---

## 七、手工部署（CI/CD 不可用时的备用方案）

1) 本地/CI 构建
```bash
# 后端
cd backend
GOOS=linux GOARCH=amd64 go build -o jobview-backend ./cmd

# 前端
cd ../frontend
npm ci && npm run build
```

2) 文件上传
```bash
# 后端二进制
scp backend/jobview-backend <user>@<host>:/opt/jobview/
ssh <user>@<host> 'sudo chmod +x /opt/jobview/jobview-backend'

# 前端静态资源
rsync -av --delete frontend/dist/ <user>@<host>:/www/wwwroot/jobview.bfsmlt.top/
```

3) 启动后端
```bash
sudo systemctl start jobview-backend
sudo systemctl status jobview-backend --no-pager -l || true
```

---

## 八、Nginx 配置与重载

示例（与你现有线上一致）：
```nginx
server {
    listen 80;
    listen 443 ssl http2;
    server_name jobview.bfsmlt.top;
    root /www/wwwroot/jobview.bfsmlt.top;
    index index.html index.htm;

    if ($server_port !~ 443) {
        rewrite ^(/.*)$ https://$host$1 permanent;
    }

    ssl_certificate     /www/server/panel/vhost/cert/jobview.bfsmlt.top/fullchain.pem;
    ssl_certificate_key /www/server/panel/vhost/cert/jobview.bfsmlt.top/privkey.pem;
    ssl_protocols TLSv1.1 TLSv1.2 TLSv1.3;
    ssl_ciphers EECDH+CHACHA20:EECDH+AES128:EECDH+AES256:RSA+AES128:RSA+AES256:!MD5;
    add_header Strict-Transport-Security "max-age=31536000";
    error_page 497 https://$host$request_uri;

    location / { try_files $uri $uri/ /index.html; }

    location /api/ {
        proxy_pass http://127.0.0.1:8010;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 其余静态/默认配置略
}
```

生效：
```bash
sudo nginx -t && sudo nginx -s reload
```

> CORS 建议：由后端中间件按白名单回显 `Access-Control-Allow-Origin`，避免在 Nginx 侧硬编码 Origin 造成覆盖。

---

## 九、上线后验证清单（必做）

- 基础健康：
  - `curl -s https://jobview.bfsmlt.top/health`
  - 期望返回 `{ "status": "ok", ... }`
- 登录/注册（若允许注册）
- 看板拖拽测试：
  - “笔试中 → 一面中”拖拽应可成功
  - Network 检查 `GET /api/v1/status-transitions/笔试中` 返回包含“一面中”
  - `POST /api/v1/job-applications/{id}/status` 返回 200
- 导出、状态历史、时间轴等页面打开正常
- 日志快速巡检：
  - `journalctl -u jobview-backend -n 200 --no-pager`

---

## 十、回滚策略（可选）

- 若使用第“四节”的备份，可按需回灌：
  - 前端：`sudo tar -C /www/wwwroot -xzf /root/jobview-frontend-*.tgz`
  - 后端：`sudo cp /root/jobview-backend.bak-* /opt/jobview/jobview-backend && sudo chmod +x /opt/jobview/jobview-backend && sudo systemctl restart jobview-backend`
- 数据库：如需回滚，请提前做逻辑备份（此处默认“无用户数据”，未做备份）。

---

## 附录 A：数据库连通性与模板检查

```bash
# 进入 psql
docker exec -it jobView psql -U ltx -d jobView_db

# 检查默认模板
SELECT id, is_default, is_active FROM status_flow_templates;
SELECT flow_config->'transitions'->'笔试中' AS from_written
FROM status_flow_templates WHERE is_default=true AND is_active=true;

# 检查实例信息（确认后端连接的是同一个库）
SELECT current_database(), current_user, inet_server_addr(), inet_server_port();
```

---

## 附录 B：常见问题

- 部署后拖拽仍“不允许”但快速更新成功：
  - 多为 transitions 接口返回缺项。确认默认模板中 `笔试中` 包含 `一面中`，并确认后端连接的数据库实例一致；重启后端使“模板补齐逻辑”生效。
- `Text file busy`：
  - 部署前未停止服务。请先 `systemctl stop jobview-backend` 再覆盖二进制。
- CORS 报错：
  - 确保后端 CORSMiddleware 白名单包含生产域；不要在 Nginx 中硬编码到 `http://localhost:3000`。
