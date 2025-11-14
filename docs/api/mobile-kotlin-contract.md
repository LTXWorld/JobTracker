# JobView Kotlin 移动端 API 契约草案

> 版本：v0.1（2025-02-15）  
> 适用客户端：Android（Kotlin + Jetpack Compose）  
> 依赖前置：用户已通过 JWT 登录并持有 `Authorization: Bearer <token>`

## 1. 通用约定

| 项 | 约定 |
|----|------|
| Base URL | `https://{env}.jobview.com/api/v1`（开发/预发/生产） |
| 认证 | HTTP Header `Authorization: Bearer <access_token>` |
| 请求头 | `Content-Type: application/json`，所有时间统一使用 ISO8601（UTC） |
| 响应包裹 | `{"code":0,"message":"ok","data":{...}}`；错误时 `code != 0` 并附 `message` |
| 分页 | 默认 `page=1`、`page_size=20`、最大 `page_size=100`，字段与 `PaginationResponse` 对齐 |
| 状态码 | 200 成功、201 创建成功、400 参数错误、401 未认证、403 无权限、404 未找到、409 冲突、422 校验失败、500 内部错误 |

## 2. 接口总览

| 场景 | 方法 & 路径 | 说明 | 备注 |
|------|-------------|------|------|
| 求职进程 | `GET /applications/mobile-overview` | 返回按状态分组的折叠列表数据 | 新增 |
| 求职进程 | `POST /applications/{id}/status` | 单条状态更新（移动端调用入口） | 复用现有逻辑 |
| 求职进程 | `POST /applications/status/bulk` | 批量状态更新（长按多选） | 新增 |
| 添加投递 | `POST /applications` | 创建投递，沿用 `CreateJobApplicationRequest` | 复用 |
| 数据统计 | `GET /analytics/mobile` | 返回 KPI、漏斗、趋势、状态分布 | 新增 |
| 个人资料 | `GET /me/profile` / `PUT /me/profile` | 获取/更新用户资料与偏好 | 复用（增加偏好字段） |
| 我的简历 | `GET /me/resumes` / `PUT /me/resumes/{id}` | 读取/更新结构化简历数据 | 复用 |

以下章节给出详细契约。

---

## 3. 求职进程接口

### 3.1 `GET /applications/mobile-overview`

**用途**：折叠列表数据源，按状态分组返回卡片数据与分页信息。

**查询参数**
| 参数 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `status` | string[] | 全部 | 需要返回的状态集合，可多值 |
| `keyword` | string | 空 | 公司/职位模糊搜索 |
| `page` | int | 1 | 针对每个状态的分页页码 |
| `page_size` | int | 20 | 每个状态返回数量 |
| `updated_after` | string | 空 | 按更新时间增量同步 |

**响应示例**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "last_sync": "2025-02-15T08:40:00Z",
    "filters": {
      "status_order": ["Applied","PhoneScreen","FirstInterview","Offer"],
      "process_types": ["campus","social"],
      "company_attributes": ["央企","私企","外企"]
    },
    "groups": [
      {
        "status": "Applied",
        "display_name": "已投递",
        "total": 42,
        "page": 1,
        "page_size": 20,
        "has_next": true,
        "applications": [
          {
            "id": 1032,
            "company_name": "字节跳动",
            "position_title": "前端工程师",
            "application_date": "2025-02-10",
            "process_type": "social",
            "status": "Applied",
            "last_status_change": "2025-02-11T12:23:00Z",
            "work_location": "北京",
            "tags": ["急招","内推"],
            "reminder": {
              "enabled": true,
              "reminder_time": "2025-02-16T09:00:00Z",
              "category": "follow_up"
            }
          }
        ]
      }
    ]
  }
}
```

### 3.2 `POST /applications/{id}/status`

**请求体**
```json
{
  "status": "FirstInterview",
  "reason": "通过电话初筛进入面试",
  "note": "面试时间 2/20 上午",
  "origin": "mobile"   // 可选，便于审计
}
```

**响应**：返回更新后的投递实体。若状态校验失败返回 422。

### 3.3 `POST /applications/status/bulk`

**用途**：长按多选后批量移动状态。

**请求体**
```json
{
  "updates": [
    { "id": 1032, "status": "Offer", "note": "终面通过" },
    { "id": 1040, "status": "Rejected", "note": "对方未回复" }
  ]
}
```

**响应**
```json
{
  "code": 0,
  "message": "updated",
  "data": {
    "success": 2,
    "failed": [],
    "snapshot": {
      "Offer": 12,
      "Rejected": 5
    }
  }
}
```

冲突策略：若单条更新失败，`failed` 中附 `id` 与 `error`.

---

## 4. 添加投递接口

### 4.1 `POST /applications`

请求体沿用 `CreateJobApplicationRequest`（参考 `backend/internal/model/job_application.go:265`）。

**字段说明**
- `company_name` (string, required)
- `position_title` (string, required)
- `application_date` (string, yyyy-MM-dd)
- `process_type` (enum: `campus`/`social`/…)
- `status` (enum，默认 `Applied`)
- `company_attribute` (string, required)
- 其余字段如 `job_description`,`work_location`,`reminder_time`,`hr_name` 等均可选。

**响应**：返回新建投递记录，带初始状态历史。

---

## 5. 数据统计接口

### 5.1 `GET /analytics/mobile`

**查询参数**
| 参数 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `period` | enum(`7d`,`30d`,`90d`,`all`) | `30d` | 统计区间 |
| `process_type` | string | all | 求职周期筛选 |

**响应示例**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "generated_at": "2025-02-15T08:40:00Z",
    "kpi": {
      "total_applications": 120,
      "offers": 5,
      "interviews": 18,
      "offer_rate": 0.0417
    },
    "trend": [
      { "date": "2025-02-10", "count": 6, "status": "Applied" },
      { "date": "2025-02-11", "count": 3, "status": "Interview" }
    ],
    "funnel": [
      { "stage": "Applied", "value": 120 },
      { "stage": "PhoneScreen", "value": 40 },
      { "stage": "Offer", "value": 5 }
    ],
    "status_distribution": [
      { "status": "Applied", "value": 45 },
      { "status": "Interview", "value": 12 }
    ]
  }
}
```

---

## 6. 个人中心接口

### 6.1 `GET /me/profile`

**响应**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "email": "user@jobview.com",
    "nickname": "银月",
    "avatar_url": "https://cdn/jobview/avatar.png",
    "phone": "+86-138****0000",
    "preferences": {
      "default_status_order": ["Applied","PhoneScreen","Interview","Offer"],
      "theme": "system",
      "language": "zh-CN",
      "notification_enabled": true
    }
  }
}
```

### 6.2 `PUT /me/profile`

**请求体**
```json
{
  "nickname": "新的昵称",
  "avatar_url": "https://...",
  "phone": "+86-139****1111",
  "preferences": {
    "theme": "dark",
    "language": "en-US",
    "notification_enabled": false,
    "default_status_order": ["Applied","Interview","Offer","Rejected"]
  }
}
```

### 6.3 `GET /me/resumes`

返回当前用户所有结构化简历分区，示例：
```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 12,
      "name": "默认简历",
      "is_default": true,
      "sections": [
        {
          "type": "education",
          "items": [
            {
              "school": "清华大学",
              "degree": "硕士",
              "start": "2021-09",
              "end": "2023-07",
              "description": "计算机科学"
            }
          ]
        }
      ],
      "updated_at": "2025-02-10T05:20:00Z"
    }
  ]
}
```

### 6.4 `PUT /me/resumes/{id}`

> 首版仅支持结构化字段更新，不包含附件上传。

**请求体**
```json
{
  "name": "校园招聘版",
  "sections": [
    {
      "type": "experience",
      "items": [
        {
          "company": "JobView",
          "position": "后端工程师",
          "start": "2023-04",
          "end": "2024-12",
          "description": "负责状态跟踪系统"
        }
      ]
    }
  ],
  "is_default": true
}
```

---

## 7. 错误码示例

| code | HTTP | 说明 | 处理建议 |
|------|------|------|----------|
| 0 | 200 | 成功 | - |
| 10001 | 400 | 参数校验失败 | 检查必填字段、格式 |
| 10002 | 401 | Token 无效或过期 | 触发刷新或跳转登录 |
| 10003 | 403 | 无访问权限 | 提示用户联系管理员 |
| 20001 | 404 | 记录不存在 | 刷新列表或提示已被删除 |
| 20002 | 409 | 状态冲突或被他人更新 | 显示冲突提示并拉最新数据 |
| 30001 | 500 | 服务器内部错误 | 记录日志并提示稍后重试 |

---

## 8. 后续待办

1. 将上述接口补充进 OpenAPI/Swagger 文档，生成客户端 SDK。
2. 与后端确认新增字段（如 `preferences.default_status_order`）的存储方案。
3. 若需离线增量同步，可在 `mobile-overview` 基础上扩展 `last_sync_token`。
4. 结合 `docs/Android/kotlin-compose-migration.md` 的迭代计划，逐步实现并回填真实响应示例。
