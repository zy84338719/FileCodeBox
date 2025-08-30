#!/bin/bash

echo "=== FileCodeBox 用户系统功能测试 ==="
echo

# 基础配置
BASE_URL="http://localhost:12345"
ADMIN_PASSWORD="FileCodeBox2025"

echo "1. 测试用户注册..."
curl -s -X POST $BASE_URL/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demouser",
    "password": "demo123456",
    "email": "demo@example.com"
  }' | jq .
echo

echo "2. 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demouser",
    "password": "demo123456"
  }')
echo $LOGIN_RESPONSE | jq .

# 提取 token
USER_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.detail.token')
echo "用户 Token: $USER_TOKEN"
echo

echo "3. 测试匿名上传..."
echo "测试匿名上传内容" > test_anonymous_upload.txt
curl -s -X POST $BASE_URL/share/file/ \
  -F "file=@test_anonymous_upload.txt" \
  -F "expire_value=1" \
  -F "expire_style=day" | jq .
echo

echo "4. 测试认证用户上传..."
echo "测试认证用户上传内容" > test_auth_upload.txt
AUTH_UPLOAD_RESPONSE=$(curl -s -X POST $BASE_URL/share/file/ \
  -H "Authorization: Bearer $USER_TOKEN" \
  -F "file=@test_auth_upload.txt" \
  -F "expire_value=1" \
  -F "expire_style=day")
echo $AUTH_UPLOAD_RESPONSE | jq .

AUTH_CODE=$(echo $AUTH_UPLOAD_RESPONSE | jq -r '.detail.code')
echo "认证上传的提取码: $AUTH_CODE"
echo

echo "5. 测试需要登录才能下载的文件上传..."
echo "这个文件需要登录才能下载" > test_require_auth_upload.txt
REQUIRE_AUTH_RESPONSE=$(curl -s -X POST $BASE_URL/share/file/ \
  -H "Authorization: Bearer $USER_TOKEN" \
  -F "file=@test_require_auth_upload.txt" \
  -F "expire_value=1" \
  -F "expire_style=day" \
  -F "require_auth=true")
echo $REQUIRE_AUTH_RESPONSE | jq .

REQUIRE_AUTH_CODE=$(echo $REQUIRE_AUTH_RESPONSE | jq -r '.detail.code')
echo "需要认证下载的提取码: $REQUIRE_AUTH_CODE"
echo

echo "6. 测试匿名用户尝试访问需要认证的文件..."
curl -s -X POST $BASE_URL/share/select/ \
  -H "Content-Type: application/json" \
  -d "{\"code\": \"$REQUIRE_AUTH_CODE\"}" | jq .
echo

echo "7. 测试认证用户访问需要认证的文件..."
curl -s -X POST $BASE_URL/share/select/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d "{\"code\": \"$REQUIRE_AUTH_CODE\"}" | jq .
echo

echo "8. 测试用户个人资料..."
curl -s -H "Authorization: Bearer $USER_TOKEN" \
  $BASE_URL/user/profile | jq .
echo

echo "9. 测试用户文件列表..."
curl -s -H "Authorization: Bearer $USER_TOKEN" \
  $BASE_URL/user/files | jq .
echo

echo "10. 测试用户统计信息..."
curl -s -H "Authorization: Bearer $USER_TOKEN" \
  $BASE_URL/user/stats | jq .
echo

echo "11. 测试文本分享（认证用户）..."
curl -s -X POST $BASE_URL/share/text/ \
  -H "Authorization: Bearer $USER_TOKEN" \
  -F "text=这是认证用户分享的文本内容" \
  -F "expire_value=1" \
  -F "expire_style=day" | jq .
echo

# 清理测试文件
rm -f test_anonymous_upload.txt test_auth_upload.txt test_require_auth_upload.txt

echo "=== 测试完成 ==="
