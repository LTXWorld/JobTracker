# 🚀 GitHub Actions CI/CD 自动部署

> 基于GitHub Actions的完整持续集成和部署流程

## 📋 GitHub Actions 工作流配置

### 1. 主要工作流文件
```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME_FRONTEND: ${{ github.repository }}/frontend
  IMAGE_NAME_BACKEND: ${{ github.repository }}/backend

jobs:
  # 测试阶段
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: jobview_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json

    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'

    # 前端测试
    - name: Install frontend dependencies
      run: |
        cd frontend
        npm ci

    - name: Run frontend tests
      run: |
        cd frontend
        npm run test:ci

    - name: Frontend build test
      run: |
        cd frontend
        npm run build

    # 后端测试
    - name: Install backend dependencies
      run: |
        cd backend
        go mod download

    - name: Run backend tests
      run: |
        cd backend
        go test ./... -v
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_USER: postgres
        DB_PASSWORD: test_password
        DB_NAME: jobview_test

    - name: Backend build test
      run: |
        cd backend
        go build -v ./cmd/main.go

  # 构建和推送Docker镜像
  build-and-push:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    permissions:
      contents: read
      packages: write

    outputs:
      frontend-image: ${{ steps.meta-frontend.outputs.tags }}
      backend-image: ${{ steps.meta-backend.outputs.tags }}

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    # 构建前端镜像
    - name: Extract frontend metadata
      id: meta-frontend
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME_FRONTEND }}
        tags: |
          type=ref,event=branch
          type=sha,prefix={{branch}}-
          type=raw,value=latest,enable={{is_default_branch}}

    - name: Build and push frontend image
      uses: docker/build-push-action@v5
      with:
        context: ./frontend
        push: true
        tags: ${{ steps.meta-frontend.outputs.tags }}
        labels: ${{ steps.meta-frontend.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

    # 构建后端镜像
    - name: Extract backend metadata
      id: meta-backend
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME_BACKEND }}
        tags: |
          type=ref,event=branch
          type=sha,prefix={{branch}}-
          type=raw,value=latest,enable={{is_default_branch}}

    - name: Build and push backend image
      uses: docker/build-push-action@v5
      with:
        context: ./backend
        push: true
        tags: ${{ steps.meta-backend.outputs.tags }}
        labels: ${{ steps.meta-backend.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  # 部署到生产环境
  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    environment:
      name: production
      url: https://jobview.bfsmlt.top

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Deploy to server
      uses: appleboy/ssh-action@v1.0.0
      env:
        FRONTEND_IMAGE: ${{ needs.build-and-push.outputs.frontend-image }}
        BACKEND_IMAGE: ${{ needs.build-and-push.outputs.backend-image }}
      with:
        host: ${{ secrets.HOST }}
        username: ${{ secrets.USERNAME }}
        key: ${{ secrets.SSH_KEY }}
        port: ${{ secrets.PORT }}
        envs: FRONTEND_IMAGE,BACKEND_IMAGE
        script: |
          # 进入项目目录
          cd /www/wwwroot/jobview.bfsmlt.top

          # 拉取最新代码
          git pull origin main

          # 登录到GitHub Container Registry
          echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin

          # 停止现有服务
          docker-compose down

          # 拉取最新镜像
          docker pull $FRONTEND_IMAGE
          docker pull $BACKEND_IMAGE

          # 更新docker-compose.yml中的镜像标签
          sed -i "s|ghcr.io/${{ github.repository }}/frontend:.*|$FRONTEND_IMAGE|g" docker-compose.prod.yml
          sed -i "s|ghcr.io/${{ github.repository }}/backend:.*|$BACKEND_IMAGE|g" docker-compose.prod.yml

          # 启动服务
          docker-compose -f docker-compose.prod.yml up -d

          # 等待服务启动
          sleep 30

          # 健康检查
          curl -f https://jobview.bfsmlt.top/api/v1/health || exit 1

          # 清理未使用的镜像
          docker image prune -f

    - name: Health Check
      run: |
        # 等待部署完成
        sleep 60

        # 检查服务状态
        response=$(curl -s -o /dev/null -w "%{http_code}" https://jobview.bfsmlt.top/api/v1/health)
        if [ $response -eq 200 ]; then
          echo "✅ Deployment successful!"
        else
          echo "❌ Deployment failed! HTTP status: $response"
          exit 1
        fi

    - name: Notify deployment status
      if: always()
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        channel: '#deployments'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        fields: repo,message,commit,author,action,eventName,ref,workflow
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### 2. 分离的工作流 - 前端构建
```yaml
# .github/workflows/frontend.yml
name: Frontend CI

on:
  push:
    paths:
      - 'frontend/**'
      - '.github/workflows/frontend.yml'
  pull_request:
    paths:
      - 'frontend/**'

jobs:
  frontend-ci:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json

    - name: Install dependencies
      run: |
        cd frontend
        npm ci

    - name: Type check
      run: |
        cd frontend
        npm run type-check

    - name: Lint check
      run: |
        cd frontend
        npm run lint

    - name: Run tests
      run: |
        cd frontend
        npm run test:ci

    - name: Build for production
      run: |
        cd frontend
        npm run build

    - name: Upload build artifacts
      uses: actions/upload-artifact@v4
      with:
        name: frontend-dist
        path: frontend/dist/
        retention-days: 7
```

### 3. 分离的工作流 - 后端构建
```yaml
# .github/workflows/backend.yml
name: Backend CI

on:
  push:
    paths:
      - 'backend/**'
      - '.github/workflows/backend.yml'
  pull_request:
    paths:
      - 'backend/**'

jobs:
  backend-ci:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: jobview_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'

    - name: Cache Go modules
      uses: actions/cache@v4
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Install dependencies
      run: |
        cd backend
        go mod download

    - name: Run go vet
      run: |
        cd backend
        go vet ./...

    - name: Run go fmt check
      run: |
        cd backend
        if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
          echo "Code is not formatted:"
          gofmt -s -l .
          exit 1
        fi

    - name: Run tests
      run: |
        cd backend
        go test -v -race -coverprofile=coverage.out ./...
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_USER: postgres
        DB_PASSWORD: test_password
        DB_NAME: jobview_test

    - name: Upload coverage reports
      uses: codecov/codecov-action@v4
      with:
        file: backend/coverage.out
        flags: backend
        name: backend-coverage

    - name: Build binary
      run: |
        cd backend
        go build -v -o jobview-server cmd/main.go

    - name: Upload build artifacts
      uses: actions/upload-artifact@v4
      with:
        name: backend-binary
        path: backend/jobview-server
        retention-days: 7
```

### 4. 数据库迁移工作流
```yaml
# .github/workflows/migration.yml
name: Database Migration

on:
  push:
    paths:
      - 'backend/migrations/**'
      - 'scripts/migrate-database.sh'
  workflow_dispatch:
    inputs:
      force_migration:
        description: 'Force run migration'
        required: false
        default: 'false'

jobs:
  migration:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    environment:
      name: production

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Run database migration
      uses: appleboy/ssh-action@v1.0.0
      with:
        host: ${{ secrets.HOST }}
        username: ${{ secrets.USERNAME }}
        key: ${{ secrets.SSH_KEY }}
        port: ${{ secrets.PORT }}
        script: |
          cd /www/wwwroot/jobview.bfsmlt.top

          # 备份数据库
          docker-compose exec -T postgres pg_dump -U ${{ secrets.DB_USER }} ${{ secrets.DB_NAME }} > backup_$(date +%Y%m%d_%H%M%S).sql

          # 运行迁移
          ./scripts/migrate-database.sh

          # 验证迁移
          docker-compose exec -T postgres psql -U ${{ secrets.DB_USER }} -d ${{ secrets.DB_NAME }} -c "\dt"

    - name: Notify migration status
      if: always()
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: 'Database migration completed'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

### 5. 定时健康检查工作流
```yaml
# .github/workflows/health-check.yml
name: Production Health Check

on:
  schedule:
    # 每15分钟检查一次
    - cron: '*/15 * * * *'
  workflow_dispatch:

jobs:
  health-check:
    runs-on: ubuntu-latest

    steps:
    - name: Check API Health
      run: |
        response=$(curl -s -o /dev/null -w "%{http_code}" https://jobview.bfsmlt.top/api/v1/health)
        if [ $response -eq 200 ]; then
          echo "✅ API is healthy"
        else
          echo "❌ API health check failed! HTTP status: $response"
          exit 1
        fi

    - name: Check Frontend
      run: |
        response=$(curl -s -o /dev/null -w "%{http_code}" https://jobview.bfsmlt.top/)
        if [ $response -eq 200 ]; then
          echo "✅ Frontend is accessible"
        else
          echo "❌ Frontend check failed! HTTP status: $response"
          exit 1
        fi

    - name: Check Database Connection
      uses: appleboy/ssh-action@v1.0.0
      if: failure()
      with:
        host: ${{ secrets.HOST }}
        username: ${{ secrets.USERNAME }}
        key: ${{ secrets.SSH_KEY }}
        port: ${{ secrets.PORT }}
        script: |
          cd /www/wwwroot/jobview.bfsmlt.top
          docker-compose exec -T postgres pg_isready -U ${{ secrets.DB_USER }} -d ${{ secrets.DB_NAME }}

    - name: Notify on failure
      if: failure()
      uses: 8398a7/action-slack@v3
      with:
        status: 'failure'
        text: '🚨 Production health check failed!'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

## 🔧 GitHub Secrets 配置

### 必需的Secrets
在GitHub仓库设置中添加以下Secrets：

```bash
# 服务器连接
HOST=your.server.ip.address
USERNAME=root
SSH_KEY=your_private_ssh_key
PORT=22

# 数据库配置
DB_USER=jobview_user
DB_PASSWORD=your_secure_database_password
DB_NAME=jobview_prod
JWT_SECRET=your-super-secret-jwt-key-must-be-at-least-32-characters-long

# 通知配置 (可选)
SLACK_WEBHOOK=your_slack_webhook_url

# GitHub Token (自动提供)
GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

### SSH密钥配置
```bash
# 在本地生成SSH密钥对
ssh-keygen -t rsa -b 4096 -C "github-actions@jobview.com" -f ~/.ssh/github_actions

# 将公钥添加到服务器
ssh-copy-id -i ~/.ssh/github_actions.pub root@your.server.ip

# 将私钥内容添加到GitHub Secrets中的SSH_KEY
cat ~/.ssh/github_actions
```

## 📦 生产环境Docker Compose配置

### docker-compose.prod.yml
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: jobview-postgres
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - /www/wwwroot/jobview.bfsmlt.top/backup:/backup
    networks:
      - jobview-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 30s
      timeout: 10s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: jobview-redis
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    networks:
      - jobview-network
    restart: unless-stopped

  backend:
    image: ghcr.io/your-username/jobview/backend:latest
    container_name: jobview-backend
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - JWT_SECRET=${JWT_SECRET}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - ENVIRONMENT=production
      - GIN_MODE=release
      - PORT=8010
      - HOST=0.0.0.0
    volumes:
      - backend_logs:/root/logs
      - backend_uploads:/root/uploads
    networks:
      - jobview-network
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

  frontend:
    image: ghcr.io/your-username/jobview/frontend:latest
    container_name: jobview-frontend
    networks:
      - jobview-network
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    container_name: jobview-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/conf.d:/etc/nginx/conf.d
      - /www/server/panel/vhost/cert/jobview.bfsmlt.top:/etc/nginx/ssl
    networks:
      - jobview-network
    depends_on:
      - frontend
      - backend
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  backend_logs:
  backend_uploads:

networks:
  jobview-network:
    driver: bridge
```

## 🔄 自动部署流程说明

### 1. 触发条件
- **代码推送到main分支** → 自动触发完整CI/CD流程
- **Pull Request** → 运行测试，不部署
- **定时任务** → 每15分钟健康检查
- **手动触发** → 支持手动运行特定工作流

### 2. 部署步骤
1. **代码检出** → 获取最新代码
2. **运行测试** → 前端和后端测试
3. **构建镜像** → 构建Docker镜像并推送到GitHub Container Registry
4. **部署到服务器** → SSH连接服务器，更新Docker容器
5. **健康检查** → 验证部署成功
6. **通知** → 发送部署结果通知

### 3. 回滚机制
```bash
# 如果部署失败，可以快速回滚到上一个版本
docker-compose down
docker tag ghcr.io/your-username/jobview/frontend:previous ghcr.io/your-username/jobview/frontend:latest
docker tag ghcr.io/your-username/jobview/backend:previous ghcr.io/your-username/jobview/backend:latest
docker-compose up -d
```

---

**🚀 GitHub Actions CI/CD配置完成！接下来创建部署脚本和服务器配置。**