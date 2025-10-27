# 5. Web 前端实现
## 5.1 框架与状态
`frontend/src/` 采用组合式 API + Pinia，存储切分为 `auth`、`jobApplication`、`statusTracking`、`mailEvent` 等。全局请求封装在 `api/request.ts`，统一拦截器支持网络探测、Token 续期、Blob 处理。

## 5.2 关键页面
- `views/KanbanBoard.vue`：拖拽列基于 `vuedraggable`，实时调用 `StatusTrackingAPI.updateStatus`。
- `views/Reminders.vue` 与 `components/ReminderManager.vue`：整合提醒与邮件事件，调用 `/v1/mail-events/*` 接口。
- `views/Resume.vue`：对接简历区块与附件上传，通过 `/static/` 地址渲染预览。
- `views/Statistics.vue`：消费 `/applications/statistics` 以及状态分析接口。

## 5.3 网络与身份
Auth Store 将 `access_token` 存在 `sessionStorage`，`refresh_token` 存在 `localStorage`，并定期验证 token；请求失败时依赖友好错误消息与重连提示。需要注意刷新逻辑：后端刷新接口为 `POST /api/auth/refresh`，前端基于 401 自动排队重试。

## 5.4 现存问题
- 前端默认分页逻辑假设 `/v1/applications` 返回 `{ data, has_next }`，若后端返回数组也能兼容，但建议统一为分页结构。
- 统计与提醒大量依赖本地推断字段（如 `reminder_category`），与后端字段命名需保持同步。
- `request.ts` 启动时即监听 `visibilitychange` 并定时检测网络，SSR/非浏览器环境需额外守护。
- 组件测试集中在 `frontend/tests/*`，覆盖率日志存在，但尚未覆盖新增状态流转与邮件面板。
