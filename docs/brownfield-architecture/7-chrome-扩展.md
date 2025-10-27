# 7. Chrome 扩展
`extension/` 基于 MV3，核心文件：
- `background.js`：封装 `JobViewAPI`，管理令牌刷新、数据缓存、简历抓取。请求使用 `AbortSignal.timeout`（需 Chromium 115+），针对 401 将请求加入等待队列。
- `content.js`：注入页面检测表单字段并调用后台填充。
- `config.js`：提供 `local` 与 `production` 两套基地址，通过 `detectEnvironment` 自动切换。
风险点：
- `host_permissions` 目前配置为 `https://*/*`，应缩小到支持站点。
- 缺少对后端证书错误与跨域失败的兜底提示。
- 缓存结构较大，需关注 `chrome.storage` 容量。
