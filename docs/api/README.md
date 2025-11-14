# 🔌 JobView API 文档

> RESTful API 接口完整参考

## 📋 目录

- [🚀 快速开始](#快速开始)
- [🔐 认证授权](#认证授权)
- [📊 投递记录API](#投递记录api)
- [🤖 银月助手API](#银月助手api)
- [📈 统计分析API](#统计分析api)
- [🔔 提醒功能API](#提醒功能api)
- [⚙️ 系统监控API](#系统监控api)
- [📱 移动端专用 API 契约](#移动端专用-api-契约)

## 🚀 快速开始

### 基础信息
- **基础URL**: `http://localhost:8010/api` (开发环境)
- **API版本**: v1
- **数据格式**: JSON
- **字符编码**: UTF-8

### 请求格式
```bash
curl -X GET \
  "http://localhost:8010/api/v1/applications" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

### 响应格式
```json
{
  "success": true,
  "data": {...},
  "message": "Success",
  "timestamp": "2025-01-21T10:00:00Z"
}
```

## 🔐 认证授权

### 用户注册
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "SecurePass123!"
}
```

### 用户登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "SecurePass123!"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com"
    }
  }
}
```

### Token刷新
```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

## 📊 投递记录API

### 获取投递列表（分页）
```http
GET /api/v1/applications/paginated?page=1&page_size=20&status=已投递
Authorization: Bearer YOUR_JWT_TOKEN
```

### 创建投递记录
```http
POST /api/v1/applications
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "company_name": "腾讯科技",
  "position_title": "前端开发工程师",
  "status": "已投递",
  "application_date": "2025-01-21",
  "salary_range": "20K-30K",
  "work_location": "深圳",
  "notes": "通过BOSS直聘投递"
}
```

### 更新投递记录
```http
PUT /api/v1/applications/123
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "status": "面试中",
  "interview_time": "2025-01-25T14:00:00Z",
  "notes": "一面-技术面试"
}
```

### 更新投递状态（携带面试体验）
```http
POST /api/v1/job-applications/123/status
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "status": "二面中",
  "metadata": {
    "interview_experience": {
      "rating": "good",
      "note": "一面沟通顺畅，反馈积极",
      "recorded_at": "2025-02-01T09:30:00Z"
    }
  }
}
```

> **说明**
> - `rating` 仅接受 `good`/`average`/`bad`，当 `skip=false`（默认值）时必填。
> - `note`、`skip_reason` 最长 200 字符，服务端会自动裁剪并清洗输入内容。
> - 若完全缺省 `interview_experience`，系统会为该次流转记录一条 `skip=true` 的默认反馈，确保历史数据完整。
> - 兼容前端老版本：未携带体验数据的请求仍遵循既有状态更新逻辑。
> - 自 2025-10-23 起，状态时间轴与状态详情弹窗会自动展示 `interview_experience` 的评分、备注或跳过原因，并在刷新历史后实时更新。

### 获取岗位面试体验记录
```http
GET /api/v1/applications/123/interview-experiences
Authorization: Bearer YOUR_JWT_TOKEN
```

**响应示例**
```json
{
  "code": 200,
  "message": "interview experiences retrieved successfully",
  "data": [
    {
      "id": 15,
      "application_id": 123,
      "from_status": "一面中",
      "to_status": "二面中",
      "rating": "good",
      "note": "反馈积极，准备深入业务题",
      "skip": false,
      "recorded_by": 42,
      "recorded_at": "2025-02-01T09:30:00Z",
      "created_at": "2025-02-01T09:30:00Z"
    }
  ]
}
```

### 批量操作
```http
POST /api/v1/applications/batch
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "operation": "update_status",
  "application_ids": [1, 2, 3],
  "data": {
    "status": "已投递"
  }
}
```

### 全文搜索
```http
GET /api/v1/applications/search?q=腾讯&page=1&page_size=10
Authorization: Bearer YOUR_JWT_TOKEN
```

### 删除投递记录
```http
DELETE /api/v1/applications/123
Authorization: Bearer YOUR_JWT_TOKEN
```

## 🤖 银月助手API

### 智能问答
```http
POST /api/v1/robot/chat
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "question": "如何使用看板功能？",
  "context": {
    "current_route": "/kanban",
    "user_context": {
      "total_applications": 50,
      "pending_interviews": 3
    }
  }
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "answer": "🎯 看板功能使用指南...",
    "source": "knowledge_base",
    "suggestions": [
      "如何设置提醒？",
      "投递数据统计"
    ]
  }
}
```

### 获取快速回复
```http
GET /api/v1/robot/quick-replies?context=kanban
Authorization: Bearer YOUR_JWT_TOKEN
```

### LLM服务状态检查
```http
GET /api/v1/robot/llm-status
Authorization: Bearer YOUR_JWT_TOKEN
```

## 📈 统计分析API

### 基础统计
```http
GET /api/v1/statistics/overview
Authorization: Bearer YOUR_JWT_TOKEN
```

**响应**:
```json
{
  "success": true,
  "data": {
    "total_applications": 120,
    "active_applications": 15,
    "success_rate": 0.25,
    "status_distribution": {
      "已投递": 45,
      "面试中": 15,
      "已收到offer": 8,
      "已拒绝": 52
    }
  }
}
```

### 趋势分析
```http
GET /api/v1/statistics/trends?period=month&start_date=2025-01-01
Authorization: Bearer YOUR_JWT_TOKEN
```

### 状态跟踪分析
```http
GET /api/v1/status-tracking/analytics?start_date=2025-01-01&end_date=2025-01-31
Authorization: Bearer YOUR_JWT_TOKEN
```

### 数据导出
```http
GET /api/v1/statistics/export?format=excel&date_range=last_month
Authorization: Bearer YOUR_JWT_TOKEN
```

## 🔔 提醒功能API

### 获取提醒列表
```http
GET /api/v1/reminders?type=interview&status=active
Authorization: Bearer YOUR_JWT_TOKEN
```

### 创建提醒
```http
POST /api/v1/reminders
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "application_id": 123,
  "type": "interview",
  "reminder_time": "2025-01-25T13:30:00Z",
  "message": "腾讯一面提醒"
}
```

### 标记提醒已读
```http
PUT /api/v1/reminders/456/dismiss
Authorization: Bearer YOUR_JWT_TOKEN
```

## ⚙️ 系统监控API

### 健康检查
```http
GET /api/v1/health
```

**响应**:
```json
{
  "status": "healthy",
  "database": {
    "status": "connected",
    "response_time_ms": 12
  },
  "services": {
    "llm_service": "available",
    "file_storage": "available"
  },
  "uptime_seconds": 3600
}
```

### 性能统计
```http
GET /api/v1/stats/database
Authorization: Bearer YOUR_JWT_TOKEN
```

### 连接池状态
```http
GET /api/v1/stats/connection-pool
Authorization: Bearer YOUR_JWT_TOKEN
```

## 📱 移动端专用 API 契约

Kotlin Compose Android 客户端的专用接口定义（折叠状态列表、移动统计、个人资料/简历等）已独立整理为 [mobile-kotlin-contract.md](./mobile-kotlin-contract.md)，主要包含：

- `GET /applications/mobile-overview` 及 `POST /applications/status/bulk`
- `GET /analytics/mobile` 的统计响应结构
- `GET/PUT /me/profile` 与 `GET/PUT /me/resumes`
- 通用约定与错误码表

若需扩展/实现移动端功能或同步 OpenAPI，请以该文件为准。

## 📝 错误处理

### 错误响应格式
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input parameters",
    "details": {
      "field": "email",
      "issue": "invalid format"
    }
  },
  "timestamp": "2025-01-21T10:00:00Z"
}
```

### 常见错误码
| 错误码 | HTTP状态 | 描述 |
|--------|----------|------|
| `AUTH_REQUIRED` | 401 | 需要登录认证 |
| `TOKEN_EXPIRED` | 401 | Token已过期 |
| `PERMISSION_DENIED` | 403 | 权限不足 |
| `RESOURCE_NOT_FOUND` | 404 | 资源不存在 |
| `VALIDATION_ERROR` | 400 | 输入验证失败 |
| `RATE_LIMIT_EXCEEDED` | 429 | 请求频率过高 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

## 🔧 请求限制

### 频率限制
- **认证接口**: 5次/分钟
- **查询接口**: 100次/分钟
- **写入接口**: 60次/分钟
- **批量接口**: 10次/分钟

### 数据限制
- **请求体大小**: 最大2MB
- **批量操作**: 最多100条记录
- **文件上传**: 最大10MB
- **查询结果**: 最多1000条记录

## 🧪 测试环境

### 测试账号
```
用户名: testuser
密码: TestPass123!
```

### 测试数据
API提供测试数据生成接口：
```http
POST /api/test/generate-data
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "applications_count": 50,
  "include_reminders": true
}
```

## 📚 SDK 和工具

### JavaScript SDK
```javascript
import { JobViewAPI } from '@jobview/sdk'

const api = new JobViewAPI({
  baseURL: 'http://localhost:8010/api',
  token: 'your-jwt-token'
})

// 获取投递列表
const applications = await api.applications.getList({
  page: 1,
  pageSize: 20
})
```

### Postman Collection
导入Postman集合文件：[JobView.postman_collection.json](./postman/JobView.postman_collection.json)

---

**🔌 让API集成变得简单高效！** ✨

> **需要帮助？** 查看具体的API示例代码或联系技术支持团队！
