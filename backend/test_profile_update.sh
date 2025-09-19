#!/bin/bash

# 测试用户信息更新API的修复
# 首先登录获取token

echo "=== 测试用户信息更新API修复 ==="

# 1. 登录获取token
echo "1. 登录获取token..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8010/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "TestPass123!"
  }')

# 提取token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ 登录失败，无法获取token"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ 登录成功，获取到token"

# 2. 获取当前用户信息
echo "2. 获取当前用户信息..."
PROFILE_RESPONSE=$(curl -s -X GET http://localhost:8010/api/auth/profile \
  -H "Authorization: Bearer $TOKEN")

echo "当前用户信息: $PROFILE_RESPONSE"

# 3. 测试更新用户名
echo "3. 测试更新用户名..."
UPDATE_RESPONSE=$(curl -s -X PUT http://localhost:8010/api/auth/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "username": "testuser_updated"
  }')

echo "更新用户名响应: $UPDATE_RESPONSE"

# 检查更新是否成功
if echo $UPDATE_RESPONSE | jq -e '.code == 200' > /dev/null; then
  echo "✅ 用户名更新成功"
else
  echo "❌ 用户名更新失败"
  echo "Error: $UPDATE_RESPONSE"
fi

# 4. 测试更新邮箱
echo "4. 测试更新邮箱..."
UPDATE_EMAIL_RESPONSE=$(curl -s -X PUT http://localhost:8010/api/auth/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "email": "testuser_updated@example.com"
  }')

echo "更新邮箱响应: $UPDATE_EMAIL_RESPONSE"

# 检查更新是否成功
if echo $UPDATE_EMAIL_RESPONSE | jq -e '.code == 200' > /dev/null; then
  echo "✅ 邮箱更新成功"
else
  echo "❌ 邮箱更新失败"
  echo "Error: $UPDATE_EMAIL_RESPONSE"
fi

# 5. 测试同时更新用户名和邮箱
echo "5. 测试同时更新用户名和邮箱..."
UPDATE_BOTH_RESPONSE=$(curl -s -X PUT http://localhost:8010/api/auth/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "username": "testuser_final",
    "email": "testuser_final@example.com"
  }')

echo "同时更新响应: $UPDATE_BOTH_RESPONSE"

# 检查更新是否成功
if echo $UPDATE_BOTH_RESPONSE | jq -e '.code == 200' > /dev/null; then
  echo "✅ 同时更新用户名和邮箱成功"
else
  echo "❌ 同时更新失败"
  echo "Error: $UPDATE_BOTH_RESPONSE"
fi

# 6. 验证最终状态
echo "6. 验证最终用户信息..."
FINAL_PROFILE_RESPONSE=$(curl -s -X GET http://localhost:8010/api/auth/profile \
  -H "Authorization: Bearer $TOKEN")

echo "最终用户信息: $FINAL_PROFILE_RESPONSE"

echo "=== 测试完成 ==="