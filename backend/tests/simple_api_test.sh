#!/bin/bash

# 简化的API测试脚本

echo "=== JobView API测试 ==="

# 测试健康检查
echo "1. 健康检查..."
curl -s http://localhost:8010/health | jq .

# 测试登录（使用默认测试用户）
echo "2. 用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8010/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"TestPass123!"}')

echo "登录响应: $LOGIN_RESPONSE"

# 提取token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token // empty')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "❌ 登录失败，无法获取token"
    exit 1
fi

echo "✅ 登录成功，Token: ${TOKEN:0:20}..."

# 测试状态定义API
echo "3. 测试状态定义API..."
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8010/api/v1/status-definitions | jq .

# 测试岗位申请API
echo "4. 测试岗位申请列表..."
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8010/api/v1/applications | jq .

echo "=== API测试完成 ==="