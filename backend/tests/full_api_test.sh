#!/bin/bash

# 注册测试用户并进行API测试

echo "=== JobView API测试 ==="

# 注册新用户
echo "1. 注册测试用户..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8010/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testapi","email":"testapi@example.com","password":"TestPass123!"}')

echo "注册响应: $REGISTER_RESPONSE"

# 登录
echo "2. 用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8010/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testapi","password":"TestPass123!"}')

echo "登录响应: $LOGIN_RESPONSE"

# 提取token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token // empty')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "❌ 登录失败，无法获取token"
    exit 1
fi

echo "✅ 登录成功，Token: ${TOKEN:0:30}..."

# 创建测试岗位
echo "3. 创建测试岗位..."
JOB_RESPONSE=$(curl -s -X POST http://localhost:8010/api/v1/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"company_name":"测试公司","position_title":"Go工程师","status":"已投递"}')

echo "创建岗位响应: $JOB_RESPONSE"

# 提取job ID
JOB_ID=$(echo $JOB_RESPONSE | jq -r '.data.id // empty')
if [ -z "$JOB_ID" ] || [ "$JOB_ID" = "null" ]; then
    echo "❌ 创建岗位失败，无法获取ID"
    exit 1
fi

echo "✅ 创建岗位成功，ID: $JOB_ID"

# 测试状态更新
echo "4. 测试状态更新..."
STATUS_UPDATE_RESPONSE=$(curl -s -X POST http://localhost:8010/api/v1/job-applications/$JOB_ID/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"简历筛选中","note":"API测试"}')

echo "状态更新响应: $STATUS_UPDATE_RESPONSE"

# 测试状态历史
echo "5. 测试状态历史查询..."
HISTORY_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8010/api/v1/job-applications/$JOB_ID/status-history")

echo "状态历史响应: $HISTORY_RESPONSE"

# 测试状态分析
echo "6. 测试状态分析..."
ANALYTICS_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8010/api/v1/job-applications/status-analytics")

echo "状态分析响应: $ANALYTICS_RESPONSE"

echo "=== API测试完成 ==="