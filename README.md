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
- **Web 前端**：使用 Vue 3 + Ant Design Vue 构建的管理后台，提供看板拖拽、时间线视图、提醒中心、数据统计、简历编辑、面试体验反馈等核心功能，同时内置银月智能助手（AI问答）与音乐播放器等体验增强模块。
- **Go 后端**：基于 Gin + Gorilla/Mux 的 RESTful API，支持 JWT 身份认证、状态跟踪系统、状态模板配置、Excel 异步导出、邮箱集成（IMAP授权、邮件同步、事件解析）、数据库性能监控等能力。经过深度优化，查询性能提升84-89%，支持100-200并发用户。
- **移动端**：React Native 0.76 客户端（基础架构完成），已完成项目搭建、导航系统、主题配置与Redux状态管理，正在开发认证与业务模块。
- **浏览器扩展**：Chrome 插件（MV3）用于在招聘网站表单中自动填充 JobView 的简历数据，支持智联、前程、BOSS等主流招聘网站，辅助用户快速投递。
- **文档中心**：`docs/` 下提供完整的架构设计、API文档、部署指南、用户手册与测试策略等全套文档，包含邮件追踪、状态跟踪等功能的详细设计文档。

## 核心功能

### ✅ 已实现功能

#### 投递记录管理
- **CRUD操作**：完整的创建、查看、编辑、删除投递记录功能
- **批量操作**：批量创建、批量更新状态、批量删除（支持最多100条记录）
- **高级搜索**：全文搜索支持公司名、职位名、工作地点多字段模糊匹配
- **分页查询**：高性能分页查询，支持自定义页面大小和排序
- **筛选功能**：多维度筛选（状态、日期范围、公司等）

#### 状态跟踪系统（v2.0+）
- **状态历史记录**：完整记录每次状态变更的历史轨迹
- **状态回退**：支持状态回退操作，保留历史记录
- **状态时间轴**：可视化展示状态变更时间轴
- **状态分析**：提供状态转化率、趋势分析、流程洞察等数据分析
- **面试体验反馈**：在状态变更时记录面试评分、备注和跳过原因

#### 状态模板与流程配置
- **自定义状态模板**：支持创建、编辑、删除状态流程模板
- **状态流转规则**：配置状态间的合规流转规则，支持直通规则
- **用户偏好设置**：个性化状态显示和流程配置
- **状态定义管理**：统一管理所有状态定义和可用转换

#### 看板与可视化
- **拖拽看板**：支持卡片在不同状态列间拖拽移动，实时更新状态
- **双模式显示**：进行中状态与失败状态分开展示
- **实时统计**：每个状态列的投递数量和汇总统计
- **搜索定位**：智能搜索公司/职位，支持高亮匹配和快速定位

#### 时间线视图
- **时间轴展示**：按时间顺序展示所有投递记录
- **多维度筛选**：支持公司、职位、状态、时间范围筛选
- **分页显示**：可配置的分页大小（10/20/50/100）
- **快速操作**：支持编辑、删除等快速操作

#### 提醒中心
- **提醒管理**：创建、编辑、删除跟进提醒和面试安排
- **提醒列表**：集中展示所有待处理提醒
- **提醒状态**：支持提醒的完成、延期等状态管理

#### 数据分析与统计
- **仪表盘统计**：投递总数、各状态分布、转化率等关键指标
- **阶段转化率**：各阶段间的转化率分析
- **趋势数据**：时间维度的投递趋势图表
- **Excel导出**：异步导出投递记录为Excel文件，支持自定义字段选择

#### 简历管理
- **简历CRUD**：创建、查看、编辑、删除简历
- **分区管理**：管理简历的各个分区内容（教育经历、工作经历等）
- **附件管理**：上传、查看、删除简历附件
- **自动填充**：Chrome扩展读取简历数据并自动填充到招聘网站表单

#### 邮箱集成（部分实现）
- **邮箱授权**：支持IMAP邮箱授权，加密保存凭证
- **邮箱管理**：绑定、查看、删除邮箱账户
- **邮件事件**：记录和查看邮件事件列表
- **邮件同步**：后端定时任务同步邮件（需配置 `MAIL_POLLING_ENABLED=true`）
- ⚠️ **邮件智能解析**：架构设计完成，待实现

#### 智能助手（v2.1+）
- **银月智能助手**：内置AI问答助手，支持三种模式
  - 内置知识库：快速回答求职相关问题
  - LLM大模型：集成本地Ollama或云端OpenAI API
  - 智能路由：根据问题类型自动选择回答模式

#### 体验增强
- **音乐播放器**：胶片唱盘设计，支持播放控制、进度条、音量调节
- **面试体验反馈**：在看板拖拽或快速更新时记录面试体验

#### 运维与监控
- **数据库监控**：实时查询性能统计、慢查询追踪
- **连接池监控**：数据库连接池状态和性能指标
- **健康检查**：系统健康检查端点（`/health`）
- **性能优化**：7个高性能索引，智能连接池配置，查询性能提升84-89%

### 🚧 开发中功能
- **移动端认证**：React Native移动端JWT认证流程
- **移动端业务模块**：投递记录CRUD、看板、统计等功能

### 📋 规划中功能
- **邮件智能解析**：自动识别邮件类型和关键信息
- **状态自动更新**：根据邮件内容自动更新求职状态
- **移动端离线功能**：WatermelonDB本地数据库集成
- **多用户/团队协作**：角色权限管理系统

## 技术栈
| 模块 | 技术 | 说明 |
|------|------|------|
| 后端 API | Go 1.24+, Gin, Gorilla/Mux, GORM, PostgreSQL 12+, JWT, Cron, Excelize, go-imap | 负责业务逻辑、状态跟踪、异步导出、邮箱同步、性能监控等功能 |
| 前端 Web | Vue 3, TypeScript, Vite, Ant Design Vue, Pinia, Vue Router, Vue ECharts, Vuedraggable | 管理界面、拖拽交互、图表渲染、AI助手集成与状态管理 |
| 移动端 | React Native 0.76, TypeScript, Redux Toolkit, Tamagui, React Navigation, WatermelonDB（规划中） | 多平台客户端，基础架构完成，正在开发业务功能 |
| 浏览器扩展 | Chrome Extension MV3, 原生 JS, Content Script, Storage API | 从 JobView 获取简历数据并自动填充招聘网站表单 |
| AI集成 | Ollama（本地）、OpenAI API（云端） | 银月智能助手支持本地和云端LLM模型 |
| 测试 | Vitest + Testing Library、Go testing + Testify、React Native Jest | 后端测试覆盖率90.5%，189个测试用例 |
| 监控 | 自定义性能监控、数据库健康检查、慢查询追踪 | 实时性能指标和系统监控 |

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

**性能优化**：系统已配置7个高性能索引和智能连接池，查询性能提升84-89%。可通过 `/api/v1/monitoring/db-stats` 查看数据库性能指标。

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
3. 登录 JobView Web 端后即可在指定招聘网站测试自动填充功能。
4. **支持的网站**：智联招聘、前程无忧、BOSS直聘等主流招聘网站。
5. 具体使用说明与调试技巧见 `extension/README.md` 和 `extension/CONFIGURATION.md`。

### 8. Docker Compose（可选）
- `docker-compose.yml`：本地一体化开发环境，包含后端、前端与 PostgreSQL 服务。
- `docker-compose.production.yml`：生产部署示例，配合外部配置覆盖敏感变量。
- `docker-compose.baota.yml`：宝塔环境部署模板。

启动示例：
```bash
docker compose up -d
```

## 环境变量与配置

### 后端配置
- 完整配置项位于 `backend/internal/config/config.go`，支持自动加载 `.env`。
- **核心环境变量**：
  ```bash
  DB_HOST=127.0.0.1
  DB_PORT=5432
  DB_USER=your_user
  DB_PASSWORD=your_password
  DB_NAME=jobview_db
  JWT_SECRET=your-32-chars-secret  # 必须≥32位
  MAIL_ENCRYPTION_KEY=jobview-mail-secret  # 邮箱凭证加密密钥
  MAIL_POLLING_ENABLED=false  # 本地开发默认关闭邮件同步
  ```
- **数据库性能配置**：系统会根据CPU核数自动优化连接池配置（生产环境：CPU核数×4，开发环境：CPU核数×2）。

### 前端配置
- 通过 `VITE_API_BASE` 控制 API 前缀，默认 `/api`。
- 可在 `frontend/.env.*` 中定制不同环境的配置。
- **AI助手配置**：
  - 本地Ollama：默认 `http://localhost:11434`
  - 云端OpenAI：需要配置 `VITE_OPENAI_API_KEY`

### 邮箱集成配置
- 邮箱集成需要 `MAIL_ENCRYPTION_KEY` 与 IMAP 服务信息。
- 启用邮件同步：设置 `MAIL_POLLING_ENABLED=true`。
- 详细配置说明见 `docs/deployment/mail-sync.md`。

### 文件存储
- 上传的简历附件默认保存在 `backend/uploads/`。
- 用户头像保存在 `backend/uploads/avatars/`。
- 生产环境建议迁移至对象存储（S3/OSS）并开启访问控制。

## 测试与质量保障
- **一键执行**：`./run_tests.sh`，支持 `--integration`、`--loadtest` 可选参数。
- **后端测试**：
  - 测试覆盖率：**90.5%**
  - 测试用例：**189个**
  - 执行命令：`cd backend && go test ./...`
  - 覆盖率报告：`go tool cover -html=coverage.out`
- **前端测试**：
  - 执行命令：`cd frontend && npm run test:run`
  - 覆盖率报告：生成于 `frontend/coverage/`
  - 使用 Vitest + Testing Library
- **移动端测试**：`cd mobile/JobViewMobile && npm run test`（Jest）
- **代码规范**：
  - Go：`gofmt`/`goimports` 格式化
  - 前端：ESLint + Prettier（`npm run lint`）
  - TypeScript：严格模式检查
- **性能测试**：
  - 数据库查询性能基准测试
  - 并发压力测试（支持100-200用户）
  - 慢查询监控（阈值100ms）
- **CI 建议**：结合 `verify_fixes.sh` 与 `docs/testing/` 中的测试策略，实现 PR 检查与自动化验证。

## 部署指引

### 快速部署
- **详细部署流程**：参考 `DEPLOYMENT_PROCESS.md`、`DEPLOYMENT_QUICK_START.md` 以及 `docs/deployment/`。
- **Docker Compose部署**：
  ```bash
  # 开发环境
  docker compose up -d
  
  # 生产环境
  docker compose -f docker-compose.production.yml up -d
  
  # 宝塔环境
  docker compose -f docker-compose.baota.yml up -d
  ```

### 生产环境配置建议
- **数据库**：
  - 独立 PostgreSQL 实例（推荐12+版本）
  - 启用定期备份与慢查询监控
  - 系统已优化7个高性能索引，查询性能提升84-89%
- **Web服务器**：
  - 配置 HTTPS 反向代理（Nginx/Caddy）
  - 将前端与后端统一为同域名并代理 `/api` 路径
  - 静态文件服务（头像、简历附件）
- **安全配置**：
  - 设置强随机 `JWT_SECRET`（≥32位）
  - 配置安全的数据库凭证
  - 邮箱同步密钥（`MAIL_ENCRYPTION_KEY`）
  - CORS白名单配置（生产域名）
- **性能监控**：
  - 数据库性能统计：`/api/v1/monitoring/db-stats`
  - 连接池状态：`/api/v1/monitoring/connection-stats`
  - 健康检查：`/health`
  - 慢查询追踪（阈值100ms）
- **邮件同步**：
  - 通过 `MAIL_POLLING_ENABLED` 控制是否启用
  - 调整同步频率（`MAIL_POLLING_INTERVAL`）
  - 生产环境建议使用外部任务调度（如Kubernetes CronJob）

### 部署检查清单
- [ ] 数据库迁移已执行（包含性能优化索引）
- [ ] 环境变量已正确配置
- [ ] HTTPS证书已配置
- [ ] 反向代理配置正确
- [ ] 静态文件服务路径正确
- [ ] 日志目录权限正确
- [ ] 监控端点可访问
- [ ] 备份策略已配置

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

### ✅ 已完成（v2.5.0）
- ✅ 数据库性能优化（查询性能提升84-89%，支持100-200并发用户）
- ✅ 状态跟踪系统（状态历史、回退、时间轴、分析）
- ✅ 批量操作（批量创建、更新、删除）
- ✅ 全文搜索与分页查询
- ✅ Excel异步导出功能
- ✅ 银月智能助手（AI问答）
- ✅ 音乐播放器
- ✅ 面试体验反馈功能
- ✅ 完整的文档体系重构
- ✅ 数据库监控与性能分析

### 🚧 进行中（v2.6.0）
- 🚧 移动端JWT认证与业务模块开发
- 🚧 邮件智能解析功能实现
- 🚧 状态自动更新功能（基于邮件解析）

### 📋 短期计划（v3.0.0 - 2025年Q2）
- 📋 移动端完整功能（投递记录CRUD、看板、统计）
- 📋 移动端离线数据同步（WatermelonDB集成）
- 📋 邮件事件驱动的状态自动更新
- 📋 更多招聘网站的表单适配（Chrome扩展）
- 📋 前端组件测试覆盖率提升

### 📋 中期计划（v3.5.0 - 2025年Q3）
- 📋 多用户/团队协作与角色权限管理
- 📋 更丰富的提醒方式（邮件、日历、推送通知）
- 📋 高级数据分析功能（投递成功率模型、智能建议）
- 📋 CI/CD流程集成与自动化测试

### 📋 长期计划（v4.0.0 - 2025年Q4）
- 📋 可插拔的智能分析模块
- 📋 开放API平台
- 📋 插件生态系统
- 📋 微服务架构评估（基于业务规模）
- 📋 容器化部署优化（Helm Chart）

## 社区与支持
- 问题反馈：GitHub Issues（暂未公开仓库可通过邮件联系）
- 讨论与建议：建议开启 GitHub Discussions 或内部协作工具
- 文档补充：欢迎提交 PR 至 `docs/`
- 邮件支持：`contact@jobview.com` / `docs@jobview.com`（可在实际部署中替换为团队邮箱）

## License
本项目尚未声明开源许可证，默认保留所有权利。若需在生产或商业场景中使用，请先与项目维护者确认授权策略。

---
> 🎯 欢迎一起完善 JobView，让求职管理更高效、更智能。
