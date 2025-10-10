# JobView —— 求职投递管理平台

![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat-square)
![Vue 3](https://img.shields.io/badge/Vue-3.x-42b883?style=flat-square)
![React Native](https://img.shields.io/badge/React%20Native-0.76-blue?style=flat-square)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-12%2B-336791?style=flat-square)

JobView 是一个帮助候选人管理求职投递全流程的多端应用，包含 Go 后端、Vue 3 Web 前端、React Native 移动端与 Chrome 浏览器扩展。项目聚焦在投递记录管理、流程追踪、提醒通知、数据分析与简历管理等场景，并提供丰富的运营与调试工具，适合作为个人求职管理工具或团队化的候选人跟进系统。

---

## 目录
- [项目简介](#项目简介)
- [核心功能](#核心功能)
- [技术栈](#技术栈)
- [目录结构](#目录结构)
- [快速开始](#快速开始)
- [环境变量与配置](#环境变量与配置)
- [测试与质量保障](#测试与质量保障)
- [部署指引](#部署指引)
- [文档与 API](#文档与-api)
- [贡献指南](#贡献指南)
- [路线图](#路线图)
- [社区与支持](#社区与支持)
- [License](#license)

---

## 项目简介
- **Web 前端**：使用 Vue 3 + Ant Design Vue 构建的管理后台，提供看板、时间线、提醒中心、数据统计、简历编辑等核心功能，同时内置知识型助手与音乐播放器等体验增强模块。
- **Go 后端**：基于 Gin + Gorilla/Mux 的 RESTful API，支持 JWT 身份认证、状态模板配置、Excel 导出、邮箱授权与数据库监控等能力。
- **移动端**：React Native 客户端（开发中），已完成底层架构、导航与主题配置，正在补齐认证与业务模块。
- **浏览器扩展**：Chrome 插件用于在招聘网站表单中自动填充 JobView 的简历数据，辅助用户快速投递。
- **文档中心**：`docs/` 下提供架构、API、部署、用户指南与测试等全套文档，适合作为团队协作基础资料。

## 核心功能
- **投递管理**：支持创建/编辑/删除投递记录、批量导入、关键字段搜索与高级筛选。
- **自定义流程**：通过状态模板实现动态列配置、合规的流转规则校验与批量状态更新。
- **时间线视图**：按阶段/日期聚合展示投递历史，内嵌进度追踪与提醒入口。
- **提醒中心**：集中管理跟进提醒、面试安排，支持邮箱授权后自动关联邮件事件。
- **数据分析**：提供仪表盘统计、阶段转化率、趋势数据与导出为 Excel 的报表能力。
- **简历管理**：在线管理简历元信息、分区内容与附件，Chrome 插件可读取并自动填充第三方表单。
- **邮箱集成**：支持 IMAP 邮箱授权、加密保存凭证、记录同步状态与邮件事件列表（需后端定时任务配合）。
- **运维监控**：数据库连接与性能指标、慢查询追踪及健康检查端点。
- **体验增强**：内置知识型助手（前端本地问答指引）和可选音乐播放器（开发环境使用 `frontend/public/music` 中样例音频）。

## 技术栈
| 模块 | 技术 | 说明 |
|------|------|------|
| 后端 API | Go 1.24, Gin, Gorilla/Mux, GORM, PostgreSQL, JWT, Cron, Excelize | 负责业务逻辑、状态机、导出、邮箱同步、监控等功能 |
| 前端 Web | Vue 3, TypeScript, Vite, Ant Design Vue, Pinia, Vue Router, Vue ECharts, Vuedraggable | 管理界面、交互体验、图表渲染与状态管理 |
| 移动端 | React Native 0.76, TypeScript, Redux Toolkit, Tamagui, React Navigation, WatermelonDB（规划中） | 多平台客户端，当前聚焦认证与基础界面 |
| 浏览器扩展 | Chrome Extension MV3, 原生 JS, Content Script, Storage API | 从 JobView 获取简历数据并自动填充招聘网站表单 |
| 测试 | Vitest + Testing Library、Go testing + Testify、React Native Jest、脚本化工具 | 覆盖前端、后端与移动端场景 |

## 目录结构
```text
backend/                 # Go 后端服务（cmd/internal/pkg/migrations/...）
frontend/                # Vue 前端应用（src/components/views/stores/...）
mobile/JobViewMobile/    # React Native 客户端（进行中）
extension/               # Chrome 插件源码
scripts/                 # 常用运维与工具脚本
migrations/              # 顶层数据库初始化脚本
docs/                    # 架构、API、部署、用户指南等文档
run_tests.sh             # 一键运行前后端测试脚本
verify_fixes.sh          # 回归校验辅助脚本
```

更多模块说明可参考 `docs/DOCUMENT_INDEX.md`。

## 快速开始
### 环境要求
- Node.js 18+
- Go 1.24+
- PostgreSQL 12+
- （可选）pnpm 或 npm 作为前端包管理器
- （可选）React Native CLI、Android Studio、Xcode（用于移动端开发）

### 1. 克隆仓库
```bash
git clone https://github.com/your-organization/jobView.git
cd jobView
```

### 2. 配置数据库
```bash
createdb jobview_db
psql jobview_db < init.sql   # 可选：导入初始数据
```

### 3. 设置环境变量
在项目根目录或 `backend/.env` 中配置核心变量：
```bash
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=jobview_db
JWT_SECRET=your-32-chars-secret
MAIL_ENCRYPTION_KEY=jobview-mail-secret
MAIL_POLLING_ENABLED=false   # 本地开发默认关闭自动同步
```
> 生产环境请使用安全的数据库凭证、HTTPS 与强随机的 `JWT_SECRET`，完整版变量说明见 [docs/deployment/env.md](docs/deployment/)（按需补充）。

### 4. 启动后端
```bash
cd backend
go mod download
go run ./cmd/main.go
```
默认监听 `http://localhost:8010`，初次启动会自动执行数据库迁移并创建所需索引。

### 5. 启动前端
```bash
cd frontend
npm install
npm run dev
```
前端开发服务器默认运行在 `http://localhost:3000`，通过 Vite 代理将 `/api` 请求转发到后端。

### 6. 启动移动端（可选）
```bash
cd mobile/JobViewMobile
npm install
npm run android   # 或 npm run ios（macOS）
```
当前移动端处于功能拓展阶段，主要用于验证导航、主题与认证流程，详细说明见 `mobile/JobViewMobile/README.md`。

### 7. 使用 Chrome 扩展（可选）
1. 打开 `chrome://extensions/` 并开启「开发者模式」。
2. 点击「加载已解压的扩展程序」，选择项目根目录下的 `extension/`。
3. 登录 JobView Web 端后即可在指定招聘网站测试自动填充功能，具体支持列表与调试技巧见 `extension/README.md`。

### 8. Docker Compose（可选）
- `docker-compose.yml`：本地一体化开发环境，包含后端、前端与 PostgreSQL 服务。
- `docker-compose.production.yml`：生产部署示例，配合外部配置覆盖敏感变量。
- `docker-compose.baota.yml`：宝塔环境部署模板。

启动示例：
```bash
docker compose up -d
```

## 环境变量与配置
- 后端完整配置项位于 `backend/internal/config/config.go`，支持自动加载 `.env`。
- 邮箱集成需要 `MAIL_ENCRYPTION_KEY` 与 IMAP 服务信息，详见 `docs/deployment/mail-sync.md`（如需启用定时拉取）。
- 前端通过 `VITE_API_BASE` 控制 API 前缀，默认 `/api`；可在 `frontend/.env.*` 中定制。
- 上传的简历附件默认保存在 `backend/uploads/`，生产环境建议迁移至对象存储并开启访问控制。

## 测试与质量保障
- **一键执行**：`./run_tests.sh`，支持 `--integration`、`--loadtest` 可选参数。
- **后端**：`cd backend && go test ./...`；覆盖率报告可通过 `go tool cover` 生成。
- **前端**：`cd frontend && npm run test:run`，覆盖率报告生成于 `frontend/coverage/`。
- **移动端**：`cd mobile/JobViewMobile && npm run test`（Jest）。
- **代码规范**：Go 代码需通过 `gofmt`/`goimports`，前端采用 ESLint + Prettier（`npm run lint`）。
- **CI 建议**：结合 `verify_fixes.sh` 与 `docs/testing/` 中的测试策略，实现 PR 检查与自动化验证。

## 部署指引
- 详细部署流程：参考 `DEPLOYMENT_PROCESS.md`、`DEPLOYMENT_QUICK_START.md` 以及 `docs/deployment/`。
- 建议的生产配置：
  - 独立 PostgreSQL 实例，启用定期备份与慢查询监控。
  - 配置 HTTPS 反向代理，将前端与后端统一为同域名并代理 `/api` 路径。
  - 设置 `JWT_SECRET`（≥32位）、数据库凭证、邮箱同步密钥等敏感参数。
  - 通过 `MAIL_POLLING_*` 调整同步频率，必要时关闭或改用外部任务调度。
  - 日志与告警可接入现有 APM/监控体系，慢查询统计接口见 `/api/v1/monitoring/db-stats`。

## 文档与 API
- 文档中心入口：`docs/README.md`，按照角色（产品、开发、运维、测试）分类导航。
- API 参考：`docs/api/` 提供 REST 接口说明及示例请求。
- 架构设计：`docs/architecture/` 描述系统组件、数据流与状态机设计。
- 用户指南：`docs/user-guide/` 覆盖看板、提醒、数据统计、简历与桌面助手等功能。
- 移动端与扩展补充文档分别位于 `mobile/JobViewMobile/docs/` 与 `extension/CONFIGURATION.md`。

## 贡献指南
1. Fork 仓库并创建 feature 分支。
2. 完成开发与测试，确保 `./run_tests.sh` 通过。
3. 遵循项目代码风格（Go 小写包名、前端 PascalCase 组件、Pinia store 以 `*.store.ts` 结尾等）。
4. 提交信息使用简洁中文祈使句，例如：`新增邮箱授权接口`。
5. 在 Pull Request 中：
   - 描述动机与核心改动
   - 附上验证步骤或关键命令输出
   - 涉及 UI 的提供截图/录屏
   - 涉及配置变更时说明回滚方案与环境变量调整
6. 等待代码审查与合并，必要时补充文档或测试。

更多贡献细节可参考 `docs/project/CONTRIBUTING.md`（若缺失请根据仓库规范补充）。

## 路线图
- **短期**
  - 移动端补齐 JWT 认证、投递记录 CRUD 与离线缓存
  - 完善邮箱事件自动解析与状态同步调度
  - 增强统计视图的图表配置与自定义导出
  - 扩展自动化测试覆盖率（前端组件、后端集成）
- **中期**
  - 支持多用户/团队协作与角色权限
  - 加入更丰富的提醒方式（如邮件、日历）
  - 完成插件对更多招聘网站的表单适配
  - 推动 run_tests 脚本融入 CI/CD 流程
- **长期**
  - 打通移动端与 Web 端数据同步，支持离线场景
  - 引入可插拔的智能分析模块（例如投递成功率模型）
  - 提供可配置的部署模板与 Helm Chart

## 社区与支持
- 问题反馈：GitHub Issues（暂未公开仓库可通过邮件联系）
- 讨论与建议：建议开启 GitHub Discussions 或内部协作工具
- 文档补充：欢迎提交 PR 至 `docs/`
- 邮件支持：`contact@jobview.com` / `docs@jobview.com`（可在实际部署中替换为团队邮箱）

## License
本项目尚未声明开源许可证，默认保留所有权利。若需在生产或商业场景中使用，请先与项目维护者确认授权策略。

---
> 🎯 欢迎一起完善 JobView，让求职管理更高效、更智能。
