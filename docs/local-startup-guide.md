# 🚀 JobView 项目本地启动完整指南

## 📋 环境要求

在开始之前，请确保你的系统安装了以下软件：

- **Node.js** 18.0+ 
- **Go** 1.24+  
- **PostgreSQL** 12+
- **Git**

## 🛠️ 第一步：项目准备

### 1. 确认项目位置
```bash
cd /Users/lutao/GolandProjects/jobView
pwd  # 确认在项目根目录
```

### 2. 检查项目结构
```bash
ls -la
# 应该看到: backend/, frontend/, docs/, .env.example 等文件夹
```

## 🗄️ 第二步：数据库设置

### 1. 启动 PostgreSQL 服务
```bash
# macOS 使用 Homebrew 安装的 PostgreSQL
brew services start postgresql@14

# 或者如果使用不同版本
brew services start postgresql
```

### 2. 创建数据库和用户
```bash
# 连接到 PostgreSQL
psql postgres

# 在 psql 中执行以下命令：
CREATE USER ltx WITH PASSWORD 'your_password_here';
CREATE DATABASE "jobView_db" OWNER ltx;
GRANT ALL PRIVILEGES ON DATABASE "jobView_db" TO ltx;

# 退出 psql
\q
```

### 3. 验证数据库连接
```bash
# 测试连接
psql -h 127.0.0.1 -p 5433 -U ltx -d jobView_db
# 输入密码后应该能连接成功
```

## ⚙️ 第三步：环境配置

### 1. 创建环境变量文件
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量
nano .env  # 或使用你喜欢的编辑器
```

### 2. 配置 .env 文件
```bash
# JobView Backend Environment Variables

# 数据库配置（根据你的实际配置修改）
DB_HOST=127.0.0.1
DB_PORT=5433
DB_USER=ltx
DB_PASSWORD=your_actual_password_here
DB_NAME=jobView_db
DB_SSLMODE=disable

# 服务器配置
SERVER_PORT=8010
ENVIRONMENT=development

# JWT配置（开发环境使用，生产环境必须更改）
JWT_SECRET=your_256_bit_jwt_secret_key_change_in_production_must_be_32_chars_minimum
JWT_ACCESS_DURATION=24h
JWT_REFRESH_DURATION=720h
```

## 🔧 第四步：后端启动

### 1. 进入后端目录
```bash
cd backend
```

### 2. 下载依赖
```bash
# 下载 Go 依赖
go mod download

# 验证依赖
go mod tidy
```

### 3. 运行数据库迁移
```bash
# 启动程序会自动运行迁移
go run cmd/main.go
```

如果看到类似输出说明启动成功：
```
[DB-POOL] Connection pool optimized:
  - MaxOpenConns: 16
  - MaxIdleConns: 5
  - ConnMaxLifetime: 30m0s
  - ConnMaxIdleTime: 15m0s
  - CPU Cores: 8
  - Environment: development
Server starting on :8010
```

### 4. 应用数据库性能优化
打开新终端窗口：
```bash
cd /Users/lutao/GolandProjects/jobView/backend

# 执行数据库优化脚本
chmod +x scripts/migrate_optimization.sh
./scripts/migrate_optimization.sh
```

## 🎨 第五步：前端启动

### 1. 打开新终端，进入前端目录
```bash
cd /Users/lutao/GolandProjects/jobView/frontend
```

### 2. 安装依赖
```bash
# 安装 npm 依赖
npm install

# 如果有权限问题或网络问题，可以尝试
npm install --legacy-peer-deps
```

### 3. 启动前端开发服务器
```bash
npm run dev
```

应该看到类似输出：
```
  VITE v7.1.2  ready in 1234 ms

  ➜  Local:   http://localhost:3000/
  ➜  Network: use --host to expose
  ➜  press h + enter to show help
```

## ✅ 第六步：验证系统运行

### 1. 检查服务状态

**后端服务检查：**
```bash
# 检查后端健康状态
curl http://localhost:8010/api/v1/health

# 检查数据库性能统计
curl http://localhost:8010/api/v1/stats/database
```

**前端服务检查：**
- 打开浏览器访问: http://localhost:3000
- 应该能看到 JobView 的登录/注册页面

### 2. 测试完整流程

1. **注册新用户**
   - 访问 http://localhost:3000
   - 点击"注册"创建新账户

2. **登录系统**
   - 使用刚注册的账户登录

3. **添加投递记录**
   - 登录后添加一条求职投递记录

4. **验证功能**
   - 查看看板视图
   - 检查数据统计
   - 测试搜索功能

## 🔍 第七步：性能监控验证

### 1. 访问性能监控API
```bash
# 数据库性能统计
curl -X GET http://localhost:8010/api/v1/stats/database | jq

# 连接池状态
curl -X GET http://localhost:8010/api/v1/stats/connection-pool | jq
```

### 2. 运行性能测试
```bash
cd backend

# 运行基准测试
go test -bench=. ./tests/service/

# 运行所有测试
go test ./... -v
```

## 🚨 常见问题解决

### 问题1: 数据库连接失败
```bash
# 检查 PostgreSQL 是否运行
brew services list | grep postgresql

# 检查端口是否被占用
lsof -i :5433

# 重启 PostgreSQL
brew services restart postgresql
```

### 问题2: 后端端口被占用
```bash
# 检查8010端口
lsof -i :8010

# 杀死占用进程
kill -9 PID_NUMBER
```

### 问题3: 前端依赖安装失败
```bash
# 清除缓存后重新安装
cd frontend
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

### 问题4: JWT_SECRET 长度不足
确保 .env 文件中的 JWT_SECRET 至少32个字符：
```bash
JWT_SECRET=your_very_long_secret_key_must_be_at_least_32_characters_long_for_security
```

### 问题5: 数据库优化脚本执行失败
```bash
# 手动执行索引创建
cd backend
psql -h 127.0.0.1 -p 5433 -U ltx -d jobView_db -f migrations/004_add_performance_indexes.sql
```

### 问题6: Go 模块下载缓慢
```bash
# 设置 Go 模块代理（中国用户）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

# 然后重新下载
go mod download
```

## 📊 确认优化效果

系统启动后，你应该能看到：

- **查询响应时间**: 20-35ms（而不是150-300ms）
- **慢查询监控**: 实时监控>100ms的查询
- **连接池状态**: 智能配置根据CPU核数调整
- **性能API**: 完整的监控接口可用

## 🔧 开发工具推荐

### VS Code 扩展
- **Go** - Go语言支持
- **Vetur** 或 **Vue Language Features (Volar)** - Vue开发支持
- **PostgreSQL** - 数据库查询支持
- **REST Client** - API测试

### 浏览器开发工具
- **Vue.js devtools** - Vue组件调试
- **网络面板** - 监控API请求
- **控制台** - 查看性能日志

## 📈 性能监控界面

启动成功后，你可以通过以下方式监控系统性能：

### 1. 性能监控API
- 数据库统计: `GET http://localhost:8010/api/v1/stats/database`
- 连接池状态: `GET http://localhost:8010/api/v1/stats/connection-pool`
- 健康检查: `GET http://localhost:8010/api/v1/health`

### 2. 日志监控
后端启动后会在控制台显示：
- 慢查询日志 (>100ms)
- 连接池状态变化
- 数据库健康检查结果

### 3. 前端性能
- 在浏览器开发者工具的Network面板查看API响应时间
- 使用Vue devtools监控组件渲染性能

## 🎯 功能验证清单

启动完成后，请按以下清单验证所有功能：

### 基础功能
- [ ] 用户注册/登录
- [ ] JWT token自动刷新
- [ ] 投递记录CRUD操作
- [ ] 看板视图拖拽
- [ ] 数据统计图表

### 优化功能 
- [ ] 分页查询 (页面加载<50ms)
- [ ] 搜索功能 (响应<20ms)
- [ ] 批量操作 (测试批量导入)
- [ ] 性能监控API (有数据返回)
- [ ] 健康检查 (返回healthy状态)

### 高级功能
- [ ] 提醒功能
- [ ] 数据导出
- [ ] 响应式布局 (移动端适配)
- [ ] 错误处理 (网络异常恢复)

## 🎉 启动完成！

如果以上步骤都成功执行，你现在应该有：

- ✅ **后端服务**: http://localhost:8010 (包含性能优化)
- ✅ **前端应用**: http://localhost:3000 (完整功能界面)  
- ✅ **数据库**: PostgreSQL with 7个优化索引
- ✅ **监控API**: 实时性能统计
- ✅ **测试套件**: 189个测试用例可运行

## 🔗 相关文档

启动成功后，建议阅读以下文档了解更多：

- [项目架构文档](architecture/database-optimization-architecture.md)
- [API接口文档](api/api-documentation.md)
- [性能测试报告](testing/FINAL_TEST_REPORT.md)
- [部署指南](deployment-guide.md)

## 💡 下一步建议

1. **体验核心功能** - 添加一些投递记录，体验看板和统计功能
2. **测试性能优化** - 对比查询响应时间，体验优化效果
3. **探索监控功能** - 查看性能监控API，了解系统运行状态
4. **自定义配置** - 根据需要调整环境变量和数据库配置

现在你可以体验到经过全面优化的 JobView 系统了！🚀

如果遇到任何问题，请检查：
1. 所有服务是否正常运行
2. 环境变量配置是否正确  
3. 数据库连接是否成功
4. 端口是否被占用

---

**文档更新**: 2025年9月8日  
**适用版本**: JobView v2.0.0 (性能优化版本)  
**支持系统**: macOS, Linux, Windows