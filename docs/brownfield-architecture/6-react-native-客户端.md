# 6. React Native 客户端
`mobile/JobViewMobile/` 已搭建 Redux、Persist、主题、网络状态等基础设施，但 API 期望与后端当前实现不匹配：
- 认证接口写死 `/api/v1/auth/*`，而后端实际路径为 `/api/auth/*`。
- CRUD 方法大量使用 `PATCH`、`DELETE`、批量端点 `/api/v1/applications/batch`，后端尚未提供对应 Handler。
- 本地存储键（`STORAGE_KEYS`）和 API 配置在 `src/config`，默认指向生产域名，开发环境需手动调整。
在将移动端投入使用前，需要同步接口契约、补足 RTK Query 定义与 UI 逻辑。
