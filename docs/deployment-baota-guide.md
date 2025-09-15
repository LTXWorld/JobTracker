# JobView 宝塔面板部署指南

本指南专门针对使用宝塔面板的服务器环境。

## 🎯 部署架构说明

在宝塔环境下，我们采用以下架构：
- **前端**：使用宝塔的Nginx直接服务静态文件（不使用Docker）
- **后端**：使用Docker运行Go应用
- **数据库**：使用Docker运行PostgreSQL
- **SSL证书**：使用宝塔面板管理的Let's Encrypt证书

## 📋 前置条件

1. 已安装宝塔面板
2. 已在宝塔创建站点：`jobview.bfsmlt.top`
3. 已配置SSL证书
4. 已安装Docker（可通过宝塔软件商店安装）

## 🚀 快速部署

### 步骤1：准备项目文件

```bash
# SSH登录服务器
ssh root@your-server-ip

# 创建项目目录
mkdir -p /opt/jobview
cd /opt/jobview

# 克隆或上传项目代码
git clone https://github.com/yourusername/jobview.git .
# 或者上传压缩包后解压
```

### 步骤2：配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件
nano .env

# 必须修改以下配置：
# DB_PASSWORD=设置强密码
# JWT_SECRET=至少32字符的随机字符串
```

### 步骤3：运行部署脚本

```bash
# 添加执行权限
chmod +x scripts/deploy-baota.sh

# 运行部署脚本
./scripts/deploy-baota.sh

# 选择 1 进行完整部署
```

## 📝 宝塔面板配置确认

确保您的宝塔站点配置包含以下内容：

1. **网站目录**：`/www/wwwroot/jobview.bfsmlt.top`
2. **SSL证书**：已配置并强制HTTPS
3. **反向代理**：`/api/` 代理到 `http://127.0.0.1:8010`

您的Nginx配置应该包含：
```nginx
# API反向代理
location /api/ {
    proxy_pass http://127.0.0.1:8010;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

# 支持前端路由
location / {
    try_files $uri $uri/ /index.html;
}
```

## 🔧 日常管理

### 查看服务状态
```bash
cd /opt/jobview
docker-compose -f docker-compose.baota.yml ps
```

### 查看日志
```bash
# 所有服务日志
docker-compose -f docker-compose.baota.yml logs -f

# 仅后端日志
docker-compose -f docker-compose.baota.yml logs -f backend

# 仅数据库日志
docker-compose -f docker-compose.baota.yml logs -f postgres
```

### 重启服务
```bash
# 重启后端
docker-compose -f docker-compose.baota.yml restart backend

# 重启所有服务
docker-compose -f docker-compose.baota.yml restart
```

### 更新应用
```bash
cd /opt/jobview
./scripts/deploy-baota.sh
# 选择 4 进行更新部署
```

## 💾 数据备份

### 手动备份
```bash
./scripts/backup.sh
```

### 设置自动备份
在宝塔面板的计划任务中添加：
- 任务类型：Shell脚本
- 任务名称：JobView备份
- 执行周期：每天凌晨2点
- 脚本内容：`/opt/jobview/scripts/backup.sh`

## 🔍 故障排查

### 1. 前端无法访问
- 检查宝塔站点是否正常运行
- 检查站点目录是否有文件：`ls /www/wwwroot/jobview.bfsmlt.top`
- 检查Nginx配置：在宝塔面板查看站点配置

### 2. API无法访问
- 检查后端容器：`docker ps | grep jobview-backend`
- 查看后端日志：`docker logs jobview-backend`
- 测试本地连接：`curl http://127.0.0.1:8010/api/auth/health`

### 3. 数据库连接失败
- 检查数据库容器：`docker ps | grep jobview-db`
- 测试数据库连接：`docker exec jobview-db pg_isready -U jobview`

### 4. 502错误
- 后端服务未启动：`docker-compose -f docker-compose.baota.yml up -d backend`
- 检查端口：`netstat -tulpn | grep 8010`

## 📊 性能监控

在宝塔面板中可以：
1. 查看服务器资源使用情况
2. 查看网站访问日志
3. 设置监控报警

使用Docker查看容器资源：
```bash
docker stats jobview-backend jobview-db
```

## 🔐 安全建议

1. **定期更新**
   - 通过宝塔面板更新系统软件
   - 定期更新Docker镜像

2. **访问控制**
   - 使用宝塔的安全功能限制SSH访问
   - 配置防火墙规则

3. **备份策略**
   - 使用宝塔的定时备份功能
   - 定期测试备份恢复

## 📞 常见问题

### Q: 如何修改数据库密码？
A: 编辑 `/opt/jobview/.env` 文件，修改 `DB_PASSWORD`，然后重新部署后端。

### Q: 如何查看实时日志？
A: 使用命令 `docker-compose -f docker-compose.baota.yml logs -f backend`

### Q: 如何完全卸载？
A: 运行 `docker-compose -f docker-compose.baota.yml down -v` 并删除项目目录。

---

如有其他问题，请查看项目文档或提交Issue。