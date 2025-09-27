# 🔧 故障排除指南

> 常见问题解决方案和技术支持

## 🚨 紧急问题

### 🔴 系统无法启动
#### 后端服务启动失败
```bash
# 检查端口占用
netstat -tlnp | grep :8010
lsof -i :8010

# 如果端口被占用，停止占用进程
kill -9 <PID>

# 检查环境变量
env | grep DB_
env | grep JWT_SECRET

# 查看详细错误日志
tail -f /var/log/jobview/app.log
journalctl -u jobview-backend -f
```

#### 前端无法加载
```bash
# 检查Nginx配置
nginx -t

# 重新加载配置
systemctl reload nginx

# 检查文件权限
ls -la /var/www/jobview/frontend/dist/
chmod -R 755 /var/www/jobview/frontend/dist/

# 检查构建文件
cd frontend && npm run build
```

### 🔴 数据库连接失败
```bash
# 测试数据库连接
psql -h localhost -U jobview_user -d jobview_db

# 检查PostgreSQL状态
systemctl status postgresql
systemctl restart postgresql

# 查看数据库日志
tail -f /var/log/postgresql/postgresql-*.log

# 检查数据库配置
cat /etc/postgresql/*/main/postgresql.conf | grep listen_addresses
cat /etc/postgresql/*/main/pg_hba.conf
```

## 💻 客户端问题

### 🔵 登录问题

#### 用户无法登录
**问题现象**: 输入正确用户名密码后提示登录失败

**解决方案**:
```bash
# 1. 检查后端API状态
curl http://localhost:8010/api/v1/health

# 2. 验证用户账号
psql -h localhost -U jobview_user -d jobview_db
SELECT username, email FROM users WHERE username = 'testuser';

# 3. 重置用户密码
UPDATE users SET password = '$2a$10$...' WHERE username = 'testuser';

# 4. 检查JWT配置
echo $JWT_SECRET | wc -c  # 应该>=32字符
```

#### Token过期问题
**问题现象**: 频繁提示重新登录

**解决方案**:
1. 检查系统时间是否正确
2. 确认JWT_SECRET配置一致
3. 清除浏览器缓存和localStorage
4. 检查token刷新逻辑

### 🔵 银月助手问题

#### 银月无法回答问题
**问题现象**: 银月总是回复"暂时无法回答"

**解决方案**:
```bash
# 1. 检查LLM服务状态
curl http://localhost:11434/api/tags

# 2. 重启Ollama服务
systemctl restart ollama
# 或
ollama serve

# 3. 检查模型是否存在
ollama list

# 4. 重新下载模型
ollama pull qwen2.5-coder:latest

# 5. 检查前端控制台错误
# 打开浏览器开发者工具查看Console
```

#### LLM响应很慢
**解决方案**:
1. 使用更轻量的模型（qwen2.5:3b）
2. 调整温度参数（temperature: 0.3）
3. 减少最大token数量（maxTokens: 200）
4. 确保足够的系统资源

### 🔵 音乐播放器问题

#### 音乐无法播放
**问题现象**: 点击播放按钮没有声音

**解决方案**:
```bash
# 1. 检查音频文件是否存在
ls -la frontend/public/music/

# 2. 验证文件格式和路径
# 确保文件路径与playlist配置一致

# 3. 检查浏览器音频权限
# 在浏览器地址栏查看权限设置

# 4. 测试音频文件
# 直接访问音频文件URL测试
```

#### 音质有问题
**解决方案**:
1. 使用高质量音频文件（320kbps MP3+）
2. 检查音频文件是否损坏
3. 调整系统音量和播放器音量比例
4. 尝试不同的音频格式

## 🖥️ 服务器问题

### 🟡 性能问题

#### 响应速度慢
**诊断步骤**:
```bash
# 1. 检查系统资源
top
htop
free -h
df -h

# 2. 查看数据库性能
curl http://localhost:8010/api/v1/stats/database

# 3. 分析慢查询
psql -U jobview_user -d jobview_db
SELECT query, mean_time, calls FROM pg_stat_statements
ORDER BY total_time DESC LIMIT 10;

# 4. 检查连接池状态
curl http://localhost:8010/api/v1/stats/connection-pool
```

**优化方案**:
```sql
-- 重建统计信息
ANALYZE;

-- 重建索引
REINDEX DATABASE jobview_db;

-- 清理无用数据
VACUUM ANALYZE job_applications;
```

#### 内存使用过高
**解决方案**:
```bash
# 1. 调整PostgreSQL配置
shared_buffers = 128MB          # 减少内存使用
work_mem = 4MB                  # 限制工作内存
maintenance_work_mem = 64MB     # 维护操作内存

# 2. 优化Go程序
export GOGC=50  # 更频繁的垃圾回收
export GOMEMLIMIT=500MiB  # 内存限制

# 3. 使用内存监控
go tool pprof http://localhost:8010/debug/pprof/heap
```

### 🟡 数据问题

#### 数据统计不准确
**解决方案**:
```sql
-- 刷新统计信息
ANALYZE job_applications;

-- 检查数据完整性
SELECT status, COUNT(*) FROM job_applications GROUP BY status;

-- 重建统计缓存
DELETE FROM pg_stat_statements;
```

#### 数据丢失或损坏
**恢复步骤**:
```bash
# 1. 停止应用服务
systemctl stop jobview-backend

# 2. 恢复数据库备份
psql -U jobview_user -d jobview_db < /backup/latest.sql

# 3. 验证数据完整性
psql -U jobview_user -d jobview_db
SELECT COUNT(*) FROM job_applications;

# 4. 重启服务
systemctl start jobview-backend
```

## 🌐 网络问题

### 🟢 连接问题

#### CORS错误
**问题现象**: 浏览器控制台显示CORS错误

**解决方案**:
```go
// 后端配置 (main.go)
config := cors.DefaultConfig()
config.AllowOrigins = []string{
    "http://localhost:3000",
    "https://your-domain.com",
}
config.AllowCredentials = true
router.Use(cors.New(config))
```

#### API请求失败
**诊断步骤**:
```bash
# 1. 测试网络连通性
ping your-server-ip
telnet your-server-ip 8010

# 2. 检查防火墙设置
firewall-cmd --list-ports
ufw status

# 3. 测试API接口
curl -v http://localhost:8010/api/v1/health
```

## 🔍 日志分析

### 📄 日志位置
```bash
# 应用日志
/var/log/jobview/app.log
/var/log/jobview/error.log

# Nginx日志
/var/log/nginx/access.log
/var/log/nginx/error.log

# PostgreSQL日志
/var/log/postgresql/postgresql-*.log

# 系统日志
journalctl -u jobview-backend
journalctl -u postgresql
```

### 📊 日志分析
```bash
# 分析错误频率
grep -i error /var/log/jobview/app.log | wc -l

# 分析访问模式
awk '{print $1}' /var/log/nginx/access.log | sort | uniq -c | sort -nr

# 分析响应时间
awk '{print $4}' /var/log/nginx/access.log | sort -n

# 分析数据库连接
grep "connection" /var/log/postgresql/postgresql-*.log
```

## 🧰 诊断工具

### 🔧 系统诊断脚本
```bash
#!/bin/bash
# diagnosis.sh

echo "=== JobView 系统诊断 ==="
echo "时间: $(date)"
echo

echo "1. 系统资源:"
free -h
df -h
echo

echo "2. 服务状态:"
systemctl is-active jobview-backend
systemctl is-active postgresql
systemctl is-active nginx
echo

echo "3. 网络连接:"
netstat -tlnp | grep -E ":8010|:5432|:80"
echo

echo "4. 进程信息:"
ps aux | grep -E "jobview|postgres|nginx" | grep -v grep
echo

echo "5. 最近错误:"
tail -20 /var/log/jobview/error.log 2>/dev/null || echo "无错误日志"
```

### 🔍 性能监控脚本
```bash
#!/bin/bash
# monitor.sh

while true; do
    echo "$(date): CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}'), Memory: $(free | grep Mem | awk '{printf "%.1f%%", $3/$2 * 100.0}')"

    # 检查API响应时间
    response_time=$(curl -o /dev/null -s -w '%{time_total}' http://localhost:8010/api/v1/health)
    echo "API响应时间: ${response_time}s"

    sleep 30
done
```

## 📞 获取帮助

### 🆘 紧急联系
- **紧急热线**: 400-123-4567
- **技术支持**: tech@jobview.com
- **问题反馈**: [GitHub Issues](https://github.com/your-repo/issues)

### 📋 问题报告模板
```
### 问题描述
请详细描述遇到的问题

### 复现步骤
1.
2.
3.

### 预期行为
描述期望的正确行为

### 实际行为
描述实际发生的情况

### 环境信息
- 操作系统:
- 浏览器版本:
- JobView版本:
- 其他相关信息:

### 错误日志
请粘贴相关的错误日志
```

### 🔧 自助服务
- **文档中心**: [docs/](../README.md)
- **API文档**: [api/](../api/README.md)
- **部署指南**: [deployment/](../deployment/README.md)
- **银月助手**: 使用内置AI助手获取即时帮助

---

**🔧 快速解决问题，保障系统稳定！** ✨

> **无法解决？** 联系技术支持团队获取专业帮助！