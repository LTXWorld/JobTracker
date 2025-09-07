# JobView 后端安全改进测试指南

## 测试概述
本文档提供了对 JobView 后端安全改进功能的完整测试指南。包含功能测试、安全测试和性能测试的详细步骤和预期结果。

## 测试环境准备

### 1. 环境配置
```bash
# 克隆项目并进入目录
cd /Users/lutao/GolandProjects/jobView

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，设置实际的数据库密码和JWT密钥

# 启动数据库 (PostgreSQL)
# 确保 PostgreSQL 在 localhost:5433 运行

# 安装 Go 依赖
cd backend
go mod tidy
```

### 2. 必需依赖包
确保以下包已正确安装：
```bash
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt  
go get github.com/gorilla/mux
go get github.com/lib/pq
```

### 3. 启动服务
```bash
cd backend
go run cmd/main.go
```

预期输出：
```
2025/01/07 10:00:00 Running database migrations...
2025/01/07 10:00:00 Successfully connected to database  
2025/01/07 10:00:00 Created default test user: username=testuser, password=TestPass123!
2025/01/07 10:00:00 Database migrations completed successfully
2025/01/07 10:00:00 === JobView Backend Server Starting ===
2025/01/07 10:00:00 Environment: development
2025/01/07 10:00:00 Server starting on port 8010
2025/01/07 10:00:00 Health check: http://localhost:8010/health
2025/01/07 10:00:00 Auth endpoints: http://localhost:8010/api/auth/*
2025/01/07 10:00:00 API endpoints: http://localhost:8010/api/v1/*
2025/01/07 10:00:00 === Ready for connections ===
```

## 功能测试

### Test 1: 健康检查
```bash
curl -X GET http://localhost:8010/health
```

**预期结果**:
```json
{
  "code": 200,
  "message": "服务正常",
  "data": {
    "status": "ok",
    "service": "jobview-backend", 
    "version": "1.0.0",
    "timestamp": 1704614400,
    "environment": "development"
  }
}
```

### Test 2: 用户注册
```bash
curl -X POST http://localhost:8010/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "email": "test2@example.com",
    "password": "SecurePass123!"
  }'
```

**预期结果**:
```json
{
  "code": 201,
  "message": "注册成功",
  "data": {
    "user": {
      "id": 2,
      "username": "testuser2", 
      "email": "test2@example.com",
      "created_at": "2025-01-07T10:00:00Z",
      "updated_at": "2025-01-07T10:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1704700800
  }
}
```

### Test 3: 用户登录
```bash
curl -X POST http://localhost:8010/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "TestPass123!"
  }'
```

**预期结果**: 与注册类似的响应，包含用户信息和token

### Test 4: 访问受保护的API (获取用户信息)
```bash
# 使用从登录获得的token
TOKEN="your_jwt_token_here"

curl -X GET http://localhost:8010/api/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

**预期结果**:
```json
{
  "code": 200,
  "message": "获取用户信息成功",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com", 
    "created_at": "2025-01-07T10:00:00Z",
    "updated_at": "2025-01-07T10:00:00Z"
  }
}
```

### Test 5: 创建求职记录
```bash
curl -X POST http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Example Corp",
    "position_title": "Backend Developer",
    "application_date": "2025-01-07",
    "salary_range": "15-20万",
    "work_location": "北京"
  }'
```

**预期结果**:
```json
{
  "code": 201,
  "message": "job application created successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "company_name": "Example Corp",
    "position_title": "Backend Developer",
    "application_date": "2025-01-07",
    "status": "已投递",
    "salary_range": "15-20万",
    "work_location": "北京",
    "created_at": "2025-01-07T10:00:00Z",
    "updated_at": "2025-01-07T10:00:00Z"
  }
}
```

### Test 6: 获取用户的求职记录
```bash
curl -X GET http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN"
```

**预期结果**: 返回该用户创建的所有求职记录数组

### Test 7: 获取统计信息
```bash
curl -X GET http://localhost:8010/api/v1/applications/statistics \
  -H "Authorization: Bearer $TOKEN"
```

**预期结果**:
```json
{
  "code": 200,
  "message": "statistics retrieved successfully",
  "data": {
    "user_id": 1,
    "total_applications": 1,
    "in_progress": 1,
    "passed": 0,
    "failed": 0,
    "status_breakdown": {
      "已投递": 1
    },
    "pass_rate": "N/A"
  }
}
```

## 安全测试

### Test 8: 无认证访问受保护API (应该失败)
```bash
curl -X GET http://localhost:8010/api/v1/applications
```

**预期结果**:
```json
{
  "code": 401,
  "message": "Authorization header is required"
}
```

### Test 9: 无效Token访问 (应该失败)
```bash
curl -X GET http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer invalid_token"
```

**预期结果**:
```json
{
  "code": 401,
  "message": "Invalid token"
}
```

### Test 10: 弱密码注册 (应该失败)
```bash
curl -X POST http://localhost:8010/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "weakuser",
    "email": "weak@example.com",
    "password": "123456"
  }'
```

**预期结果**:
```json
{
  "code": 400,
  "message": "密码必须至少包含一个大写字母、至少包含一个小写字母、至少包含一个数字、至少包含一个特殊字符"
}
```

### Test 11: SQL注入测试 (应该失败)
```bash
curl -X POST http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Company; DROP TABLE users; --",
    "position_title": "Developer"
  }'
```

**预期结果**:
```json
{
  "code": 400,
  "message": "公司名称包含非法字符"
}
```

### Test 12: XSS攻击测试 (应该失败)
```bash
curl -X POST http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Normal Company",
    "position_title": "<script>alert(\"xss\")</script>Developer"
  }'
```

**预期结果**:
```json
{
  "code": 400,
  "message": "职位名称包含非法字符"
}
```

### Test 13: 数据权限隔离测试
```bash
# 1. 创建第二个用户并登录
curl -X POST http://localhost:8010/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user2",
    "email": "user2@example.com", 
    "password": "SecurePass123!"
  }'

# 保存user2的token
TOKEN2="user2_jwt_token_here"

# 2. user2创建求职记录  
curl -X POST http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN2" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Company B",
    "position_title": "Frontend Developer"
  }'

# 3. user1获取记录 (应该看不到user2的记录)
curl -X GET http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN"
```

**预期结果**: user1只能看到自己的记录，看不到user2创建的记录

### Test 14: Token刷新测试
```bash
# 使用refresh token获取新的访问token
curl -X POST http://localhost:8010/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your_refresh_token_here"
  }'
```

**预期结果**: 返回新的访问token和刷新token

## 性能和限流测试

### Test 15: 速率限制测试
```bash
# 快速连续发送多个认证请求 (应该被限流)
for i in {1..15}; do
  echo "Request $i:"
  curl -X POST http://localhost:8010/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "username": "testuser",
      "password": "TestPass123!"
    }'
  echo ""
done
```

**预期结果**: 前10个请求成功，之后的请求返回429 Too Many Requests

### Test 16: API限流测试
```bash
# 快速连续发送多个API请求
for i in {1..70}; do
  echo "API Request $i:"
  curl -s -X GET http://localhost:8010/api/v1/applications \
    -H "Authorization: Bearer $TOKEN"
done
```

**预期结果**: 前60个请求成功，之后的请求被限流

## 边界条件测试

### Test 17: 大文本输入测试
```bash
# 创建超长的备注信息
LONG_TEXT=$(python3 -c "print('A' * 3000)")

curl -X POST http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"company_name\": \"Test Company\",
    \"position_title\": \"Developer\",
    \"notes\": \"$LONG_TEXT\"
  }"
```

**预期结果**:
```json
{
  "code": 400,
  "message": "长度不能超过2000位"
}
```

### Test 18: 特殊字符处理测试
```bash
curl -X POST http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "测试公司™",
    "position_title": "软件工程师 (高级)",
    "work_location": "北京/上海"
  }'
```

**预期结果**: 应该成功创建，正确处理中文和特殊字符

## 数据完整性测试

### Test 19: 用户删除级联测试
```bash
# 注意：这个测试需要直接操作数据库
# 在数据库中执行：DELETE FROM users WHERE id = 2;
# 然后检查 job_applications 表中 user_id = 2 的记录是否也被删除

# 或者通过API创建用户和记录，然后检查级联删除
```

### Test 20: 数据库约束测试
```bash
# 尝试创建重复用户名
curl -X POST http://localhost:8010/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser", 
    "email": "different@example.com",
    "password": "SecurePass123!"
  }'
```

**预期结果**:
```json
{
  "code": 400,
  "message": "用户名已存在"
}
```

## 故障恢复测试

### Test 21: 数据库连接中断测试
```bash
# 1. 停止数据库服务
# 2. 尝试访问API
curl -X GET http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN"

# 3. 重启数据库
# 4. 再次尝试访问API
```

**预期结果**: 
- 数据库断开时返回500错误
- 数据库恢复后正常响应

### Test 22: 配置错误测试
```bash
# 1. 停止服务
# 2. 设置错误的JWT_SECRET (少于32字符)
export JWT_SECRET="short"

# 3. 尝试启动服务
go run cmd/main.go
```

**预期结果**: 服务启动失败，显示配置验证错误

## 测试结果验证

### 成功标准
- [ ] 所有功能测试通过
- [ ] 安全测试按预期阻止恶意访问
- [ ] 限流机制正常工作
- [ ] 数据权限隔离有效
- [ ] 输入验证和错误处理正确
- [ ] 配置验证和环境检查正常
- [ ] 性能在可接受范围内

### 测试报告模板
```
测试执行时间: 2025-01-07
测试环境: 开发环境
测试执行人: [姓名]

功能测试结果:
- Test 1-7: ✅ 通过 / ❌ 失败 (说明原因)

安全测试结果:  
- Test 8-14: ✅ 通过 / ❌ 失败 (说明原因)

性能测试结果:
- Test 15-16: ✅ 通过 / ❌ 失败 (说明原因)

边界条件测试:
- Test 17-18: ✅ 通过 / ❌ 失败 (说明原因)

总体评估: 通过/失败
需要修复的问题: [列出具体问题]
```

## 常见问题排查

### 1. 服务启动失败
- 检查数据库连接配置
- 确认环境变量设置正确
- 验证端口是否被占用

### 2. 认证失败
- 检查JWT_SECRET配置
- 确认token格式正确
- 验证token是否过期

### 3. 数据库连接问题
- 确认PostgreSQL服务运行
- 检查数据库用户权限
- 验证网络连接

### 4. API访问被拒绝
- 确认认证header格式: `Authorization: Bearer <token>`
- 检查用户权限
- 验证API路径正确

## 持续集成建议

### 1. 自动化测试脚本
创建包含所有测试用例的自动化脚本：
```bash
#!/bin/bash
# test-suite.sh

# 设置测试环境
source setup-test-env.sh

# 执行功能测试
echo "Running functional tests..."
./test-functional.sh

# 执行安全测试
echo "Running security tests..."
./test-security.sh

# 执行性能测试  
echo "Running performance tests..."
./test-performance.sh

# 生成测试报告
./generate-report.sh
```

### 2. 监控和告警
在生产环境中设置以下监控：
- API响应时间监控
- 错误率监控
- 认证失败率告警
- 数据库连接状态监控
- 内存和CPU使用率监控

---
**推荐测试工程师阅读完整测试指南后，按顺序执行所有测试用例，并记录详细的测试结果。任何失败的测试都应该立即报告给开发团队进行修复。**