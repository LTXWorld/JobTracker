# Repository Guidelines

## 项目结构与模块组织
仓库按端到端分层：`backend/` 承载 Go 服务，业务入口集中在 `cmd/`，领域逻辑位于 `internal/` 与 `pkg/`，数据库迁移保存在 `migrations/`，集成与负载测试放在 `tests/`，静态上传位于 `uploads/`。`frontend/` 是 Vue 3 + Vite 前端，核心模块放在 `src/`，组件测试集中至 `tests/` 和 `src/__tests__/`。`mobile/JobViewMobile/` 提供 React Native 客户端，页面代码位于 `src/`，Jest 用例位于 `__tests__/`。公共运维脚本置于 `scripts/`，根目录 `run_tests.sh` 串联关键测试流程，部署细节参考 `docs/deployment/` 与多环境 `docker-compose*.yml`。

## 构建、测试与开发命令
- `cd backend && go mod download && go run ./cmd/main`：启动本地 API（默认端口 `8010`）。
- `cd backend && go test ./...`：执行 Go 单元与集成测试，可通过 `TEST_DB_URL` 指定测试库。
- `cd frontend && npm install && npm run dev`：启动前端开发服务器，代理 `/api` 到后端。
- `cd frontend && npm run test:coverage`：运行 Vitest 并生成覆盖率报告。
- `cd mobile/JobViewMobile && npm install && npm run ios|android`：调试移动端，首次需执行 `npm run setup:scripts` 授权脚本。
- `./run_tests.sh`：一键串行校验前后端与移动端关键测试。

## 编码风格与命名约定
Go 代码提交前必须通过 `gofmt` 与 `goimports`，包名保持小写，结构体字段用驼峰并配套 JSON 标签。前端组件文件使用 PascalCase（示例 `JobCard.vue`），组合式函数以 `use` 前缀命名，样式建议遵循 BEM。Pinia store 以 `*.store.ts` 结尾，公共类型置于 `src/types/`。移动端启用 ESLint + Prettier，Hook 命名保持 `useXxx`，测试文件使用 `*.test.tsx`。

## 测试指南
后端采用标准 testing + Testify，新增 handler/service 需编写表驱动测试并维持覆盖率 ≥85%。前端使用 Vitest 与 Testing Library，组件测试放在同级 `__tests__/` 或 `tests/`，快照更新需说明差异。移动端使用 Jest，异步逻辑请结合 `@testing-library/react-native`。提交前分别运行 `go test ./...`、`npm run test:run`、`npm run test:coverage`（移动端可用 `npm run test:coverage`），并在 PR 中附测试摘要或日志路径。

## 提交与合并请求规范
提交信息保持简洁中文祈使句，例如 `优化岗位筛选缓存` 或 `[#123] 修复投递导出错误`；单次提交聚焦单一主题，必要时添加 `Co-authored-by` 注记。Pull Request 需：说明动机与核心改动、列出验证步骤及关键命令输出摘要、关联需求或缺陷链接、界面改动附前后对比截图或录屏、涉及配置时写明回滚方案与环境变量变更。

## 安全与配置提示
敏感信息通过环境变量管理，后端默认加载根目录或 `backend/.env`，生产环境务必设置长度 ≥32 的 `JWT_SECRET` 与独立数据库凭据。Docker 部署使用 `docker-compose.production.yml` 并在 `.env` 中覆盖端口、数据库和对象存储配置。`backend/uploads/` 仅供开发演示，生产场景请迁移至受控对象存储并开启访问控制。发布前复核 `CORS` 白名单、HTTPS 终端与日志脱敏策略，避免将密钥、测试数据或大型资产提交至仓库。
