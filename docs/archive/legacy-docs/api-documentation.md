# JobView API 文档

> 最后更新：2024年9月14日
> API 版本：v1.0
> 基础路径：`/api`

## 📋 概述

JobView 提供 RESTful API 接口，支持 JSON 格式的请求和响应。所有 API 都遵循统一的设计规范和错误处理机制。

### 认证方式
- JWT Bearer Token 认证
- Token 有效期：24小时
- Refresh Token 有效期：30天

### 通用响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 错误响应格式
```json
{
  "code": 400,
  "message": "错误描述",
  "error": "详细错误信息"
}
```

### 状态码说明
- `200`: 成功
- `201`: 创建成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 无权限
- `404`: 资源不存在
- `500`: 服务器内部错误

## 🔐 认证 API

### 1. 用户注册
**POST** `/auth/register`

请求体：
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "Test@123456",
  "confirmPassword": "Test@123456"
}
```

响应：
```json
{
  "code": 201,
  "message": "注册成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "created_at": "2024-09-14T10:00:00Z"
    }
  }
}
```

### 2. 用户登录
**POST** `/auth/login`

请求体：
```json
{
  "username": "testuser",
  "password": "Test@123456",
  "remember": true
}
```

响应：
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "avatar": "/static/avatars/user1.jpg"
    }
  }
}
```

### 3. 刷新 Token
**POST** `/auth/refresh`

请求头：
```
Authorization: Bearer {refresh_token}
```

响应：
```json
{
  "code": 200,
  "message": "Token 刷新成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 86400
  }
}
```

### 4. 退出登录
**POST** `/auth/logout`

请求头：
```
Authorization: Bearer {access_token}
```

响应：
```json
{
  "code": 200,
  "message": "退出成功"
}
```

### 5. 获取当前用户
**GET** `/auth/current-user`

请求头：
```
Authorization: Bearer {access_token}
```

响应：
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "avatar": "/static/avatars/user1.jpg",
    "created_at": "2024-09-14T10:00:00Z"
  }
}
```

### 6. 更新个人资料
**PUT** `/auth/profile`

请求头：
```
Authorization: Bearer {access_token}
```

请求体：
```json
{
  "full_name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000"
}
```

### 7. 修改密码
**POST** `/auth/change-password`

请求体：
```json
{
  "old_password": "OldPass@123",
  "new_password": "NewPass@456",
  "confirm_password": "NewPass@456"
}
```

### 8. 上传头像
**POST** `/auth/avatar`

请求类型：`multipart/form-data`
- `avatar`: 图片文件（最大 2MB）

## 📝 投递记录 API

### 1. 获取投递列表（分页）
**GET** `/v1/applications`

查询参数：
- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 20）
- `status`: 状态筛选
- `company`: 公司名称筛选
- `sort_by`: 排序字段（date/company/status）
- `sort_order`: 排序方向（asc/desc）

响应：
```json
{
  "code": 200,
  "data": {
    "applications": [
      {
        "id": 1,
        "company_name": "阿里巴巴",
        "position_title": "前端工程师",
        "status": "一面中",
        "application_date": "2024-09-10",
        "salary_range": "20-30K",
        "work_location": "杭州",
        "company_attribute": "私企"
      }
    ],
    "pagination": {
      "current_page": 1,
      "page_size": 20,
      "total_count": 100,
      "total_pages": 5
    }
  }
}
```

### 2. 创建投递记录
**POST** `/v1/applications`

请求体：
```json
{
  "company_name": "字节跳动",
  "position_title": "后端工程师",
  "application_date": "2024-09-14",
  "status": "已投递",
  "company_attribute": "私企",
  "salary_range": "25-40K",
  "work_location": "北京",
  "job_description": "负责后端服务开发...",
  "hr_name": "李经理",
  "hr_phone": "13900139000",
  "hr_email": "hr@bytedance.com"
}
```

### 3. 获取单个投递详情
**GET** `/v1/applications/{id}`

响应包含完整的投递信息，包括状态历史。

### 4. 更新投递记录
**PUT** `/v1/applications/{id}`

请求体同创建，但所有字段可选。

### 5. 删除投递记录
**DELETE** `/v1/applications/{id}`

### 6. 批量创建
**POST** `/v1/applications/batch`

请求体：
```json
{
  "applications": [
    { /* 投递记录1 */ },
    { /* 投递记录2 */ }
  ]
}
```

### 7. 批量更新
**PUT** `/v1/applications/batch`

请求体：
```json
{
  "ids": [1, 2, 3],
  "updates": {
    "status": "简历筛选中"
  }
}
```

### 8. 批量删除
**DELETE** `/v1/applications/batch`

请求体：
```json
{
  "ids": [1, 2, 3]
}
```

### 9. 搜索投递
**GET** `/v1/applications/search`

查询参数：
- `q`: 搜索关键词
- `field`: 搜索字段（company/position/all）

## 📊 状态跟踪 API

### 1. 获取状态历史
**GET** `/v1/applications/{id}/status-history`

响应：
```json
{
  "code": 200,
  "data": {
    "history": [
      {
        "status": "已投递",
        "timestamp": "2024-09-10T10:00:00Z",
        "duration": 1440,
        "note": "通过官网投递"
      },
      {
        "status": "简历筛选中",
        "timestamp": "2024-09-11T10:00:00Z",
        "duration": 2880
      }
    ],
    "metadata": {
      "total_duration": 4320,
      "status_count": 2,
      "current_stage": "简历筛选中"
    }
  }
}
```

### 2. 更新状态
**PUT** `/v1/applications/{id}/status`

请求体：
```json
{
  "status": "一面中",
  "note": "电话面试",
  "interview_scheduled": "2024-09-15T14:00:00Z"
}
```

### 3. 获取状态统计
**GET** `/v1/stats/status-distribution`

响应：
```json
{
  "code": 200,
  "data": {
    "已投递": 15,
    "简历筛选中": 8,
    "一面中": 5,
    "二面中": 3,
    "已收到offer": 2,
    "已拒绝": 10
  }
}
```

## 📈 统计分析 API

### 1. 获取统计概览
**GET** `/v1/stats/overview`

响应：
```json
{
  "code": 200,
  "data": {
    "total_applications": 100,
    "active_applications": 30,
    "success_rate": 15.5,
    "average_process_time": 7.5,
    "status_distribution": {},
    "trends": {
      "period": "month",
      "data": []
    }
  }
}
```

### 2. 获取阶段分析
**GET** `/v1/stats/stage-analysis`

响应包含各面试阶段的通过率分析。

### 3. 获取公司统计
**GET** `/v1/stats/company`

返回按公司维度的投递统计。

## 🔔 提醒管理 API

### 1. 获取提醒列表
**GET** `/v1/reminders`

查询参数：
- `type`: 提醒类型（interview/follow_up）
- `status`: 状态（pending/sent）

### 2. 创建提醒
**POST** `/v1/reminders`

请求体：
```json
{
  "application_id": 1,
  "type": "interview",
  "reminder_time": "2024-09-15T09:00:00Z",
  "message": "下午2点面试提醒"
}
```

### 3. 标记已读
**PUT** `/v1/reminders/{id}/read`

### 4. 删除提醒
**DELETE** `/v1/reminders/{id}`

## 📄 导入导出 API

### 1. 导出 Excel
**POST** `/v1/export/excel`

请求体：
```json
{
  "format": "xlsx",
  "fields": ["company_name", "position_title", "status", "application_date"],
  "filters": {
    "status": ["一面中", "二面中"],
    "date_range": {
      "start": "2024-09-01",
      "end": "2024-09-30"
    }
  }
}
```

响应：
```json
{
  "code": 200,
  "data": {
    "task_id": "export_123456",
    "status": "processing"
  }
}
```

### 2. 获取导出任务状态
**GET** `/v1/export/tasks/{task_id}`

### 3. 下载导出文件
**GET** `/v1/export/download/{task_id}`

### 4. 导入 Excel
**POST** `/v1/import/excel`

请求类型：`multipart/form-data`
- `file`: Excel 文件
- `mapping`: 字段映射配置（可选）

## 🏥 系统健康 API

### 1. 健康检查
**GET** `/health`

响应：
```json
{
  "status": "healthy",
  "database": "connected",
  "uptime": 86400,
  "version": "2.0.0"
}
```

### 2. 数据库性能统计
**GET** `/v1/stats/database`

响应：
```json
{
  "code": 200,
  "data": {
    "query_count": 10000,
    "average_query_time": 50,
    "slow_queries": 5,
    "connection_pool": {
      "active": 10,
      "idle": 20,
      "max": 50
    }
  }
}
```

## 🔧 用户偏好设置 API

### 1. 获取用户偏好
**GET** `/v1/preferences`

### 2. 更新用户偏好
**PUT** `/v1/preferences`

请求体：
```json
{
  "notification_settings": {
    "email_enabled": true,
    "push_enabled": false,
    "reminder_frequency": "daily"
  },
  "display_preferences": {
    "show_durations": true,
    "timeline_compact": false,
    "kanban_show_counts": true
  }
}
```

## 📚 简历管理 API

### 1. 获取简历列表
**GET** `/v1/resumes`

### 2. 创建简历
**POST** `/v1/resumes`

### 3. 更新简历
**PUT** `/v1/resumes/{id}`

### 4. 删除简历
**DELETE** `/v1/resumes/{id}`

### 5. 导出简历为 PDF
**GET** `/v1/resumes/{id}/export/pdf`

## 🔑 辅助 API

### 1. 检查用户名可用性
**GET** `/auth/check-username`

查询参数：
- `username`: 要检查的用户名

### 2. 检查邮箱可用性
**GET** `/auth/check-email`

查询参数：
- `email`: 要检查的邮箱

### 3. 获取系统配置
**GET** `/v1/config`

返回前端需要的系统配置信息。

## 📖 使用示例

### JavaScript/TypeScript 示例
```typescript
// 登录
const response = await fetch('/api/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'testuser',
    password: 'Test@123456'
  })
});

const data = await response.json();
const token = data.data.access_token;

// 使用 Token 调用 API
const applications = await fetch('/api/v1/applications', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

### cURL 示例
```bash
# 登录
curl -X POST http://localhost:8010/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Test@123456"}'

# 获取投递列表
curl -X GET http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer {token}"
```

## 🚨 错误处理

### 常见错误码
- `400001`: 参数验证失败
- `401001`: Token 已过期
- `401002`: Token 无效
- `403001`: 无权限访问
- `404001`: 资源不存在
- `409001`: 资源冲突（如用户名已存在）
- `500001`: 数据库错误
- `500002`: 文件处理错误

### 错误响应示例
```json
{
  "code": 400001,
  "message": "参数验证失败",
  "error": "username: 用户名长度必须在3-20个字符之间"
}
```

## 🔄 版本管理

API 版本通过 URL 路径管理：
- 当前版本：`/api/v1`
- 认证 API 无版本前缀：`/api/auth`

## 📝 注意事项

1. **请求频率限制**：每个 IP 每分钟最多 100 次请求
2. **文件上传限制**：单个文件最大 10MB
3. **批量操作限制**：单次最多 100 条记录
4. **Token 自动刷新**：前端会在 Token 过期前自动刷新
5. **时区处理**：所有时间使用 UTC 格式，前端负责转换
6. **分页限制**：单页最大 100 条记录

---
*JobView API - 为开发者提供强大而简洁的接口*