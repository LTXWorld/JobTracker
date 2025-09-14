# Chrome扩展程序配置指南

## 配置文件说明

扩展程序包含灵活的环境配置，支持本地开发和云服务器部署。

### 需要修改的文件

1. **config.js** - Background Service Worker使用的配置
2. **popup.js** - 弹窗界面使用的配置（已内嵌）

## 本地开发环境（默认配置）

当前默认配置支持本地开发：

```javascript
local: {
  API_BASE_URL: 'http://localhost:8010/api/v1',
  FRONTEND_URL: 'http://localhost:3000'
}
```

- 后端运行在: http://localhost:8010
- 前端运行在: http://localhost:3000

## 云服务器部署配置

当您部署到云服务器后，需要修改以下配置：

### 1. 修改 config.js

打开 `extension/config.js`，找到 `production` 配置部分：

```javascript
production: {
  API_BASE_URL: 'https://your-domain.com/api/v1',  // 替换为您的实际域名
  FRONTEND_URL: 'https://your-domain.com'
}
```

替换为您的实际域名，例如：

```javascript
production: {
  API_BASE_URL: 'https://jobview.example.com/api/v1',
  FRONTEND_URL: 'https://jobview.example.com'
}
```

### 2. 修改 popup.js

打开 `extension/popup.js`，找到第10-16行的 `ENV_CONFIG` 配置：

```javascript
production: {
  API_BASE_URL: 'https://your-domain.com/api/v1',  // 替换为您的实际域名
  FRONTEND_URL: 'https://your-domain.com'
}
```

同样替换为您的实际域名。

## 环境自动检测

扩展程序会自动检测运行环境：

- 如果访问的是 `localhost` 或 `127.0.0.1`，使用本地配置
- 其他情况使用生产环境配置

## 打包发布

### 开发版本

1. 直接加载 `extension` 文件夹作为未打包扩展程序

### 生产版本

1. 修改配置文件中的域名
2. 在 Chrome 扩展管理页面点击"打包扩展程序"
3. 选择 `extension` 文件夹
4. 生成 `.crx` 文件用于分发

## 测试检查清单

修改配置后，请测试以下功能：

- [ ] 扩展程序能正确连接到后端API
- [ ] 登录跳转到正确的前端地址
- [ ] 简历数据能正常同步
- [ ] 自动填充功能正常工作

## 常见问题

### Q: 如何支持多个环境？

可以在配置中添加更多环境，例如：

```javascript
const ENV_CONFIG = {
  local: { /* ... */ },
  staging: { /* ... */ },
  production: { /* ... */ }
};
```

然后修改 `detectEnvironment()` 函数的检测逻辑。

### Q: 如何处理HTTPS和HTTP混合？

如果您的服务器使用HTTPS，确保：
1. API_BASE_URL 使用 `https://`
2. 在 manifest.json 的 `host_permissions` 中添加对应的HTTPS域名

### Q: 配置修改后需要重新发布吗？

是的，配置修改后需要：
1. 重新加载扩展程序（开发模式）
2. 或重新打包发布（生产模式）

## 安全建议

1. **不要在代码中硬编码敏感信息**
2. **使用HTTPS保护数据传输**
3. **定期更新访问令牌**
4. **限制host_permissions到必要的域名**

## 技术支持

如遇到配置问题，请检查：
1. Chrome开发者工具的Console
2. 扩展程序的背景页面日志
3. 网络请求是否正确发送