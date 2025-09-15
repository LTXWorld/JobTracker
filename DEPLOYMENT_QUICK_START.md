# JobView 快速部署指南

本指南帮助您快速将 JobView 系统部署到云服务器。

## 🚀 快速开始（5分钟部署）

### 1. 准备服务器

```bash
# SSH登录到您的云服务器
ssh root@your-server-ip

# 安装Docker（如果尚未安装）
curl -fsSL https://get.docker.com | sh

# 安装Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
```

### 2. 获取代码

```bash
# 克隆项目
cd /opt
git clone https://github.com/yourusername/jobview.git
cd jobview

# 或者上传代码包
scp jobview.tar.gz root@your-server-ip:/opt/
ssh root@your-server-ip
cd /opt && tar -xzf jobview.tar.gz && cd jobview
```

### 3. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置（必须修改的项）
nano .env

# 必须修改：
# DB_PASSWORD=设置一个强密码
# JWT_SECRET=设置至少32字符的随机字符串
# DOMAIN=您的域名（如：jobview.example.com）
```

### 4. 一键部署

```bash
# 使用部署脚本
./scripts/deploy.sh

# 或手动部署
docker-compose -f docker-compose.production.yml up -d --build
```

### 5. 验证部署

```bash
# 运行健康检查
./scripts/health-check.sh

# 查看服务状态
docker-compose -f docker-compose.production.yml ps

# 查看日志
docker-compose -f docker-compose.production.yml logs -f
```

## 🌐 配置域名访问

### 1. 配置DNS

在您的域名服务商控制台添加A记录：
- 类型：A
- 主机记录：jobview（或@）
- 记录值：您的服务器IP
- TTL：600

### 2. 配置SSL证书（HTTPS）

```bash
# 安装Certbot
apt install certbot -y

# 停止前端服务（释放80端口）
docker-compose -f docker-compose.production.yml stop frontend

# 申请证书
certbot certonly --standalone -d jobview.example.com

# 重启服务
docker-compose -f docker-compose.production.yml up -d
```

### 3. 更新Nginx配置

已包含SSL配置，证书会自动挂载到容器中。

## 📊 管理命令

### 查看状态
```bash
docker-compose -f docker-compose.production.yml ps
```

### 重启服务
```bash
docker-compose -f docker-compose.production.yml restart
```

### 停止服务
```bash
docker-compose -f docker-compose.production.yml down
```

### 查看日志
```bash
# 所有服务
docker-compose -f docker-compose.production.yml logs -f

# 特定服务
docker-compose -f docker-compose.production.yml logs -f backend
```

### 进入容器
```bash
# 数据库
docker exec -it jobview-db psql -U jobview -d jobview_db

# 后端
docker exec -it jobview-backend sh

# 前端
docker exec -it jobview-frontend sh
```

## 💾 数据备份

### 手动备份
```bash
./scripts/backup.sh
```

### 自动备份（每天凌晨2点）
```bash
# 添加到crontab
crontab -e

# 添加以下行
0 2 * * * /opt/jobview/scripts/backup.sh >> /var/log/jobview-backup.log 2>&1
```

### 恢复备份
```bash
# 恢复数据库
gunzip < /backup/jobview/db_backup_20240101_020000.sql.gz | \
  docker exec -i jobview-db psql -U jobview -d jobview_db
```

## 🔧 故障排查

### 服务无法启动
```bash
# 检查端口占用
netstat -tulpn | grep -E "80|443|8010|5432"

# 查看错误日志
docker-compose -f docker-compose.production.yml logs --tail=50
```

### 数据库连接失败
```bash
# 检查数据库状态
docker exec jobview-db pg_isready -U jobview

# 查看数据库日志
docker logs jobview-db --tail=50
```

### 前端无法访问
```bash
# 检查Nginx配置
docker exec jobview-frontend nginx -t

# 重启前端
docker-compose -f docker-compose.production.yml restart frontend
```

## 🔄 更新应用

```bash
# 备份数据
./scripts/backup.sh

# 拉取最新代码
git pull origin main

# 重新构建并部署
docker-compose -f docker-compose.production.yml down
docker-compose -f docker-compose.production.yml up -d --build

# 验证更新
./scripts/health-check.sh
```

## 📝 默认配置

- 前端端口：80, 443（HTTPS）
- 后端端口：8010
- 数据库端口：5432（仅容器内部）
- 默认测试账号：testuser / TestPass123!

## 🛡️ 安全建议

1. **修改默认密码**：首次部署后立即修改所有默认密码
2. **配置防火墙**：只开放80、443、22端口
3. **定期备份**：配置自动备份脚本
4. **监控日志**：定期检查系统日志
5. **更新依赖**：定期更新Docker镜像和依赖包

## 📞 获取帮助

- 查看完整文档：`docs/system-data-flow-and-deployment-guide.md`
- 提交问题：[GitHub Issues](https://github.com/yourusername/jobview/issues)

---

祝您部署顺利！🎉