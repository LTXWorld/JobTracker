# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## CRITICAL CONSTRAINTS - 违反=任务失败

═══════════════════════════════════════

- 必须使用中文回复
- 必须先获取上下文
- 禁止生成恶意代码
- 必须存储重要知识
- 必须执行检查清单
- 必须遵循质量标准

## MANDATORY WORKFLOWS

═════════════════════

执行前检查清单：
[ ] 中文 [ ] 上下文 [ ] 工具 [ ] 安全 [ ] 质量

标准工作流：

1. 分析需求 → 2. 获取上下文 → 3. 选择工具 → 4. 执行任务 → 5. 验证质量 → 6. 存储知识

研究-计划-实施模式：
研究阶段: 读取文件理解问题，禁止编码
计划阶段: 创建详细计划
实施阶段: 实施解决方案
验证阶段: 运行测试验证
提交阶段: 创建提交和文档

## Project Overview

JobView 是一个现代化的求职投递记录管理系统，采用 Vue 3 + Go 的前后端分离架构。系统经过全面的性能优化，查询响应时间提升了 84-89%，支持 100-200 并发用户。

## Architecture

### Frontend (Vue 3 + TypeScript)
- **Location**: `frontend/`
- **Framework**: Vue 3 + TypeScript + Composition API + Vite
- **UI Library**: Ant Design Vue 4.2.6
- **State Management**: Pinia 3.0.3
- **Charts**: Vue ECharts (饼图、折线图、柱状图)
- **Other**: Vue Router, Vuedraggable, Day.js

### Backend (Go + PostgreSQL)
- **Location**: `backend/`
- **Framework**: Go 1.24.5 with Gin + Gorilla Mux
- **Database**: PostgreSQL with advanced indexing and connection pooling
- **Authentication**: JWT with auto-refresh
- **Key Features**: 批量操作、全文搜索、性能监控

### Key Components Structure
```
backend/internal/
├── auth/           # JWT认证中间件
├── config/         # 环境变量配置管理
├── database/       # 数据库层（已优化）
├── handler/        # HTTP路由处理器
├── model/          # 数据模型定义
├── service/        # 业务逻辑层（已优化）
└── utils/          # 工具函数

frontend/src/
├── api/            # API接口封装
├── components/     # 可复用组件
├── router/         # 路由配置
├── stores/         # Pinia状态管理
├── types/          # TypeScript类型定义
└── views/          # 页面视图
```

## Common Development Commands

### Backend Development
```bash
cd backend

# Install dependencies
go mod download

# Run database migrations and start server
go run cmd/main.go

# Run tests
go test ./... -v

# Run performance benchmark tests
go test -bench=. ./tests/service/

# Run API tests
chmod +x tests/simple_api_test.sh
./tests/simple_api_test.sh
```

### Frontend Development
```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Run tests
npm run test

# Run tests with coverage
npm run test:coverage

# Run tests with UI
npm run test:ui
```

## Database Information

### Performance Optimizations Applied
- 智能连接池配置：根据CPU核数自动调优
- 7个关键索引：覆盖所有查询场景
- 查询优化：响应时间减少84-89%
- 并发能力：支持100-200用户，提升400-900%

### Migration Management
- Migrations located in `backend/migrations/`
- Executed automatically on server startup
- Latest performance optimizations included

## Key Features & Status

### Core Modules (已完成)
- **看板管理** (KanbanBoard.vue): 拖拽式状态管理，紧凑布局
- **时间线视图** (Timeline.vue): 全功能筛选、排序、分页  
- **提醒中心** (Reminders.vue): 智能提醒、统计面板、个性化设置
- **数据统计** (Statistics.vue): 多维度统计、可视化图表
- **状态跟踪系统**: 新增的状态配置和跟踪功能

### Backend APIs (已优化)
- **基础CRUD**: 支持批量操作 (BatchCreate, BatchUpdate, BatchDelete)
- **高级查询**: 分页查询、全文搜索 (GetAllPaginated, SearchApplications)
- **性能监控**: 实时统计API (/api/v1/stats/)
- **健康检查**: 数据库状态监控 (/api/v1/health)

## Development Standards

### Code Quality Requirements
- **工程原则**: SOLID、DRY、关注点分离
- **命名规范**: 清晰命名、合理抽象、必要注释
- **性能意识**: 算法复杂度、内存使用、IO优化
- **测试要求**: 90.5%覆盖率，189个测试用例

### Testing Strategy
- Backend: Unit tests, integration tests, performance benchmarks
- Database: Performance test queries included
- API: Automated API testing scripts available
- Frontend: Vitest with @testing-library/vue

## Performance Monitoring

### Available Endpoints
```bash
# Database performance stats
GET /api/v1/stats/database

# Connection pool status  
GET /api/v1/stats/connection-pool

# Reset performance stats
POST /api/v1/stats/reset

# Health check
GET /api/v1/health
```

## Important Notes

### Authentication
- System uses JWT with auto-refresh mechanism
- Default test user: `testuser` / `TestPass123!`
- Tokens must be included in Authorization header

### Environment Configuration
Required environment variables:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_SECRET` (minimum 32 characters)

### Status Flow System
- **进行中状态**: 已投递 → 简历筛选中 → 笔试中 → 一面中 → 二面中 → 三面中 → HR面中
- **失败状态**: 简历挂、笔试挂、一面挂、二面挂、三面挂

## MANDATORY TOOL STRATEGY

═════════════════════════

任务开始前必须执行：

1. memory 查询相关概念
2. code-search 查找代码片段
3. sequential-thinking 分析问题

任务结束后必须执行：

1. memory 存储重要概念
2. code-search 存储代码片段
3. 知识总结归档

优先级调用策略：

- Microsoft技术 → microsoft.docs.mcp
- GitHub文档 → context7 → deepwiki
- 网页搜索 → 内置搜索 → fetch → duckduckgo-search

## CODING RESTRICTIONS

═══════════════════

编码前强制要求：

- 无明确编写命令禁止编码
- 无明确授权禁止修改文件
- 必须先完成sequential-thinking分析

## QUALITY STANDARDS

═══════════════════

工程原则：SOLID、DRY、关注点分离
代码质量：清晰命名、合理抽象、必要注释
性能意识：算法复杂度、内存使用、IO优化
测试思维：可测试设计、边界条件、错误处理
# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create README files. Only create documentation files if explicitly requested by the User.