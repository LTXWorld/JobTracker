# JobView Kotlin Compose 移动端实施方案

> 版本：v0.1（2025-02-15）  
> 适用范围：Android 平台（后续可扩展至 KMP/iOS）

## 1. 目标与范围

- **目标**：以 Kotlin + Jetpack Compose 重构移动客户端，优先落地求职进程、数据统计、添加投递、个人信息（含简历）四大模块，首阶段仅支持 Android。
- **范围**：
  - 复用现有 Go 后端 API，保持数据模型一致。
  - 设计可拓展的多模块 Kotlin 工程，后续可演进为 Kotlin Multiplatform。
  - 提供折叠式状态列表、统计图表、中心 FAB 表单、个人中心等核心交互。
  - 暂不实现简历附件上传与 iOS 版本。

## 2. 技术选型与架构

| 层级 | 技术/组件 | 说明 |
|------|-----------|------|
| UI | Jetpack Compose + Material 3 | 响应式 UI、支持 Dark Theme 与组件复用 |
| 状态管理 | MVVM + ViewModel + Kotlin Flow | 结合 UIState/UiEvent，确保可测试性 |
| 依赖注入 | Hilt（或 Koin） | 模块化依赖图，方便替换实现 |
| 网络层 | Retrofit + OkHttp + Kotlinx Serialization | 统一 APIResult、拦截器处理 JWT/错误 |
| 本地存储 | Room（列表缓存/草稿）、DataStore（配置与 Token）、EncryptedDataStore（敏感信息） | 支持离线浏览与草稿 |
| 后台任务 | WorkManager | 定时刷新、失败重试、未来附件续传 |
| 图片加载 | Coil | 头像、简历预览 |
| 测试 | JUnit5、MockK、Turbine（Flow）、Compose UI Test、MockWebServer | 单元/集成/UI 测试体系 |

### 多模块结构

```
app/
  - navigation/ 底部导航 + 路由
  - di/ 全局注入
core/
  model/ DTO、枚举、错误类型
  network/ Retrofit Service、拦截器、ApiClient
  database/ Room、DataStore、Entity 映射
  common-ui/ 主题、组件、图表包装
feature-progress/
feature-stats/
feature-add/
feature-profile/
testing/（可选测试共享模块）
```

## 3. 关键接口与数据契约

| 场景 | HTTP | 说明 |
|------|------|------|
| 获取状态概览 | `GET /api/v1/applications/mobile-overview` | 返回按状态分组的数据、分页信息、筛选配置 |
| 更新状态 | `POST /api/v1/applications/{id}/status` | 单条状态变更 |
| 批量状态更新 | `POST /api/v1/applications/status/bulk` | 长按多选后批量更新 |
| 获取统计 | `GET /api/v1/analytics/mobile?period=7d/30d/all` | KPI、漏斗、趋势、状态分布 |
| 创建投递 | `POST /api/v1/applications` | 请求体沿用 `CreateJobApplicationRequest` |
| 获取/更新个人资料 | `GET/PUT /api/v1/me/profile` | 头像、昵称、邮箱、偏好设置 |
| 简历管理 | `GET/PUT /api/v1/me/resumes` | 首版仅处理结构化字段，无文件上传 |

> 注：如接口尚未在后端实现，需要在 Swagger/OpenAPI 中补充定义，并确保 CORS/认证策略与 Web 一致。

### 添加投递字段映射（来自 `backend/internal/model/job_application.go:265`）

- 必填：`company_name`、`position_title`、`application_date`、`process_type`、`status`、`company_attribute`
- 可选：`job_description`、`salary_range`、`work_location`、`contact_info`、`notes`、`interview_time`、`reminder_time`、`reminder_enabled`、`reminder_category`、`follow_up_date`、`hr_name`、`hr_phone`、`hr_email`、`interview_location`、`interview_type`
- 表单建议按 4 个步骤分组：基础信息 → 职位与地点 → 联络提醒 → 面试备注；支持草稿保存与离线提交排队。

## 4. 功能设计要点

### 4.1 求职进程（折叠列表）
- `LazyColumn` 渲染状态卡，卡片内 `LazyList` 显示岗位。
- 默认展开当前/活跃状态，其余可折叠。
- 岗位卡支持：长按批量选择 → 底部操作栏（移动状态、删除、提醒）；右滑快速操作。
- 状态排序与展示偏好存 DataStore，Room 缓存最近数据，WorkManager 周期刷新。

### 4.2 数据统计
- 模块包含：关键指标（总投递/Offer 数等）、状态分布、阶段漏斗、时间趋势。
- 使用 Compose Charts 或封装 MPAndroidChart；Skeleton + 错误提示 + “数据更新时间”。
- 支持周期筛选（7/30/全部）与导出截图（后续迭代）。

### 4.3 添加投递（中央 FAB）
- BottomSheet 或全屏多步表单。
- 草稿：默认自动保存至 Room，退出提示是否保留。
- 提交后刷新求职进程列表并跳转至对应状态；失败展示错误提示并支持重试。
- 首版不包含附件上传；保留 WorkManager Hook 以便后续扩展。

### 4.4 个人信息 / 我的简历
- 个人资料：头像、昵称、邮箱、手机号、主题/语言偏好。
- 安全设置：改密、退出登录、JWT 刷新日志。
- 我的简历：展示现有分区，可编辑字段、设为默认，暂不含附件。
- 应用设置：通知开关（占位）、主题切换、调试工具入口。

## 5. 交互与体验

- 底部导航中间按钮使用凸起 FAB，Shadow + 品牌色，点击展开表单。
- 状态卡折叠过渡动画使用 `animateContentSize`，保证性能。
- 提供 `Empty`、`Error`、`Loading` 统一组件，弱网抛出离线模式提示。
- Snackbar 用于状态更新/添加成功的撤销提示，Undo 触发 API 回滚。

## 6. 开发迭代计划（估算 10~12 周）

| 阶段 | 内容 | 输出 |
|------|------|------|
| 1. 初始化 (1 周) | Gradle 多模块、CI（Detekt/Ktlint）、编码规范、主题/导航骨架 | 基础工程、lint、底部导航 Demo |
| 2. 基础能力 (1.5 周) | 认证流、网络/缓存、错误处理、通用组件 | 登录流程、统一 ApiResult、DataStore |
| 3. 求职进程 (2.5 周) | 折叠列表、筛选、详情、状态更新、Room 缓存 | MVP 功能 + 单元/UI 测试 |
| 4. 添加投递 (1.5 周) | FAB、表单、多步校验、草稿、提交回调 | 表单模块 + 草稿存储 |
| 5. 数据统计 (2 周) | 图表组件、聚合接口、周期切换、缓存 | 统计页 + 测试 |
| 6. 个人信息/简历 (2 周) | 资料编辑、简历分区 CRUD、设置 | Profile 模块、主题切换 |
| 7. 稳定性与发布 (1 周) | 弱网/离线策略、Crashlytics、灰度包 | Beta 包、监控接入 |

## 7. 风险与对策

| 风险 | 影响 | 对策 |
|------|------|------|
| 折叠列表与批量操作交互复杂 | 体验拉胯、实现耗时 | 优先交付折叠 + 底部操作栏，再评估横向看板 |
| 后端缺少移动专用聚合接口 | 数据冗余、性能差 | 同步补齐 `/mobile-overview`、`/analytics/mobile`，减少端侧计算 |
| JWT 与敏感信息安全 | 账号泄露 | Encrypted DataStore + 拦截器自动刷新，失败强制登出 |
| 图表性能 | 掉帧、耗电 | 使用懒加载、降采样、后台预计算，必要时自定义绘制 |
| 离线/弱网一致性 | 数据冲突 | 服务端以状态时间戳为准，端侧仅允许有网时提交；本地只读浏览 |

## 8. 下一步

1. 与后端确认/实现 `mobile-overview`、`analytics/mobile`、`me/profile` 接口，并更新 OpenAPI。
2. 产出底部导航、折叠列表、表单的低保真原型，锁定交互细节。
3. 搭建 Kotlin 多模块工程与 CI，完成认证和导航骨架。
4. 进入 Stage 3（求职进程模块）开发，随后按路线推进。

---
如需更新或扩展此方案，请在 PR 中同步修改该文档，并在 `docs/DOCUMENT_INDEX.md` 中登记。
