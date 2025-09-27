# 🚀 JobView 完整部署指南

> 基于Docker + GitHub Actions的现代化CI/CD部署解决方案

## 📋 部署文档导航

### 🐳 Docker容器化部署
- [🐳 **Docker容器化部署指南**](./docker-deployment.md) - 完整的Docker配置和本地开发环境
- [🚀 **GitHub Actions CI/CD**](./github-actions-cicd.md) - 自动化构建、测试和部署流程
- [🛠️ **自动化部署脚本**](./automation-scripts.md) - 服务器初始化和管理脚本

### 📚 传统部署方案 (备用)
- [🚀 **生产环境完整部署指南**](./production-deployment-guide.md) - 传统PM2部署方案
- [🗄️ **数据库迁移脚本**](./database-migration.md) - 数据库字段迁移解决方案
- [🛣️ **路径配置指南**](./path-configuration.md) - 路径一致性配置

## 🎯 推荐部署方案

### ⚡ Docker + CI/CD 一键部署 (推荐)

这是完全自动化的现代部署方案，当您推送代码到main分支时会自动触发部署：

```bash
# 1. 克隆代码到服务器
git clone https://github.com/your-username/jobView.git /www/wwwroot/jobview.bfsmlt.top
cd /www/wwwroot/jobview.bfsmlt.top

# 2. 运行一键部署脚本
chmod +x scripts/deploy.sh
./scripts/deploy.sh --domain jobview.bfsmlt.top --env production

# 3. 验证部署
curl https://jobview.bfsmlt.top/api/v1/health
```

### 🔄 自动部署流程
```
代码推送到main分支
    ↓
GitHub Actions 自动触发
    ├── 运行前端和后端测试
    ├── 构建Docker镜像
    ├── 推送到GitHub Container Registry
    └── 部署到生产服务器
        ↓
自动健康检查和通知
```

## 📊 Docker容器化架构

### 🏗️ 现代化容器架构
```
用户浏览器 (HTTPS)
    ↓
Nginx Container (jobview-nginx:443)
    ├── / → Frontend Container (jobview-frontend:80)
    └── /api/ → Backend Container (jobview-backend:8010)
        ↓
PostgreSQL Container (jobview-postgres:5432)
        ↓
Redis Container (jobview-redis:6379) [可选]
```

### 🐳 Docker服务组件
- **jobview-frontend**: Vue 3 + Nginx → 静态文件服务
- **jobview-backend**: Go + Gin → API服务容器
- **jobview-postgres**: PostgreSQL → 数据持久化
- **jobview-redis**: Redis → 缓存和会话存储
- **jobview-nginx**: Nginx → 反向代理和SSL终端

### 🔄 CI/CD组件
- **GitHub Actions**: 自动构建、测试、部署
- **GitHub Container Registry**: Docker镜像存储
- **Health Checks**: 自动服务监控
- **Backup System**: 数据库自动备份

## 🗄️ 数据库迁移和配置

### ⚠️ Docker环境下的数据库管理
使用Docker容器化部署时，数据库迁移会自动执行：

```bash
# 数据库迁移在容器启动时自动执行
docker-compose up -d

# 手动执行迁移 (如需要)
./scripts/migrate-database.sh

# 验证数据库结构
docker-compose exec postgres psql -U jobview_user -d jobview_prod -c "\d job_applications"
```

### 🔑 自动创建的关键字段
Docker部署会自动确保以下字段存在：
- `job_applications.status_version` (状态版本控制)
- `job_applications.status_history` (状态历史JSON)
- `job_applications.last_status_change` (最后状态变更时间)
- `status_history` 表 (状态历史记录)
- `user_preferences` 表 (用户偏好设置)

### 💾 数据持久化
```bash
# Docker volumes确保数据持久化
docker volume ls | grep jobview
# postgres_data - 数据库数据
# backend_logs - 后端日志
# backend_uploads - 文件上传
```

## 🛣️ Docker环境路径配置

### 🎯 容器化路径配置原则
1. **容器内统一**: 所有服务在容器内使用标准路径
2. **主机映射**: 通过Docker volumes映射到主机路径
3. **环境适配**: 开发和生产使用相同的容器配置
4. **自动代理**: Nginx容器自动处理前后端路由

### 📂 Docker Compose路径配置

#### 关键路径映射
```yaml
services:
  frontend:
    # 容器内: /usr/share/nginx/html
    # 自动构建: 不需要主机路径映射

  backend:
    volumes:
      - backend_logs:/root/logs          # 日志持久化
      - backend_uploads:/root/uploads    # 文件上传持久化

  postgres:
    volumes:
      - postgres_data:/var/lib/postgresql/data  # 数据库持久化

  nginx:
    volumes:
      - ./nginx/conf.d:/etc/nginx/conf.d        # Nginx配置
      - /www/server/panel/vhost/cert/jobview.bfsmlt.top:/etc/nginx/ssl  # SSL证书
```

#### 环境变量配置
```bash
# Docker Compose 环境变量 (.env.production)
DOMAIN=jobview.bfsmlt.top
DB_PASSWORD=your_secure_password
JWT_SECRET=your-super-secret-jwt-key-must-be-at-least-32-characters-long

# 容器内部配置 (自动设置)
DB_HOST=postgres      # 容器服务名
REDIS_HOST=redis      # 容器服务名
VITE_API_BASE=/api    # 前端API路径
```

## 🔍 Docker部署验证清单

### ✅ 部署前检查
- [ ] 服务器已安装Docker和Docker Compose
- [ ] GitHub Secrets已正确配置
- [ ] SSL证书路径配置正确
- [ ] 防火墙端口开放 (80, 443, 22)
- [ ] 域名DNS解析指向服务器

### ✅ 部署过程检查
- [ ] 代码成功推送到main分支
- [ ] GitHub Actions工作流成功运行
- [ ] Docker镜像构建成功
- [ ] 容器成功启动并运行
- [ ] 数据库迁移自动执行成功

### ✅ 部署后验证
- [ ] 所有Docker容器状态为"Up"
- [ ] 前端页面可以正常访问
- [ ] API接口响应正常
- [ ] 数据库连接和查询正常
- [ ] 银月助手功能正常
- [ ] 音乐播放器功能正常

### 🧪 Docker环境功能测试
```bash
# 1. 检查Docker容器状态
docker-compose ps

# 2. 健康检查
curl https://jobview.bfsmlt.top/api/v1/health

# 3. 用户登录测试
curl -X POST https://jobview.bfsmlt.top/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"TestPass123!"}'

# 4. 前端路由测试
curl -I https://jobview.bfsmlt.top/dashboard

# 5. 数据库连接测试
docker-compose exec postgres pg_isready -U jobview_user -d jobview_prod

# 6. 容器日志检查
docker-compose logs backend
docker-compose logs frontend
```

## 🔧 Docker环境故障排除

### 🚨 常见Docker问题

#### 1. 容器无法启动
```bash
# 检查容器状态和日志
docker-compose ps
docker-compose logs [service-name]

# 重启特定服务
docker-compose restart [service-name]

# 完全重新部署
docker-compose down
docker-compose up -d --build
```

#### 2. 前端页面无法访问
```bash
# 检查Nginx容器状态
docker-compose logs nginx

# 检查前端容器构建
docker-compose logs frontend

# 重启前端服务
docker-compose restart frontend nginx
```

#### 3. API请求失败
```bash
# 检查后端容器状态
docker-compose logs backend

# 检查后端健康状态
docker-compose exec backend wget -qO- http://localhost:8010/api/v1/health

# 检查容器网络连接
docker network ls
docker network inspect jobview_jobview-network
```

#### 4. 数据库连接失败
```bash
# 检查PostgreSQL容器状态
docker-compose logs postgres

# 测试数据库连接
docker-compose exec postgres pg_isready -U jobview_user -d jobview_prod

# 重启数据库 (注意数据安全)
docker-compose restart postgres
```

#### 5. GitHub Actions部署失败
```bash
# 检查服务器上的部署日志
cd /www/wwwroot/jobview.bfsmlt.top
git log --oneline -5

# 检查Docker镜像
docker images | grep jobview

# 手动拉取最新镜像
docker login ghcr.io
docker pull ghcr.io/your-username/jobview/frontend:latest
docker pull ghcr.io/your-username/jobview/backend:latest
```

### 🔍 日志查看和调试
```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres

# 进入容器调试
docker-compose exec backend /bin/sh
docker-compose exec postgres psql -U jobview_user -d jobview_prod

# 检查容器资源使用
docker stats

# 检查Docker系统信息
docker system df
docker system info
```

## 🔄 Docker环境更新和维护

### 📦 自动更新流程 (推荐)
Docker + CI/CD环境下，更新变得非常简单：

```bash
# 1. 推送代码到main分支 (GitHub Actions自动处理后续步骤)
git add .
git commit -m "你的更新内容"
git push origin main

# 2. 监控部署进度
# 访问 GitHub Actions 页面查看部署状态
# 或在服务器上监控容器更新:
watch docker-compose ps
```

### 🔄 手动更新流程 (紧急情况)
```bash
# 1. 进入项目目录
cd /www/wwwroot/jobview.bfsmlt.top

# 2. 备份数据
docker-compose exec -T postgres pg_dump -U jobview_user jobview_prod > backup_$(date +%Y%m%d_%H%M%S).sql

# 3. 拉取最新代码
git pull origin main

# 4. 重新部署
docker-compose down
docker-compose up -d --build

# 5. 验证更新
docker-compose ps
curl https://jobview.bfsmlt.top/api/v1/health
```

### 🔄 服务管理脚本 (推荐使用)
```bash
# 使用提供的管理脚本
./scripts/manage-service.sh status    # 查看状态
./scripts/manage-service.sh restart   # 重启服务
./scripts/manage-service.sh update    # 更新服务
./scripts/manage-service.sh logs      # 查看日志
./scripts/manage-service.sh backup    # 备份数据
```

### 📊 Docker环境性能监控
```bash
# 容器资源使用监控
docker stats

# Docker系统资源使用
docker system df

# 检查容器健康状态
docker-compose ps
docker inspect $(docker-compose ps -q) --format='table {{.Name}}\t{{.State.Health.Status}}'

# 数据库性能监控 (通过API)
curl https://jobview.bfsmlt.top/api/v1/stats/database
curl https://jobview.bfsmlt.top/api/v1/stats/connection-pool

# 自动化监控 (使用监控脚本)
./scripts/monitor.sh health   # 健康检查
./scripts/monitor.sh report   # 性能报告
./scripts/monitor.sh all      # 完整检查
```

### 🔐 Docker环境安全维护
```bash
# 定期清理未使用的资源
docker system prune -f
docker volume prune -f
docker image prune -f

# 更新Docker镜像
docker-compose pull
docker-compose up -d

# 检查容器安全
docker scan $(docker images -q)

# 备份Docker volumes
docker run --rm -v jobview_postgres_data:/from -v $(pwd):/to alpine sh -c "cd /from ; tar -czf /to/postgres_backup_$(date +%Y%m%d).tar.gz ."
```

### ⏰ 定时任务配置
```bash
# 配置定时任务 (使用脚本)
./scripts/setup-cron.sh

# 手动配置crontab
crontab -e
# 添加以下内容:
# 每15分钟健康检查
*/15 * * * * cd /www/wwwroot/jobview.bfsmlt.top && ./scripts/monitor.sh health
# 每天凌晨2点备份
0 2 * * * cd /www/wwwroot/jobview.bfsmlt.top && ./scripts/manage-service.sh backup
# 每周清理Docker资源
0 3 * * 0 docker system prune -f
```

## 📞 获取支持

### 🆘 Docker环境紧急情况处理
- **容器无法启动**: 检查Docker服务和容器日志
- **数据库连接错误**: 检查PostgreSQL容器状态，必要时重启
- **GitHub Actions部署失败**: 检查Secrets配置和服务器连接
- **网站无法访问**: 检查Nginx容器和SSL证书配置

### 📋 技术支持资源
- **Docker部署**: [Docker容器化部署指南](./docker-deployment.md)
- **CI/CD流程**: [GitHub Actions CI/CD](./github-actions-cicd.md)
- **自动化脚本**: [自动化部署脚本](./automation-scripts.md)
- **故障排除**: [完整故障排除指南](../TROUBLESHOOTING.md)

### 🔧 配置文件参考
- **Docker配置**: `docker-compose.yml`, `frontend/Dockerfile`, `backend/Dockerfile`
- **Nginx配置**: `nginx/conf.d/jobview.conf`
- **GitHub Actions**: `.github/workflows/deploy.yml`
- **环境变量**: `.env.production`

### 📖 快速参考

#### 常用Docker命令
```bash
# 查看所有容器状态
docker-compose ps

# 查看日志
docker-compose logs [service-name]

# 重启服务
docker-compose restart [service-name]

# 重新部署
docker-compose down && docker-compose up -d --build

# 进入容器
docker-compose exec [service-name] /bin/sh
```

#### GitHub Actions状态检查
- 访问 `https://github.com/your-username/jobView/actions`
- 查看最新的workflow运行状态
- 检查部署日志和错误信息

#### 服务器文件位置
- **项目根目录**: `/www/wwwroot/jobview.bfsmlt.top/`
- **SSL证书**: `/www/server/panel/vhost/cert/jobview.bfsmlt.top/`
- **Docker数据**: Docker volumes (自动管理)
- **备份目录**: `/backup/jobview/`

---

## 🎉 部署完成验证

### ✅ 最终验证清单
1. **基础功能**
   - [ ] 网站可正常访问: `https://jobview.bfsmlt.top`
   - [ ] API健康检查通过: `https://jobview.bfsmlt.top/api/v1/health`
   - [ ] 用户可正常登录注册

2. **核心功能**
   - [ ] 求职记录管理功能正常
   - [ ] 看板拖拽功能正常
   - [ ] 数据统计图表显示正常
   - [ ] 银月智能助手回答正常
   - [ ] 音乐播放器功能正常

3. **性能和稳定性**
   - [ ] 页面加载速度正常 (< 3秒)
   - [ ] API响应时间正常 (< 500ms)
   - [ ] 数据库查询性能良好
   - [ ] 所有Docker容器健康运行

4. **自动化流程**
   - [ ] GitHub Actions自动部署正常
   - [ ] 代码推送后自动更新生效
   - [ ] 健康检查和监控正常
   - [ ] 定时备份任务执行正常

🚀 **恭喜！JobView已成功部署到生产环境，享受现代化的Docker + CI/CD自动部署体验！**

> **重要提醒**:
> 1. 定期监控GitHub Actions工作流状态
> 2. 确保SSL证书定期更新
> 3. 定期检查Docker镜像安全更新
> 4. 保持服务器系统和Docker版本最新