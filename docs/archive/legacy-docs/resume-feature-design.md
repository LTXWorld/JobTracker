# JobView「我的简历」功能设计方案（Draft）

本文档描述在「看板视图」顶部标签栏新增「我的简历」功能的整体设计。目标是让用户在系统内维护结构化简历数据，作为后续一键投递、模板渲染、简历比对与求职分析等能力的基础。

## 1. 目标与范围

- 在看板视图顶部标签栏新增「我的简历」入口（位于现有标签的最后）。
- 提供结构化、分步骤的简历填写页（向导式/分区编辑），并支持自动保存与草稿恢复。
- 以结构化数据为核心，支持导出为 PDF/Docx（后续迭代），以及上传附件版 PDF 简历作为备用。
- 与后续功能衔接：一键投递、岗位匹配、模板渲染（中英双语）、多版本管理、AI 优化建议（后续迭代）。

不在本次范围：模板渲染、PDF 生成、AI 生成与 OCR 导入（留待后续）。

## 2. 典型简历字段与页面信息架构

参考主流中文招聘网站（如 BOSS、前程无忧、拉勾、智联）的字段归纳，采用「分区 + 可重复项」的结构：

- 基本信息（必填）
  - 头像（可选）、姓名、性别、出生年月、手机号、邮箱、所在城市/意向城市、学历最高层次、工作年限、求职状态
  - 社交与链接（GitHub、博客/作品集、LinkedIn、个人网站）
- 求职意向（必填）
  - 期望职位、期望地点、期望薪资区间、到岗时间、工作性质（全职/实习/兼职）
- 教育经历（可多条）
  - 学校、专业、学历、起止时间、GPA/排名（可选）、主修课程（可选）
- 工作/实习经历（可多条）
  - 公司/部门、职位、起止时间、职责描述（要点式）、业绩与量化指标（要点式）
- 项目经历（可多条）
  - 项目名称、起止时间、背景简介、职责/贡献（要点式）、技术栈、结果与量化指标
- 技能与证书
  - 技能条目（名称、熟练度、年限/评分、标签），证书（名称、颁发机构、时间、证书编号/链接）
- 奖项与荣誉（可多条）
- 自我评价 / 个人简介（文本、多行）
- 附件简历（PDF，备用/投递附件）

UI 交互建议：
- Tabs + 分步向导（Stepper）二选一（建议：左侧目录 + 右侧分区编辑，移动端折叠）。
- 区块内支持增删排序（如工作/项目经历），表单实时校验，字段级自动保存。
- 页头展示「完善度」进度条；缺失必填项在目录中以红点标识。

## 3. 数据模型设计

采用「结构化 + JSONB」的混合存储：
- 主表 `resumes`：用户与版本等通用元字段。
- 子表 `resume_sections`：每个分区一条记录，`content` 使用 JSONB 存储灵活结构。
- 附件表 `resume_attachments`：存储上传的 PDF 简历等文件元信息（路径保存在数据库，文件存磁盘）。
- 版本表（可选，后续）：`resume_histories` 记录每次发布版本的快照，支持回滚与对比。

### 3.1 表结构（PostgreSQL）

```
resumes (
  id              serial primary key,
  user_id         int not null references users(id) on delete cascade,
  title           varchar(100) default '默认简历',
  summary         text,                   -- 简要自我介绍，可做 SEO/推荐
  privacy         varchar(20) default 'private', -- private/shared/public
  current_version int default 1,
  is_completed    boolean default false,  -- 完成度阈值达成
  completeness    int default 0,          -- 0~100
  created_at      timestamptz default now(),
  updated_at      timestamptz default now()
);

resume_sections (
  id          serial primary key,
  resume_id   int not null references resumes(id) on delete cascade,
  type        varchar(30) not null,   -- base, intent, edu, exp, project, skill, cert, honor, summary, links
  sort_order  int default 0,
  content     jsonb not null,         -- 该区块具体字段 JSONB
  created_at  timestamptz default now(),
  updated_at  timestamptz default now()
);

resume_attachments (
  id          serial primary key,
  resume_id   int not null references resumes(id) on delete cascade,
  file_name   varchar(255) not null,
  file_path   varchar(500) not null,  -- ./uploads/resumes/{user}/{resume}/file.pdf
  mime_type   varchar(100),
  file_size   bigint,
  etag        varchar(64),
  created_at  timestamptz default now()
);

-- 可选：版本快照（后续迭代）
resume_histories (
  id           serial primary key,
  resume_id    int not null references resumes(id) on delete cascade,
  version      int not null,
  snapshot     jsonb not null,         -- 将所有 resume_sections 拼成一个结构化 JSON 快照
  created_at   timestamptz default now()
);
```

### 3.2 JSONB `content` 示例

- base（基本信息）
```json
{
  "name": "张三",
  "gender": "男",
  "birthday": "1999-08-01",
  "phone": "13800000000",
  "email": "me@example.com",
  "city": "上海",
  "years": 2,
  "degree": "本科",
  "avatar_url": "https://.../avatar.png",
  "links": [
    {"type": "github", "url": "https://github.com/xxx"},
    {"type": "blog",   "url": "https://blog.example.com"}
  ]
}
```
- exp（工作/实习）
```json
{
  "items": [
    {
      "company": "某互联网公司",
      "department": "后端平台",
      "position": "Golang 实习生",
      "from": "2023-06",
      "to": "2023-09",
      "highlights": [
        "参与微服务重构，拆分 3 个核心服务",
        "落地链路追踪，P99 时延下降 25%"
      ]
    }
  ]
}
```

## 4. API 设计（v1）

基于 REST，认证方式与现有保持一致（Bearer）。所有响应统一包一层 APIResponse。

- GET    `/api/v1/resumes/me`           获取当前用户的简历摘要（resumes + 统计完成度）
- POST   `/api/v1/resumes`              创建简历（首次进入自动创建）
- GET    `/api/v1/resumes/{id}`         获取简历及所有 sections（服务端聚合返回）
- PUT    `/api/v1/resumes/{id}`         更新简历元信息（title、privacy 等）
- DELETE `/api/v1/resumes/{id}`         软删除
- GET    `/api/v1/resumes/{id}/sections`        列出分区
- PUT    `/api/v1/resumes/{id}/sections/{type}` upsert 指定分区（content JSON）
- POST   `/api/v1/resumes/{id}/attachments`     上传附件（multipart/form-data: file）
- DELETE `/api/v1/resumes/{id}/attachments/{attId}` 删除附件

说明：
- 分区 upsert 采用 `type`（如 `exp`、`project`）作为幂等键；`content` 为 JSONB（后端做 schema 校验）。
- 服务端返回 `completeness`（0-100）与 `missing_required`（缺失字段列表），用于前端标识完善度。

## 5. 前端信息架构与交互

### 5.1 入口
- 在「看板视图」顶部标签栏新增 `我的简历`（最后一个 Tab）。
- Tab 内右侧操作区新增：
  - `保存全部`（可选，默认自动保存）
  - `预览`（后续打开模板渲染预览）
  - `导出`（后续）

### 5.2 页面结构
- 左侧目录（区块导航）：基本信息、求职意向、教育经历、工作/实习、项目、技能与证书、奖项荣誉、自我评价、附件简历
- 右侧编辑区：区块式卡片；可增删排序（经历/项目）；字段级校验；自动保存（debounce 600ms）。
- 顶部进度条：展示完善度，缺失必填项在目录处红点提示。

### 5.3 状态管理
- 新增 Pinia store：`useResumeStore`
  - state：`resume`, `sections`, `attachments`, `loading`, `saving`, `lastSavedAt`
  - actions：`fetchMyResume`, `fetchSections`, `upsertSection(type, content)`, `uploadAttachment(file)`, `deleteAttachment(id)`
  - autosave：节流 + 合并提交；失败提示与重试队列
  - 本地草稿缓存：localStorage（键含 user_id 与 section type），恢复时与服务端合并（以更新时间为准）

### 5.4 表单校验
- 前端基本校验：手机号/邮箱/URL/时间格式；必填项（姓名、手机号、邮箱、求职意向）
- 后端 schema 校验（按 `type` 做字段白名单与类型检查）

## 6. 安全、隐私与合规

- 认证：与现有 JWT 一致；所有简历接口均需登录态。
- 访问控制：简历默认为 private；后续支持 share/public 的 token 链接。
- 输入安全：后端对文本字段做 XSS 过滤（保留基础 Markdown/换行），对 URL 做白名单协议校验（http/https/mailto/tel）。
- 上传安全：限制类型（PDF）、大小（<= 5MB）、病毒扫描（可选预留）。
- 日志与审计：记录创建/更新/上传操作（user_id、IP、UA）。

## 7. 兼容性与迁移

- 新增表不影响现有业务；与 users/job_applications 无外键冲突（除 user_id 引用）。
- `uploads/resumes` 目录与现有 `uploads/avatars` 并行；静态路由共用 `/static` 前缀。
- 头像/附件本地路径统一返回绝对 URL，避免端口/域名切换引发显示问题。

## 8. 性能与可用性

- JSONB 适配灵活字段，结合必要索引（`resume_sections(resume_id, type)`）即可。
- 自动保存采用节流；失败进入重试队列；页面卸载前做最后一次 flush。
- 附件上传与图像处理使用原子写入；回滚逻辑保障并发安全。

## 9. 开发计划与任务拆分（建议）

- P0 基础能力（本次）
  1) DB 表：`resumes`、`resume_sections`、`resume_attachments`
  2) API：`GET/POST/PUT resumes`，`PUT sections/{type}`，`POST/DELETE attachments`
  3) 静态服务：`/static` 已有，新增 `uploads/resumes`
  4) 前端：新增 `我的简历` 页面与 Store；分区表单（基础信息/意向/教育/经历/项目/技能/证书/评价/附件）
  5) 自动保存与完善度逻辑；缺失提示
- P1 增强
  - 模板渲染预览（HTML/PDF）、多版本管理、导出 PDF/Docx、一键投递（与岗位表单联动）
- P2 AI 与推荐
  - 文案优化、JD 匹配、经验拆解、要点生成、关键字覆盖率分析

## 10. 接口草案（响应统一包裹 APIResponse）

- GET /api/v1/resumes/me
```json
{
  "code": 200,
  "message": "ok",
  "data": { "resume": {"id": 1, "title": "默认简历", "completeness": 72}, "sections": ["base","intent"] }
}
```
- PUT /api/v1/resumes/{id}/sections/base
```json
{
  "name": "张三",
  "phone": "13800000000",
  "email": "me@example.com",
  "city": "上海",
  "links": [{"type":"github","url":"https://github.com/xxx"}]
}
```
- POST /api/v1/resumes/{id}/attachments（multipart/form-data: file）
```json
{
  "code": 200,
  "message": "上传成功",
  "data": {"id": 10, "file_name": "cv.pdf", "url": "https://host/static/resumes/1/1/cv.pdf"}
}
```

## 11. 风险与对策

- 字段发散：采用 JSONB + 后端 schema 校验，前端表单与校验规则解耦。
- 自动保存覆盖：加 `updated_at`/`client_ts` 解决并发编辑（后续引入乐观锁）。
- 文件体积与存储：限制大小与数量、定期清理旧附件；保留最近 N 版本。

## 12. UI 低保真草图（文字描述）

- 看板标签：…… | 数据统计 | 我的简历（新）
- 我的简历页：
  - 顶部：标题 + 完成度进度条 + 预览/导出按钮
  - 左侧：区块目录（带完成度/缺失红点）
  - 右侧：表单卡片（区块式，教育/经历/项目为可重复列表，支持拖动排序）
  - 底部：自动保存状态（已保存于 12:30:15）

## 13. 结论

本方案以结构化数据为基础，兼顾灵活与约束，落地成本低，向后兼容强，并为后续模板渲染、导出、一键投递、AI 优化等功能提供可持续的扩展面。建议按 P0→P1→P2 的路线迭代实现。

