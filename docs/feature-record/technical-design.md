# 记录投递插件 技术设计（Technical Design）

## 总览
- 形态：Chrome/Edge WebExtension（MV3），复用现有 extension 目录的架构（background.js、content.js、popup.js、config.js）。
- 能力：页面解析（Content Script）、字段确认与微表单（Popup/内嵌浮层）、调用平台 API 新建投递记录（Background）。
- 兼容：现有扩展已实现登录桥接、数据拉取与定期同步；本设计在此基础上新增“记录投递”的数据上行链路与站点解析适配器。

## 架构与模块
- Content Script（新增能力）
  - 站点检测：根据 `location.hostname`、`pathname` 确定是否进入解析流程。
  - 数据提取：
    - 优先 JSON-LD（`<script type="application/ld+json">`，schema.org JobPosting）。
    - 备选 OpenGraph/Meta（`og:title`、`og:site_name` 等）。
    - 兜底 DOM 选择器（按站点适配器配置或“训练模式”映射）。
  - UI：
    - 浮动按钮/贴片：提示“记录到平台”，点击打开确认弹窗（或调用 popup）。
    - 轻量微表单：展示解析结果（公司/职位/薪资/地点/链接/来源/投递日期/企业属性），用户可修改必填项。
  - 事件钩子：侦测“投递/申请/立即投递”按钮点击前后再次抓取，确保字段最新。

- Background（扩展后台）
  - 鉴权：沿用现有 token 获取与刷新逻辑（`JobViewAPI`）。
  - 统一提交：接收 Content Script 的标准化职位对象，补齐元信息并调用后端：`POST /api/v1/applications`。
  - 去重与重试：
    - 去重规则：同日同 URL 或（公司+职位）视为潜在重复，查询本地缓存或后端模糊查重接口（后续可加）。
    - 队列：IndexedDB/`chrome.storage` 排队；失败指数退避重试；离线恢复后自动补交。

- Popup（插件弹窗）
  - 登录状态与一键“连接平台”入口（已有）。
  - 插件设置：
    - 默认 `company_attribute`（例如“私企”）；
    - 站点开关、是否自动弹出确认对话框；
    - 隐私项控制（是否上传联系信息等）。
  - 手动触发：从 popup 点击“记录本页职位”。

## 数据模型与接口
- 标准化抓取模型（Content Script → Background）
  - `source_domain`：`location.hostname`
  - `job_url`：`location.href`
  - `company_name`：字符串
  - `position_title`：字符串
  - `salary_text`：原始薪资文本（后台转换为 `salary_range`）
  - `location_text`：原始地点文本（后台转换为 `work_location`）
  - `job_description`：文本（长度上限，如 5–8KB）
  - `captured_at`：时间戳

- 后端创建接口（已存在）
  - `POST /api/v1/applications`，Body 为 `CreateJobApplicationRequest`：
    - 必填：`company_name`、`position_title`、`company_attribute`
    - 可选映射：`application_date`（默认今天）、`salary_range`、`work_location`、`job_description`、`contact_info`、`notes`

- 后端可选增强（非阻塞，建议）：
  - 模糊查重接口（GET `/api/v1/applications/search?q=...` 已存在，可复用）以辅助重复确认。

## 站点适配器（Adapters）
- 目录结构建议（扩展内）：
  - `extension/adapters/{domain}.js`：导出 `extract(document) => JobPostingLike`。
  - `extension/adapters/index.js`：按域名路由到对应适配器，失败回落到 `genericExtractor`。
- 适配优先级：`jsonLdExtractor` → `metaExtractor` → `{domain}Adapter` → `genericDomExtractor`。
- 适配承诺：
  - 不注入站点全局变量，不破坏站点行为；
  - 选择器尽量靠近语义文本；
  - 失败返回最小对象并打点，便于回归。

## 关键逻辑细节
- 薪资解析：
  - 规则：匹配 `([\\d\\.]+)\\s*[-~—]\\s*([\\d\\.]+)\\s*(K|k|千|万)?\\s*(\\/月|\\/年)?`，统一到 `x-y k/月` 或 `面议`。
- 地点解析：
  - 识别省市区组合，过滤噪声（远程/多地时以分号拼接）。
- 企业属性：
  - MVP：来自用户设置的默认值或确认弹窗的下拉；后续考虑公司库或域名映射（如 `*.gov.*` 归类为“央国企”）。
- 去重策略：
  - 本地缓存最近 N 条（URL 哈希 + 公司 + 职位 + 日期）；
  - 命中后给出“覆盖/跳过/另存为新条”选项；
  - 视需要增加后端幂等键（提案：`client_hash` 字段，后端侧容忍重复）。

## 错误处理与可观测性
- 分类：解析错误/鉴权错误/API 错误/网络错误。
- 提示：
  - 解析错误：展示“快速添加”最小表单。
  - 鉴权错误：提示去平台登录，保留当前数据草稿。
  - API/网络错误：入队列，托盘提示“已加入稍后同步”。
- 日志：
  - 本地 Debug 日志仅在开发环境输出；
  - 匿名打点（后续）：适配器成功率、字段缺失率、创建成功/失败率。

## 安全与隐私
- 仅在受支持域名启用 Content Script 解析；
- 明确提示将要上传的字段，敏感字段默认关闭；
- Token 仅用于调用平台域名；
- 逐步引入本地加密（Crypto.subtle）存储敏感配置（后续）。

## 与现有代码的集成点
- `extension/content.js`：
  - 新增职位解析与“记录到平台”的 UI 流程；可复用现有浮动按钮基础设施。
- `extension/background.js`：
  - 新增 `CREATE_APPLICATION` 消息处理，封装 `POST /api/v1/applications`；
  - 新增简单去重与重试队列。
- `extension/config.js`：
  - 复用 `API_V1_BASE_URL`；无需改动。
- `manifest.json`：
  - 站点 `host_permissions` 已含常见域名；如有新增站点需同步更新。

## 性能与体验
- Content Script 延迟加载适配器（按需 import）；
- 解析避免全量 DOM 扫描，优先结构化数据；
- 提交请求并行去重校验，主线程不阻塞；
- 弹窗操作 ≤ 2 步完成提交。

## 测试计划（MVP）
- 单站点回归：各域名 5 条不同类型职位（校招/社招/薪资显示/不显示/多地点）。
- 边界：无结构化数据、SPA 路由切换、弱网/断网、未登录、CORS 校验。
- 验收脚本：
  - 解析正确率 ≥ 80%（公司/职位必填正确）。
  - 端到端提交成功率 ≥ 95%。

