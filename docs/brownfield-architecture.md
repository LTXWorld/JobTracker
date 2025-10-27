# JobView Brownfield 架构文档

## 1. 现状概览
JobView 当前以单仓多模块形式交付，核心能力围绕求职投递管理、状态追踪、邮箱事件同步与多端体验扩展。在实际代码层面，后端由 Go 单体服务提供 REST 接口，前端使用 Vue 3 + Ant Design Vue，浏览器扩展通过 MV3 服务工作器与后端交互，React Native 客户端处于骨架阶段。文档体系较为完整，但部分架构稿件与代码实现已出现偏差，本文件以实际仓库 `main` 分支为准，描述当下可运行的 Brownfield 状态与主要技术债。

运行拓扑（开发模式）：
- `backend/cmd/main.go` 启动 Gorilla Mux 服务器（默认端口 `8010`），内建 JWT 鉴权、状态机、导出、邮箱同步等业务。
- `frontend/vite` 开发服务器监听 `3000`，通过代理将 `/api` 转发到后端。
- Chrome 扩展 (`extension/`) 直接调用后端公开 `https://jobview.bfsmlt.top/api/*` 或本地 `http://localhost:8010/api/*`。
- React Native 工程 (`mobile/JobViewMobile/`) 使用 Redux Toolkit 和 RTK Query，暂未与后端完全对齐。
- PostgreSQL 为唯一持久化存储（Docker Compose 与部署 Scripts 均针对 PostgreSQL 12+）。

## 2. 模块清单
| 模块 | 目录 | 技术栈 | 说明 |
| --- | --- | --- | --- |
| 后端 API | `backend/` | Go 1.24、Gorilla Mux、GORM、PostgreSQL、robfig/cron | 单体 REST 服务，内含数据库监控、导出、邮箱同步 |
| Web 前端 | `frontend/` | Vue 3、TypeScript、Vite、Pinia、Ant Design Vue、Vuedraggable | 管理后台、看板、时间线、提醒与简历维护 |
| 移动端 | `mobile/JobViewMobile/` | React Native 0.76、Redux Toolkit、Tamagui | 认证与列表骨架已建，API 协议尚未对齐 |
| 浏览器扩展 | `extension/` | Chrome MV3、原生 JS、Storage API | 自动填充简历、悬浮按钮、令牌续期 |
| 文档体系 | `docs/` | Markdown | 业务、部署、测试、旧架构文档，部分内容已与现状不符 |
| 运维脚本 | `scripts/`、`.github/workflows/` | Bash、GitHub Actions | 一键部署、状态修复、Baota 环境整合 |

## 3. 后端单体服务现状
### 3.1 启动与路由
`backend/cmd/main.go` 完成配置加载、数据库初始化、仓储服务装配以及路由注册。核心路由分为三类：
- `/api/auth/*`：注册、登录、刷新、资料、头像、可用性检查，注册登录接口挂载速率限制器。
- `/api/v1/*`：投递 CRUD、状态历史与撤销、状态配置、Excel 导出、简历、邮箱绑定、邮件事件、数据库监控。
- `/health`、`/static/*`：健康检查与静态上传目录映射，其中静态内容指向 `./uploads`。

### 3.2 中间件与安全
`backend/internal/auth/middleware.go` 提供日志、安全响应头、CORS、JWT 认证与内存速率限制 middleware。CORS 白名单写死常见本地端口与生产域名，并允许 `chrome-extension://*`，需要在发布前替换为具体扩展 ID。速率限制器采用进程内 `map[string][]time.Time`，对分布式部署或容器多副本无效。

### 3.3 领域服务
主要服务分布在 `backend/internal/service/`：
- `JobApplicationService`：调用 `repository.JobApplicationRepository` 完成投递 CRUD、批量操作、统计、搜索。仓储层大量使用 GORM `Raw` SQL，并依赖 JSON 字段如 `reminder_category`。
- `StatusTrackingService`：围绕 `job_status_history` 与 `status_version` 维护状态流转、撤销、面试体验；通过 `SetLocalFlag` 设置 Postgres session 变量配合触发器，允许在回退场景跳过自动写历史。
- `StatusConfigService`：确保默认模板存在直通规则，JSON 存储在 `status_flow_templates.flow_config`，方法 `EnsureDirectTransitionsInDefaultTemplate` 会幂等补齐基础规则。
- `ResumeService`：在缺失简历时自动创建，简历区块与附件分别存储在 `resume_sections`、`resume_attachments`，附件实际写入 `backend/uploads` 并通过 `/static/` 暴露。
- `ExportService`：根据数据量同步或异步导出 Excel，异步任务采用 goroutine + 本地临时目录，无外部队列，任务进度存放于 `export_tasks`。
- `EmailIntegrationService`、`EmailEventProcessor`、`EmailSyncManager`、`MailEventService`：完成邮箱授权、IMAP 拉取、事件解析、自动状态更新与人工确认。同步管理器 `EmailSyncManager` 基于 `robfig/cron` 在应用进程内调度，`Process` 会一次性加载用户全量投递记录匹配邮件，数据量大时存在性能风险。
- `MonitoringService`：封装 `database.QueryMonitor`，handlers 暴露 `/monitoring/db-stats` 等接口。

### 3.4 数据访问层
`backend/internal/database/db.go` 构造 `sql.DB` 与可选 `gorm.DB`，开启查询监控与健康检查。`database.RunMigrations` 在进程启动时保证表结构存在，并与 `backend/migrations/*.sql` 组合使用，注意存在重复编号文件（两个 `004_*`），但内容互不冲突。

`backend/internal/repository/*.go` 使用 `gorm.DB.Raw` 与事务组合实现：`status_tracking_repository.go` 暴露 `StatusTrackingTx`，对撤销、批量更新等场景执行 `BEGIN ... FOR UPDATE` 并插入历史；`email_repository.go` 与 `mail_event_repository.go` 则是纯 SQL 封装。

### 3.5 已知问题与风险
- 速率限制与邮箱调度均为进程内实现，不支持多副本水平扩展。
- 多个数据表字段（如 `job_applications.application_date`、`follow_up_date`）仍以 `VARCHAR` 存储，时间计算依赖应用层解析。
- `EmailEventProcessor` 在匹配事件时加载用户全部投递，随着数据增长容易触发 O(N) 扫描。
- `status_history` 与 `status_duration_stats` JSON 需要由业务层维护，缺少数据库约束防止结构漂移。
- `RunMigrations` 与 SQL 脚本共存，迁移顺序需要人工把控，`init.sql` 定义的 `application_status` 枚举未包含最新状态（如“简历筛选未通过”），若直接执行旧脚本会与运行时期望不一致。

## 4. 数据模型与存储
### 4.1 核心表概览
- `users`：`RunMigrations` 创建的基础表，存储用户名、邮箱、密码散列及时间戳。
- `job_applications`：投递主表，新增字段包括 `user_id`、提醒/面试信息、`company_attribute`、`status_history`、`status_version` 等；多处索引来自 `004_add_user_id_and_indexes.sql` 与 `005_add_advanced_optimization.sql`。
- `job_status_history`：记录状态流转与元数据，时长字段 `duration_minutes` 由服务层计算。
- `status_flow_templates`、`user_status_preferences`：使用 JSONB 存储流程定义与个人化配置。
- `export_tasks`：导出任务管理（参见 `backend/internal/repository/export_repository.go`）。
- `resumes`、`resume_sections`、`resume_attachments`：简历主表、区块与附件。
- `user_mailboxes`、`mail_events`、`interview_experiences`：邮箱授权、邮件解析事件与面试体验记录。`mail_events` 记录处理状态、置信度以及解析出的 JSON payload。

### 4.2 触发器与会话变量
状态跟踪依赖以下机制：
- `job_applications` 拥有触发器（见 `backend/migrations/006_add_status_tracking_system.sql`）用于维护 JSON 历史；当服务层设置 `SET LOCAL jobview.skip_history = 'on'` 时跳过自动写历史。
- 回退操作需要显式设置 `jobview.allow_backward`，否则触发器拒绝。
- 数据库视图 `user_data_summary`、`monthly_application_trend` 等为统计接口提供支撑。

### 4.3 技术债
- 字段类型不统一：`application_date` 存储格式不做约束，需要 API 层保证 `YYYY-MM-DD`。
- 迁移编号复用（双 `004`）不影响执行但增加维护成本。
- `init.sql` 与真实结构差异较大，建议仅用于参考，不应直接导入生产库。
- JSONB 字段缺少 schema 校验，版本演进需更新服务层兼容逻辑。

## 5. Web 前端实现
### 5.1 框架与状态
`frontend/src/` 采用组合式 API + Pinia，存储切分为 `auth`、`jobApplication`、`statusTracking`、`mailEvent` 等。全局请求封装在 `api/request.ts`，统一拦截器支持网络探测、Token 续期、Blob 处理。

### 5.2 关键页面
- `views/KanbanBoard.vue`：拖拽列基于 `vuedraggable`，实时调用 `StatusTrackingAPI.updateStatus`。
- `views/Reminders.vue` 与 `components/ReminderManager.vue`：整合提醒与邮件事件，调用 `/v1/mail-events/*` 接口。
- `views/Resume.vue`：对接简历区块与附件上传，通过 `/static/` 地址渲染预览。
- `views/Statistics.vue`：消费 `/applications/statistics` 以及状态分析接口。

### 5.3 网络与身份
Auth Store 将 `access_token` 存在 `sessionStorage`，`refresh_token` 存在 `localStorage`，并定期验证 token；请求失败时依赖友好错误消息与重连提示。需要注意刷新逻辑：后端刷新接口为 `POST /api/auth/refresh`，前端基于 401 自动排队重试。

### 5.4 现存问题
- 前端默认分页逻辑假设 `/v1/applications` 返回 `{ data, has_next }`，若后端返回数组也能兼容，但建议统一为分页结构。
- 统计与提醒大量依赖本地推断字段（如 `reminder_category`），与后端字段命名需保持同步。
- `request.ts` 启动时即监听 `visibilitychange` 并定时检测网络，SSR/非浏览器环境需额外守护。
- 组件测试集中在 `frontend/tests/*`，覆盖率日志存在，但尚未覆盖新增状态流转与邮件面板。

## 6. React Native 客户端
`mobile/JobViewMobile/` 已搭建 Redux、Persist、主题、网络状态等基础设施，但 API 期望与后端当前实现不匹配：
- 认证接口写死 `/api/v1/auth/*`，而后端实际路径为 `/api/auth/*`。
- CRUD 方法大量使用 `PATCH`、`DELETE`、批量端点 `/api/v1/applications/batch`，后端尚未提供对应 Handler。
- 本地存储键（`STORAGE_KEYS`）和 API 配置在 `src/config`，默认指向生产域名，开发环境需手动调整。
在将移动端投入使用前，需要同步接口契约、补足 RTK Query 定义与 UI 逻辑。

## 7. Chrome 扩展
`extension/` 基于 MV3，核心文件：
- `background.js`：封装 `JobViewAPI`，管理令牌刷新、数据缓存、简历抓取。请求使用 `AbortSignal.timeout`（需 Chromium 115+），针对 401 将请求加入等待队列。
- `content.js`：注入页面检测表单字段并调用后台填充。
- `config.js`：提供 `local` 与 `production` 两套基地址，通过 `detectEnvironment` 自动切换。
风险点：
- `host_permissions` 目前配置为 `https://*/*`，应缩小到支持站点。
- 缺少对后端证书错误与跨域失败的兜底提示。
- 缓存结构较大，需关注 `chrome.storage` 容量。

## 8. 运维与部署
### 8.1 环境变量
`.env.example` 与 `backend/internal/config/config.go` 定义核心变量：数据库连接、JWT、邮件加密密钥、轮询 Cron 表达式。生产环境必须提供 `JWT_SECRET`（>=32 字符）与安全的数据库密码。

### 8.2 本地运行
1. `docker-compose.yml` 启动 PostgreSQL 与后端（前端需单独启动或通过 Vite）。
2. 手动执行 `go run ./cmd/main.go` 前需准备 `.env` 或环境变量。
3. 前端执行 `npm install && npm run dev`，可通过 `VITE_API_BASE` 指向后端。
4. `run_tests.sh` 提供一键测试，默认执行前端 Vitest 与后端认证、Handler 测试，可通过 `--integration`、`--loadtest` 触发带标签测试。

### 8.3 CI/CD
`.github/workflows/deploy.yml` 使用 `appleboy/ssh-action` 在远端执行 `docker-compose.baota.yml`，并用容器化 Node 构建前端产物同步到宝塔站点目录。工作流依赖多个 Secrets（服务器地址、SSH Key、容器仓库凭据等）。另有 `scripts/deploy.sh`、`scripts/diagnose-status-flow.sh` 等用于生产修复。

## 9. 测试与质量
- 后端测试位于 `backend/tests/`，通过 build tag 区分 `integration`、`loadtest`，但大量用例仍为占位或日志输出，尚未覆盖状态机、邮箱同步等关键路径。
- 前端 Vitest 覆盖主要 Pinia Store 逻辑，组件与 API 的组合测试偏少。
- 移动端仅保留 React Native 默认快照测试。
- 文档 `docs/testing/backend-test-tags.md`、`backend/tests/TEST_STRATEGY.md` 对测试策略有规划，但落地程度有限。
- 研发需要结合 `verify_fixes.sh` 与 `docs/testing/*` 的脚本，在重要发布前手动执行集成与压力测试。

## 10. 技术债与风险优先级
**高优先级（必须处理）**
- 对齐 API 契约：移动端与部分前端接口仍假设 `/api/v1/auth`、`PATCH` 方法、批量端点，需要统一在 `backend` 或客户端修正。
- 校正数据库枚举与脚本：`init.sql` 过时，若继续对新环境执行将缺失状态枚举；应提供最新初始化脚本或弃用旧脚本。
- 邮件同步性能：当前实现对每位用户全表扫描，需引入分页或匹配索引避免数据量放大后阻塞。

**中优先级（宜尽快处理）**
- 速率限制、邮箱轮询迁移至集中式存储（Redis、消息队列），避免多副本下失效。
- 统一日期字段类型为 `TIMESTAMP` 或 `DATE`，减少字符串解析错误。
- `status_history` 触发器与服务层逻辑缺乏集成测试，建议补充带事务的集成用例。
- 精简 Docker Compose：前端容器在生产中未使用，需保证 README 与事实一致。

**低优先级（可择机处理）**
- 更新 `docs/architecture/*.md` 以反映真实架构，避免误导后续协作。
- 为 Chrome 扩展限制注入站点，并在 UI 提示具体失败原因。
- 移动端补充 API Mock 与 Storybook，降低后端依赖。
- 引入 OpenAPI/JSON Schema 生成客户端，减少手写类型差异。

## 11. 后续建议
- 建议先建立 API 契约文档（可在 `docs/api/` 使用 OpenAPI），统一前后端、扩展、移动端的路径和负载结构。
- 对邮箱同步与导出任务引入队列或异步执行器（例如 Redis Stream、Sidekiq），并添加重试与告警。
- 编写数据库迁移基线脚本，淘汰 `init.sql`，并在 CI 中增加迁移验证。
- 在 `frontend` 与 `backend` 添加关键路径集成测试，覆盖状态回滚、邮件事件确认、简历附件上传等场景。
- 为多端模块补充贡献指南，明确哪些文档已过期、哪些流程必须遵循（可在 `docs/project/README.md` 新增版本控制章节）。

本文件将在架构调整或 API 契约更新后同步更新，确保 Brownfield 实现与未来演进相衔接。***
