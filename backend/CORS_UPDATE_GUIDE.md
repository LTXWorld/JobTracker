# 后端CORS配置更新和部署指南

## 已更新的文件

1. **backend/cmd/main.go** - 添加了新的允许域名
2. **backend/internal/auth/middleware.go** - 改进了CORS中间件以支持Chrome扩展

## 更新的CORS配置

现在支持以下origin：
- `http://localhost:3000` - 本地前端开发
- `http://localhost:3001` - 备用前端端口
- `http://localhost:8010` - 本地后端
- `https://jobview.bfsmlt.top` - 您的生产环境
- `chrome-extension://*` - Chrome扩展

## 部署步骤

### 本地测试
```bash
cd /Users/lutao/GolandProjects/jobView/backend

# 重新编译
go build -o main cmd/main.go

# 运行测试
go test ./...

# 启动服务
./main
```

### 部署到云服务器

1. **提交代码到Git仓库**
```bash
git add .
git commit -m "更新CORS配置以支持Chrome扩展和生产环境"
git push origin main
```

2. **在服务器上更新代码**
```bash
ssh your-server
cd /path/to/jobView/backend
git pull origin main
```

3. **重新编译和重启服务**
```bash
# 编译
go build -o main cmd/main.go

# 停止旧服务
sudo systemctl stop jobview-backend

# 启动新服务
sudo systemctl start jobview-backend

# 或者如果使用Docker
docker-compose down
docker-compose up -d --build
```

## 验证部署

1. **检查服务状态**
```bash
sudo systemctl status jobview-backend
# 或
docker-compose ps
```

2. **测试CORS响应**
```bash
curl -H "Origin: https://jobview.bfsmlt.top" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: X-Requested-With" \
     -X OPTIONS \
     https://jobview.bfsmlt.top/api/v1/health -v
```

3. **测试Chrome扩展连接**
- 重新加载Chrome扩展
- 点击扩展图标
- 检查是否能正常连接到API

## 注意事项

1. **Chrome扩展的Origin**
   - Chrome扩展的origin格式是：`chrome-extension://[extension-id]`
   - 每个扩展都有唯一的ID
   - 在Chrome扩展管理页面可以看到扩展ID

2. **如果仍有CORS问题**
   - 检查nginx配置是否有额外的CORS设置
   - 确保后端服务正确重启
   - 查看后端日志确认新配置已生效

3. **临时解决方案**
   如果需要快速测试，可以临时修改CORS配置为允许所有域名：
   ```go
   router.Use(auth.CORSMiddleware([]string{"*"}))
   ```
   **警告：仅用于测试，生产环境请指定具体域名！**

## 扩展程序端的配置

确保扩展程序的配置文件已更新：
- `config.js` - API地址已设置为 `https://jobview.bfsmlt.top/api/v1`
- `popup.js` - 前端地址已设置为 `https://jobview.bfsmlt.top`
- `manifest.json` - host_permissions包含了您的域名