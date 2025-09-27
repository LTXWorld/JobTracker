# 🛣️ 路径配置指南

> 确保本地开发和生产环境路径配置的一致性

## 📋 配置原则

### 🎯 核心原则
1. **统一API前缀**: 所有环境使用 `/api` 作为API路径前缀
2. **环境适配**: 通过环境变量区分本地和生产环境
3. **相对路径**: 前端使用相对路径，避免硬编码域名
4. **代理机制**: 开发环境使用Vite代理，生产环境使用Nginx反向代理

## 🔧 前端路径配置

### 1. 环境变量配置

#### 开发环境 (.env.development)
```bash
# frontend/.env.development
VITE_API_BASE=/api
NODE_ENV=development
```

#### 生产环境 (.env.production)
```bash
# frontend/.env.production
VITE_API_BASE=/api
NODE_ENV=production
```

### 2. Vite 配置 (vite.config.ts)
```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    // 关键配置：开发环境代理
    proxy: {
      '/api': {
        target: 'http://localhost:8010',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/api')
      }
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'static',
    // 生产环境路径配置
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia'],
          ui: ['ant-design-vue']
        }
      }
    }
  }
})
```

### 3. API 基础配置 (api/request.ts)
```typescript
import axios from 'axios'

// 关键配置：使用环境变量
const baseURL = import.meta.env.VITE_API_BASE || '/api'

const request = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

export default request
```

### 4. 路由配置 (router/index.ts)
```typescript
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  // 使用HTML5 History模式
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/dashboard'
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('@/views/Dashboard.vue')
    },
    {
      path: '/kanban',
      name: 'Kanban',
      component: () => import('@/views/KanbanBoard.vue')
    },
    {
      path: '/timeline',
      name: 'Timeline',
      component: () => import('@/views/Timeline.vue')
    },
    {
      path: '/statistics',
      name: 'Statistics',
      component: () => import('@/views/Statistics.vue')
    },
    {
      path: '/reminders',
      name: 'Reminders',
      component: () => import('@/views/Reminders.vue')
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue')
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/Register.vue')
    }
  ]
})

export default router
```

## ⚡ 后端路径配置

### 1. 环境变量配置

#### 开发环境 (.env.development)
```bash
# backend/.env.development
PORT=8010
HOST=localhost
DB_HOST=localhost
DB_PORT=5432
DB_USER=jobview_user
DB_PASSWORD=your_dev_password
DB_NAME=jobview_dev
JWT_SECRET=your-dev-jwt-secret-key
UPLOAD_PATH=./uploads
LOG_PATH=./logs
ENVIRONMENT=development
GIN_MODE=debug
```

#### 生产环境 (.env.production)
```bash
# backend/.env.production
PORT=8010
HOST=127.0.0.1
DB_HOST=localhost
DB_PORT=5432
DB_USER=jobview_user
DB_PASSWORD=your_secure_password
DB_NAME=jobview_prod
JWT_SECRET=your-super-secret-jwt-key-must-be-at-least-32-characters-long
UPLOAD_PATH=/www/wwwroot/jobview.bfsmlt.top/uploads
LOG_PATH=/www/wwwroot/jobview.bfsmlt.top/logs
ENVIRONMENT=production
GIN_MODE=release
```

### 2. 主服务配置 (cmd/main.go)
```go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    "your-module/internal/config"
    "your-module/internal/handler"
    "your-module/internal/database"
)

func main() {
    // 加载环境变量
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }

    envFile := ".env." + env
    if err := godotenv.Load(envFile); err != nil {
        log.Printf("Warning: %s not found, using system env vars", envFile)
    }

    // 初始化配置
    cfg := config.Load()

    // 设置Gin模式
    gin.SetMode(cfg.GinMode)

    // 初始化数据库
    db, err := database.Connect(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect database:", err)
    }

    // 创建上传目录
    if err := os.MkdirAll(cfg.UploadPath, 0755); err != nil {
        log.Printf("Warning: Failed to create upload directory: %v", err)
    }

    // 创建日志目录
    if err := os.MkdirAll(cfg.LogPath, 0755); err != nil {
        log.Printf("Warning: Failed to create log directory: %v", err)
    }

    // 初始化路由
    router := handler.SetupRoutes(db, cfg)

    // 启动服务
    addr := cfg.Host + ":" + cfg.Port
    log.Printf("Server starting on %s", addr)
    if err := router.Run(addr); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

### 3. 路由配置 (internal/handler/routes.go)
```go
package handler

import (
    "net/http"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"

    "your-module/internal/config"
    "your-module/internal/middleware"
)

func SetupRoutes(db *sql.DB, cfg *config.Config) *gin.Engine {
    router := gin.New()

    // 中间件
    router.Use(gin.Logger())
    router.Use(gin.Recovery())

    // CORS配置
    corsConfig := cors.DefaultConfig()
    if cfg.Environment == "development" {
        corsConfig.AllowOrigins = []string{"http://localhost:3000"}
    } else {
        corsConfig.AllowOrigins = []string{"https://jobview.bfsmlt.top"}
    }
    corsConfig.AllowCredentials = true
    corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
    router.Use(cors.New(corsConfig))

    // 健康检查
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })

    // API路由组 (关键：统一使用/api前缀)
    v1 := router.Group("/api/v1")
    {
        // 认证路由
        auth := v1.Group("/auth")
        {
            auth.POST("/register", authHandler.Register)
            auth.POST("/login", authHandler.Login)
            auth.POST("/refresh", authHandler.RefreshToken)
        }

        // 需要认证的路由
        protected := v1.Group("")
        protected.Use(middleware.AuthMiddleware())
        {
            // 投递记录
            applications := protected.Group("/applications")
            {
                applications.GET("", applicationHandler.GetAll)
                applications.GET("/paginated", applicationHandler.GetPaginated)
                applications.GET("/search", applicationHandler.Search)
                applications.POST("", applicationHandler.Create)
                applications.PUT("/:id", applicationHandler.Update)
                applications.DELETE("/:id", applicationHandler.Delete)
                applications.POST("/batch", applicationHandler.BatchOperation)
            }

            // 统计数据
            stats := protected.Group("/stats")
            {
                stats.GET("/overview", statsHandler.GetOverview)
                stats.GET("/trends", statsHandler.GetTrends)
                stats.GET("/database", statsHandler.GetDatabaseStats)
            }

            // 提醒功能
            reminders := protected.Group("/reminders")
            {
                reminders.GET("", reminderHandler.GetAll)
                reminders.POST("", reminderHandler.Create)
                reminders.PUT("/:id/dismiss", reminderHandler.Dismiss)
            }

            // 银月机器人
            robot := protected.Group("/robot")
            {
                robot.POST("/chat", robotHandler.Chat)
                robot.GET("/quick-replies", robotHandler.GetQuickReplies)
                robot.GET("/llm-status", robotHandler.CheckLLMStatus)
            }
        }

        // 文件上传
        v1.POST("/upload", uploadHandler.HandleUpload)
    }

    // 静态文件服务 (生产环境可选)
    if cfg.Environment == "development" {
        router.Static("/uploads", cfg.UploadPath)
    }

    return router
}
```

## 🌐 Nginx 配置

### 生产环境 Nginx 配置
```nginx
server {
    listen 80;
    listen 443 ssl http2;
    server_name jobview.bfsmlt.top;

    # 关键配置：前端静态文件路径
    root /www/wwwroot/jobview.bfsmlt.top/frontend/dist;
    index index.html index.htm;

    # SSL 配置 (保持不变)
    if ($server_port !~ 443){
        rewrite ^(/.*)$ https://$host$1 permanent;
    }
    ssl_certificate    /www/server/panel/vhost/cert/jobview.bfsmlt.top/fullchain.pem;
    ssl_certificate_key  /www/server/panel/vhost/cert/jobview.bfsmlt.top/privkey.pem;
    ssl_protocols TLSv1.1 TLSv1.2 TLSv1.3;
    ssl_ciphers EECDH+CHACHA20:EECDH+CHACHA20-draft:EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:EECDH+3DES:RSA+3DES:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    add_header Strict-Transport-Security "max-age=31536000";
    error_page 497  https://$host$request_uri;

    # 关键配置：前端路由支持 (SPA应用必需)
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 关键配置：API反向代理
    location /api/ {
        proxy_pass http://127.0.0.1:8010;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # 缓冲配置
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # 文件上传支持
    location /api/upload {
        proxy_pass http://127.0.0.1:8010;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        client_max_body_size 10M;
        proxy_request_buffering off;
    }

    # 静态资源优化
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header Vary Accept-Encoding;

        # 启用gzip压缩
        gzip on;
        gzip_vary on;
        gzip_min_length 1024;
        gzip_types
            text/plain
            text/css
            application/json
            application/javascript
            text/xml
            application/xml
            application/xml+rss
            text/javascript
            image/svg+xml;
    }

    # 安全配置
    location ~ ^/(\.user.ini|\.htaccess|\.git|\.env|\.svn|\.project|LICENSE|README.md) {
        return 404;
    }

    # SSL验证
    location ~ \.well-known {
        allow all;
    }

    # 日志配置
    access_log  /www/wwwlogs/jobview.bfsmlt.top.log;
    error_log  /www/wwwlogs/jobview.bfsmlt.top.error.log;
}
```

## 🔍 路径验证

### 1. 开发环境验证
```bash
# 启动前端开发服务器
cd frontend
npm run dev

# 启动后端服务器
cd backend
go run cmd/main.go

# 测试API代理
curl http://localhost:3000/api/v1/health
```

### 2. 生产环境验证
```bash
# 测试前端访问
curl -I https://jobview.bfsmlt.top/

# 测试API代理
curl https://jobview.bfsmlt.top/api/v1/health

# 测试前端路由
curl -I https://jobview.bfsmlt.top/dashboard
curl -I https://jobview.bfsmlt.top/kanban
```

### 3. 路径调试
```bash
# 检查前端构建输出
cat frontend/dist/index.html | grep -o 'src="[^"]*"'

# 检查API请求
# 在浏览器开发者工具中查看Network标签页
# 确认所有API请求都使用/api前缀
```

## 📋 常见问题

### 1. CORS错误
**问题**: 开发环境或生产环境出现跨域错误
**解决**:
```go
// 检查CORS配置
corsConfig.AllowOrigins = []string{
    "http://localhost:3000",      // 开发环境
    "https://jobview.bfsmlt.top"  // 生产环境
}
```

### 2. API路径404
**问题**: API请求返回404错误
**解决**:
1. 检查前端API_BASE配置
2. 检查后端路由注册
3. 检查Nginx代理配置

### 3. 静态资源404
**问题**: CSS/JS文件无法加载
**解决**:
1. 检查Nginx root路径配置
2. 检查文件权限
3. 检查构建输出目录

### 4. 前端路由刷新404
**问题**: 刷新前端页面时出现404
**解决**:
```nginx
# 确保Nginx配置包含
location / {
    try_files $uri $uri/ /index.html;
}
```

## 📝 配置检查清单

### ✅ 前端配置检查
- [ ] 环境变量 VITE_API_BASE 设置为 `/api`
- [ ] Vite代理配置正确
- [ ] 路由使用 HTML5 History 模式
- [ ] API请求使用相对路径

### ✅ 后端配置检查
- [ ] 环境变量正确配置
- [ ] API路由统一使用 `/api/v1` 前缀
- [ ] CORS配置包含正确的域名
- [ ] 文件上传路径配置正确

### ✅ Nginx配置检查
- [ ] root路径指向前端构建目录
- [ ] `/api/` 反向代理配置正确
- [ ] 前端路由fallback配置
- [ ] 静态资源缓存配置

### ✅ 生产环境检查
- [ ] 前端构建文件存在
- [ ] 后端服务正常运行
- [ ] 数据库连接正常
- [ ] SSL证书有效

---

**🛣️ 正确的路径配置是系统正常运行的基础！**

> **记住**: 保持配置的一致性和简洁性，使用环境变量区分不同环境，避免硬编码路径。