## JobView 上线部署全流程（Docker + GitHub Actions + CI/CD）

### 1. 服务器基础准备
1. 系统更新与工具安装
   ```bash
   sudo apt update && sudo apt upgrade -y
   sudo apt install -y git curl ufw
   ```
2. 安装 Docker 与 Compose 插件
   ```bash
   curl -fsSL https://get.docker.com | bash
   sudo apt install -y docker-compose-plugin
   sudo usermod -aG docker <部署用户>
   ```
3. 防火墙配置
   ```bash
   sudo ufw allow OpenSSH
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```
4. 目录结构
   ```
   /opt/jobview/
     ├── docker-compose.yml
     ├── backend/
     ├── frontend/
     ├── extension/
     ├── scripts/deploy.sh
     └── .env
   ```

### 2. Docker 化与配置管理
| 服务 | 关键点 |
| --- | --- |
| PostgreSQL | 使用 `postgres:15`，挂载 `/opt/jobview/data/postgres`，通过 `.env` 注入 `POSTGRES_*`，首次初始化执行迁移/`init.sql`。 |
| 后端 (Go) | 多阶段镜像，读取数据库、JWT、前端 URL 环境变量，启动命令 `./jobview-backend` 自动迁移。 |
| 前端 (Vue) | `node:18` 构建，`npm ci && npm run build` 后用 `nginx:alpine` 托管 `dist/`，配置 HTTPS + SPA 重写。 |
| 扩展 | CI 内打包 zip；发布至 Release/对象存储。 |
| `.env` | 包含 `DB_HOST/PORT/USER/PASSWORD`、`JWT_SECRET`、`API_BASE_URL`、`FRONTEND_URL` 等，生产版通过 Secrets 注入。 |

### 3. GitHub Actions 流水线
**触发策略**
- `push` 至 `feature/*`、PR → `main`：仅执行 CI（构建/测试/lint）。
- `push` 至 `main` 或手动 `workflow_dispatch`：执行 CD（镜像构建 + 部署）。

**CI 工作流 `ci.yml` 要点**
- Backend：`go test ./...`，缓存 Go modules。
- Frontend：`npm ci && npm run build && npm test`，上传 `frontend/dist` 为 artifact。
- Extension：`npm install && npm run build`，上传扩展 zip。
- 可选：`golangci-lint`、`eslint`。

**CD 工作流 `deploy.yml` 要点**
```yaml
name: Deploy
on:
  push:
    branches: [main]
  workflow_dispatch:
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: jobview-frontend-dist
          path: frontend/dist
      - name: Docker login
        run: docker login ${{ secrets.REGISTRY_URL }} -u ${{ secrets.REGISTRY_USER }} -p ${{ secrets.REGISTRY_PAT }}
      - name: Build & push backend image
        run: |
          docker build -t ${{ secrets.REGISTRY_URL }}/jobview-backend:${{ github.sha }} backend
          docker push ${{ secrets.REGISTRY_URL }}/jobview-backend:${{ github.sha }}
      - name: Build & push frontend image
        run: |
          docker build -t ${{ secrets.REGISTRY_URL }}/jobview-frontend:${{ github.sha }} frontend
          docker push ${{ secrets.REGISTRY_URL }}/jobview-frontend:${{ github.sha }}
      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SERVER_SSH_KEY }}
          script: |
            cd /opt/jobview
            ./scripts/deploy.sh ${{ github.sha }}
```

**Secrets 清单**：`SERVER_HOST`、`SERVER_USER`、`SERVER_SSH_KEY`、`REGISTRY_URL`、`REGISTRY_USER`、`REGISTRY_PAT`、`DB_PASSWORD`、`JWT_SECRET` 等。

### 4. 服务器部署脚本 `scripts/deploy.sh`
```bash
#!/usr/bin/env bash
set -euo pipefail

SHA=${1:-latest}
cd /opt/jobview

docker compose pull

# 可在此写入 .env，使用 GitHub Secrets 渲染

docker compose up -d --remove-orphans
docker image prune -f
curl -f http://localhost:8010/api/v1/health
```

### 5. 运维与监控
- 日志：保留 Docker 默认日志，结合 `logrotate`；可后续引入 Loki/ELK。
- 健康检查：在 Compose 中配置 `healthcheck`，服务设定 `restart: unless-stopped`。
- 备份：`cron` 定期 `pg_dump` 备份数据库到对象存储。
- 监控：基础监控进程与磁盘，可扩展至 `Prometheus + Grafana`。

### 6. Chrome 扩展发布（可选）
- 在 CI 中构建扩展并上传 artifact。
- 如需自动发布到 Chrome Web Store，可集成 `chrome-extension-upload-action` 并提供相关证书。
