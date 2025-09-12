# JobView 头像存储与上传方案

本文档记录本次“头像上传 + 本地存储 + 前后端联动”的技术设计与实现细节，用于团队交流与后续维护。

## 目标

- 前端选择图片后立刻显示预览；上传成功后全局显示用户头像。
- 后端安全、简单地将头像保存到本地文件系统，并通过静态路由提供访问。
- 数据库存储头像“元信息”，便于缓存控制与统计，不保存二进制大对象。

## 架构与路径规范

- 存储介质：本地文件系统（后续可无缝迁移到 S3/OSS）。
- 目录结构：`./uploads/avatars/{user_id}/avatar_v{n}.{ext}`
  - 单用户独立目录；`v{n}` 表示版本号，便于缓存刷新与回滚。
- 对外访问 URL：`/static/avatars/{user_id}/avatar_v{n}.{ext}?v={n}`
  - 统一静态前缀 `/static` 指向 `./uploads`。
  - 通过 `?v=n` 控制缓存；也可叠加 ETag/Last-Modified。

## 数据库元信息

在 `users` 表添加以下字段（见代码中的迁移实现）：

- `avatar_path VARCHAR(255)`：相对路径，例如 `avatars/123/avatar_v3.jpg`
- `avatar_etag VARCHAR(64)`：内容 hash（预留，可选）
- `avatar_version INTEGER DEFAULT 0`：版本号
- `avatar_updated_at TIMESTAMP WITH TIME ZONE`：更新时间

选择“文件 + 元信息”的原因：

- 更轻更快：静态文件访问与备份分离，性能与维护性更优。
- 便于演进：未来迁往对象存储时，DB 字段保持兼容，仅切换存储实现。

## 后端实现

### 路由

- `POST /api/auth/avatar`（需认证）：上传头像（multipart/form-data，字段名 `avatar`）。
- 静态资源：`/static/*` → `./uploads`（已在 `cmd/main.go` 中挂载）。

### 处理流程

1. 认证：从上下文获取 `user_id`。
2. 解析上传表单：限制最大 5MB（前端 2MB 双重限制）。
3. 类型/大小校验：通过 `http.DetectContentType` 与大小判断，限制为 JPEG/PNG/WebP/GIF。
4. 计算新版本号：查询 `users.avatar_version`，`newVersion = old + 1`。
5. 构造保存路径并原子写入：写临时文件后 `rename` 覆盖，避免半成品。
6. 更新用户表：`avatar_path/avatar_version/avatar_updated_at/updated_at`。
7. 返回 `avatar_url`：`/static/{avatar_path}?v={version}`。

### 关键代码位置

- 迁移：`backend/internal/database/migrations.go`
  - 添加 `users` 表头像相关列（IF NOT EXISTS）。
- 模型：`backend/internal/model/user.go`
  - 在 `User`/`UserProfile` 中加入头像字段；`ToProfile()` 自动拼接 `/static/` 前缀与 `?v=`。
- 服务：`backend/internal/service/auth_service.go`
  - `UpdateAvatar(userID, file, header)`：类型校验、版本号自增、原子写入、DB 更新、返回 URL。
- 处理器：`backend/internal/handler/auth_handler.go`
  - `UploadAvatar`：解析 `multipart/form-data`，调用服务层，统一返回 JSON。
- 静态服务与路由：`backend/cmd/main.go`
  - `/static/*` → `./uploads`；新增 `POST /api/auth/avatar`。

## 前端实现

- 立即预览：选择图片后生成 `URL.createObjectURL(file)`，用户立即看到。
- 上传成功：使用服务端返回的 `avatar_url` 覆盖预览，并更新 `authStore.user.avatar`（持久化到 `localStorage`）。
- 头像展示：
  - 个人资料页与顶栏统一读取 `authStore.user?.avatar`。
  - 悬浮提示层：头像 hover 显示相机图标与“更换头像”。

关键代码：

- `frontend/src/views/auth/Profile.vue`
  - `beforeAvatarUpload`/`handleAvatarUpload`/`avatarPreview` 预览与上传。
- `frontend/src/components/AppLayout.vue`
  - 顶栏头像 `a-avatar` 渲染用户头像 URL。

## 安全与稳定性

- 类型校验：使用 `DetectContentType` 与扩展名双重校验。
- 大小限制：前端/后端双重限制（2MB/5MB）。
- 原子写入：避免并发上传导致半成品文件。
- 目录安全：使用固定 `uploads/avatars/{user}` 路径与服务端生成文件名，避免路径穿越。
- 缓存刷新：版本号自增 + `?v=version`，浏览器强缓存可配置 `Cache-Control`。

## 未来演进

- 对象存储：将物理存储替换为 S3/OSS，保留 DB 元信息与 `/static` 兼容层。
- 图片处理：接入图片压缩/裁剪/水印（可选依赖 `golang.org/x/image` 或第三方库）。
- 清理策略：保留最近 N 个版本（例如 2 个），上传成功后清理更老版本。可在服务层实现。

---

如需扩展为云存储版本，我可以保留相同接口，替换服务层 `UpdateAvatar` 为 S3/OSS 客户端实现，前端与上层代码不需要变化。

