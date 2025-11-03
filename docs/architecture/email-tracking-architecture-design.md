# JobView 邮件追踪功能架构设计

## 1. 文档信息

| 项目 | 内容 |
|------|------|
| 文档版本 | v1.0 |
| 撰写日期 | 2025-02-19 |
| 作者 | 架构组（Codex Architect Agent） |
| 关联PRD | `docs/product/邮件追踪功能PRD.md` |
| 适用范围 | JobView 邮件追踪功能（后端/平台侧） |

---

## 2. 背景与范围

- PRD指出的业务目标：自动监听用户邮箱、智能解析招聘邮件、驱动求职状态更新并提供提醒与分析。
- 现状：JobView 后端为 Go 语言模块化单体，已有求职申请、通知、分析等基础能力，缺少与邮箱系统的深度集成及异步处理管线。
- 架构设计范围：后端服务、调度与解析链路、数据模型、与现有前端/求职域/通知域的接口，以及运维与安全方案。不覆盖前端组件细节（参考 `email-tracking-frontend-architecture.md`）。

---

## 3. 需求概述

### 3.1 功能需求映射（摘自 PRD 第3章）

| PRD模块 | 架构支撑能力 | 备注 |
|---------|--------------|------|
| 邮箱账户管理 | OAuth2 / 应用密码集成，凭据安全存储，连接健康检查 | 支持 Gmail、QQ、163、Outlook，提供多租户隔离 |
| 邮件同步服务 | 定时+事件驱动同步器，增量拉取，失败重试 | 支持 IMAP IDLE；POP3 仅可定时拉取 |
| 邮件智能解析 | 规则+模型双轨解析引擎，结构化抽取与置信度评价 | 解析结果可人工修订，沉淀规则配置 |
| 求职状态自动更新 | 事件总线驱动求职域更新，幂等保障 | 支持状态变更审批/人工确认策略 |
| 智能提醒系统 | 通知中心整合站内信/邮件/日历推送，时间触发 | 与现有提醒策略合并配置 |
| 数据统计分析 | 事件仓库+维度聚合，提供同期对比与准确率指标 | 与 BI/报表服务通过公共视图对接 |

### 3.2 非功能需求（摘自 PRD 第4章）

- 性能：邮件同步 T+5 分钟内完成，解析响应 <3 秒；并发 100+ 用户。
- 可用性：99.5%，同步失败需自动重试并具备降级能力。
- 安全：凭据 AES-256 加密，TLS 传输，操作审计，数据隔离。
- 合规：GDPR / PIPL，对用户授权、数据留存、删除提供机制。
- 可运维：可观测指标覆盖同步、解析、状态更新链路；支持容量扩展。

---

## 4. 架构目标与原则

1. **事件驱动的模块化单体**：保持现有 Go 单体形态，但引入清晰领域边界与内部事件总线，兼顾交付速度与可扩展。
2. **可配置 + 可演进**：解析规则、同步频率、提醒策略可配置；预留机器学习服务接入能力。
3. **安全优先**：最小权限访问邮箱，与 JobView 身份体系打通；凭据全程密文。
4. **鲁棒性**：网络抖动、第三方限流、解析失败等情境下保障数据一致与用户体验。
5. **可观测性闭环**：端到端追踪同步任务、解析事件、状态更新，便于交付 SLA。

---

## 5. 总体架构

### 5.1 架构形态

- **API & 后台服务**：Go + Gin，延续现有模块化单体结构，引入领域包（mailbox、sync、parser、tracking、notification、analytics）。
- **后台任务节点**：与主服务共代码库，通过独立进程（`cmd/sync-worker`）运行，复用内部包。
- **事件总线**：优先选用 Redis Stream + `asynq` 任务队列；提供可选 RabbitMQ 扩展方案，用于需要更强持久化/多语言订阅的场景。

### 5.2 高层架构图

```
┌────────────────────────────────────────────────────────────┐
│                       JobView 平台                         │
│                                                            │
│  ┌────────────┐     ┌──────────────────────────────────┐   │
│  │  Web / App │───► │ Email Tracking REST / GraphQL API│   │
│  └────────────┘     └───────▲───────────────┬──────────┘   │
│                              │               │              │
│                              │               │              │
│                        ┌─────┴─────┐   ┌─────┴─────┐        │
│                        │Sync Worker│   │Parser Svc │        │
│                        └─────┬─────┘   └─────┬─────┘        │
│                              │◄──────────────┘              │
│                              │                               │
│                      ┌───────▼────────┐                     │
│                      │ Redis Stream   │  事件/任务总线       │
│                      └───────▲────────┘                     │
│                              │                               │
│      ┌──────────────┐  ┌─────┴─────┐  ┌──────────────────┐ │
│      │ Postgres     │  │ Redis KV  │  │  JobView 核心域  │ │
│      │ (mail数据等) │  │ (缓存)    │  │ 求职/通知/分析等 │ │
│      └──────────────┘  └───────────┘  └──────────────────┘ │
│                                                            │
└────────────────────────────────────────────────────────────┘
         ▲
         │ IMAP/POP3/OAuth
┌────────┴────────┐
│  外部邮箱服务商  │
└─────────────────┘
```

---

## 6. 核心模块设计

### 6.1 邮箱账户管理（Mailbox Domain）
- **职责**：管理邮箱账户绑定、凭据存储、OAuth 授权更新、连接健康。
- **数据**：`mailbox_accounts`（凭据加密字段、同步策略）、`mailbox_tokens`（OAuth Refresh Token）、`mailbox_health_logs`。
- **接口**：
  - `POST /mailboxes`：创建并触发 OAuth。
  - `GET /mailboxes/{id}/health`：连接状态。
  - `PATCH /mailboxes/{id}/settings`：同步频率、通知策略。
- **安全**：凭据加密（KMS + AES-256-GCM），OAuth 回调使用短时临时凭证。

### 6.2 邮件同步调度（Sync Scheduler）
- **职责**：编排周期同步（Cron）、实时推送（IMAP IDLE）、手动触发。
- **实现**：`cron` + Redis 去重；每次触发生成 `mail.sync.job` 任务。
- **幂等**：任务参数包含邮箱ID + lastUID；同步偏移存于 Postgres。
- **回退策略**：指数退避（最大间隔 30min），超过阈值报警并标记邮箱为“需要关注”。

### 6.3 邮件同步执行（Sync Worker）
- **职责**：连接 IMAP/POP3，拉取原始邮件（含附件），标准化后写入暂存层。
- **流程**：
  1. 获取 OAuth Token / 应用密码。
  2. 读取 IMAP UID，过滤已处理邮件。
  3. 转换为统一结构（RFC822 → JSON），存入对象存储（MinIO/S3）与 `mail_raw_messages`。
  4. 发布 `mail.parsing.request` 事件。
- **容错**：连接失败、鉴权错误、邮箱限流 —— 记录 `mail_sync_logs`，触发告警。

### 6.4 邮件解析引擎（Parser Service）
- **职责**：分类、信息抽取、置信度评估。
- **架构**：管道式解析（Headers → 模板匹配 → 关键字规则 → NLP 模型）。
- **多方案对比**：
  | 方案 | 描述 | 优点 | 风险/成本 | 推荐 |
  |------|------|------|-----------|------|
  | 规则优先 + 轻量 ML (Go + `prose`/`regexp`) | 在 Go 内实现规则库，调用内置轻量模型 | 实现快、部署简单、与现有代码同库 | 复杂语句准确率有限，需要持续维护规则 | ✅ 首期采用 |
  | 独立 Python NLP 微服务 | FastAPI + spaCy/LLM，HTTP/gRPC 调用 | 高准确率，算法更新快 | 需额外部署与观察，跨语言通信 | ⬜ 第二阶段增强 |
  | 第三方 LLM API | 调用 OpenAI/国内模型 | 快速获得解析结果 | 成本高、隐私顾虑、合规风险 | ❌ 仅用于人工工具 |
- **输出**：`mail_events`（类型、公司、职位、时间、地点、置信度）。
- **人工校准**：支持将解析结果送入“待确认”队列，前端提示用户确认。

### 6.5 求职状态自动更新（Application Tracking）
- **职责**：根据解析事件匹配求职申请，执行状态迁移。
- **流程**：
  1. `mail.parsed` 事件触发状态匹配器。
  2. 按邮箱 → 公司 → 职位名称 → 自然语言模糊匹配求职记录。
  3. 若置信度 ≥ 阈值，生成 `application.status.change` 任务，幂等键 `application_id + mail_event_id`。
  4. 支持“自动更新”或“待确认”两种策略（依据用户设置与置信度）。
- **数据回写**：更新 JobView 求职申请表、写入 `mail_application_links`。

### 6.6 智能提醒与通知（Notification & Reminder）
- **职责**：统一提醒策略，避免重复通知。
- **实现**：
  - 解析结果中包含时间信息 → 投递到 `reminder_scheduler`（Redis Delayed Task）。
  - 即时通知：WebSocket / Server-Sent Events / 推送消息。
  - 日历集成：可生成 ICS 文件或调用第三方日历 API（第二阶段）。

### 6.7 数据分析与指标体系（Analytics）
- **职责**：统计邮件类型分布、解析准确率、状态更新效率。
- **技术**：Postgres 物化视图 + 定时刷新；可选对接 ClickHouse 进行事件分析。
- **指标**：解析命中率、状态自动更新率、平均响应时间、失败率。

---

## 7. 关键技术选型

| 组件 | 推荐技术 | 备选方案 | 选择理由 | 备注 |
|------|----------|----------|----------|------|
| API 框架 | Go + Gin + GORM | Go + Fiber / Echo | 与现有代码一致，团队熟悉 | 沿用现有模块 |
| 任务队列 | Redis Stream + `asynq` | RabbitMQ、Kafka | 轻量部署，支持延迟任务、幂等控制 | 若需多语言 / 高吞吐可切 RabbitMQ |
| 邮件协议库 | `github.com/emersion/go-imap` | `github.com/emersion/go-pop3` | 已在 go.mod 中，社区活跃 | POP3 仅异步拉取 |
| 对象存储 | MinIO（自建） | AWS S3 / 阿里云 OSS | 与现有部署方案匹配，可本地化 | 邮件原文加密存储 |
| NLP 工具 | 正则 + `github.com/jdkato/prose` | spaCy + FastAPI | 首期快速实现，可渐进增强 | 规则库需配置化 |
| 缓存 | Redis | Memcached | 支持 Stream、延迟任务，一体化 | |
| 日志/监控 | OpenTelemetry + Prometheus + Grafana | DataDog | 自建成本低，可复用现有观察系统 | 汇报 SLA |

---

## 8. 数据设计

### 8.1 核心表（示例字段）

```sql
CREATE TABLE mailbox_accounts (
  id               UUID PRIMARY KEY,
  user_id          UUID NOT NULL,
  provider         TEXT NOT NULL,           -- gmail / qq / outlook
  email_address    TEXT NOT NULL,
  auth_mode        TEXT NOT NULL,           -- oauth2 / app_password
  credential_blob  BYTEA NOT NULL,          -- AES-256-GCM
  sync_strategy    JSONB NOT NULL,          -- 频率、时间段
  status           TEXT NOT NULL DEFAULT 'active',
  last_health_at   TIMESTAMP,
  created_at       TIMESTAMP,
  updated_at       TIMESTAMP
);

CREATE TABLE mail_events (
  id                UUID PRIMARY KEY,
  mailbox_id        UUID NOT NULL,
  message_uid       BIGINT NOT NULL,
  raw_object_key    TEXT NOT NULL,
  detected_type     TEXT NOT NULL,          -- interview / offer / rejection ...
  confidence        NUMERIC(5,4) NOT NULL,
  metadata          JSONB,                  -- 公司、职位、时间等
  parsed_at         TIMESTAMP NOT NULL,
  confirmation      TEXT NOT NULL DEFAULT 'auto',
  created_at        TIMESTAMP
);

CREATE TABLE mail_application_links (
  id                UUID PRIMARY KEY,
  mail_event_id     UUID NOT NULL,
  application_id    UUID NOT NULL,
  status_before     TEXT,
  status_after      TEXT,
  resolution        TEXT NOT NULL,          -- auto_applied / user_confirmed / rejected
  resolved_at       TIMESTAMP,
  UNIQUE(mail_event_id, application_id)
);
```

### 8.2 数据流转

1. `mailbox_accounts` → 生成同步任务。
2. `mail_raw_messages`（对象存储）保存原始邮件，元数据写入 `mail_events`。
3. `mail_events` 与 `job_applications` 关联后写入 `mail_application_links`。
4. 指标视图：
   - `vw_mail_parse_accuracy`：按类型统计准确率。
   - `vw_mail_sync_latency`：同步耗时。
   - `vw_mail_reminder_effect`：提醒触达率。

### 8.3 数据生命周期

- 原始邮件保留 90 天（可配置），超期转冷存储或删除。
- 解析结果与状态关联长期保存，以支持历史追溯。
- 用户请求删除时，需级联清理邮箱凭据、原始邮件与解析数据。

---

## 9. 业务流程

### 9.1 邮件同步主流程

1. **任务触发**：Cron / 手动 / IMAP IDLE → `mail.sync.job(task_id, mailbox_id)`.
2. **同步执行**：Sync Worker 获取增量邮件，写入对象存储和 `mail_raw_messages`.
3. **事件发布**：成功后推送 `mail.parsing.request(mail_event_ids[])`.
4. **失败处理**：重试 + 失败计数，超过阈值推送 `mail.sync.failure` 告警。

### 9.2 邮件解析与状态更新

1. Parser Service 订阅 `mail.parsing.request`，逐条解析。
2. 解析结果（含置信度）写入 `mail_events`，并投递 `mail.parsed`.
3. 状态匹配器消费 `mail.parsed`，匹配求职申请：
   - 置信度 ≥ 0.8 → 自动更新并通知用户；
   - 0.5 ≤ 置信度 < 0.8 → 进入待确认队列；
   - < 0.5 → 仅记录，不触发变更。
4. 通知中心根据用户偏好推送消息；提醒调度器按时间安排事件。

### 9.3 人工确认闭环

1. 用户在前端确认/修订解析结果。
2. 前端调用 `PATCH /mail-events/{id}`，更新 metadata、确认状态。
3. 若用户矫正了状态，经 `mail.manual-confirmed` 重触发状态变更以保持一致。
4. 规则学习：将人工修订写入 `mail_parser_feedback`，供规则/模型训练。

---

## 10. 安全与合规设计

- **凭据保护**：使用 KMS 生成数据密钥，对邮箱凭据进行 AES-256-GCM 加密；密钥旋转周期 90 天。
- **访问边界**：邮箱同步和解析服务运行在隔离的 VPC 子网；对外通信仅走出口代理，限制 IP。
- **最小权限**：仅授予读取邮件所需的 OAuth Scope；POP3/IMAP 账号仅限读权限。
- **数据脱敏**：日志中脱敏邮箱地址、公司名称等敏感字段。
- **审计**：记录敏感操作（绑定邮箱、强制同步、人工确认）的操作人、时间、来源 IP。
- **合规**：提供用户授权记录、撤销流程；支持导出/删除邮件数据，满足 GDPR / PIPL 要求。

---

## 11. 可运维性与监控

- **指标**：
  - `mail_sync_latency_seconds`（同步时延直方图）
  - `mail_parse_duration_seconds`
  - `mail_parse_confidence_bucket`
  - `mail_state_update_success_total / failure_total`
  - `mail_reminder_delivery_success_total`
- **日志**：结构化 JSON，按任务 ID/邮箱 ID 打标签，便于追踪。
- **分布式追踪**：OpenTelemetry 贯穿 API → 队列 → Worker → 数据库。
- **告警**：
  - 同步失败率 > 10%（5 分钟内） → PagerDuty.
  - 邮件解析滞留队列长度 > 阈值 → 提示扩容。
  - 自动状态更新失败 → 降级到人工确认。
- **配置治理**：同步频率、解析策略存储于 `mail_configs` 表，并提供变更历史。

---

## 12. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 邮箱服务商限流 / 认证失败 | 邮件漏同步 | 缓存 Token、退避重试、切换备用凭据、告警人工干预 |
| 解析准确率不足 | 状态误判、用户不满 | 置信度阈值 + 人工确认闭环 + 规则快速迭代 |
| 数据量增长导致队列积压 | 同步延迟 | 任务分片（按邮箱划分）、Worker 自动扩容、队列监控 |
| 邮件格式多样复杂 | 解析失败 | 模板库与反馈系统持续迭代，引入 NLP 微服务路线 |
| 安全合规漏洞 | 法规风险 | 审核 OAuth Scope、定期安全测试、数据最小化 |

---

## 13. 实施路线建议

1. **Sprint 1-2**：完成邮箱账户管理、Cron 同步、基础解析（规则）、前端绑定流程。
2. **Sprint 3-4**：引入事件总线、自动状态更新闭环、提醒调度、监控指标。
3. **Sprint 5**：解析准确率提升（反馈机制、规则管理界面）、安全审计、性能调优。
4. **Sprint 6 以后**：评估 Python NLP 微服务或第三方模型、日历集成、跨邮箱协同分析。

---

## 14. 后续扩展思考

- **机器学习模型训练**：基于历史邮件与用户反馈训练分类模型，支持公司/职位实体识别。
- **团队协作**：支持团队共享邮箱或多人协同处理，同步读写冲突管理。
- **多渠道扩展**：对接招聘网站通知、短信等其他事件源，统一进入事件中心。
- **开箱配置模板**：针对常见招聘平台建立解析模板库，支持远程热更新。
- **合规审计**：提供企业版审计报表、敏感操作审批流。

---

## 15. 参考资料

- PRD：`docs/product/邮件追踪功能PRD.md`
- 前端架构：`docs/architecture/email-tracking-frontend-architecture.md`
- 平台整体架构：`docs/architecture/system-architecture.md`
- 数据库优化建议：`docs/architecture/database-optimization-architecture.md`

---

**状态**：草案，可根据评审反馈继续迭代。
