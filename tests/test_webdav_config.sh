#!/bin/bash

# 测试 WebDAV 存储配置

BASE_URL="http://localhost:12345"
# 注意：静态管理员令牌已弃用。请先通过管理员用户名/密码登录并获取 JWT，用于后续请求（或通过 ADMIN_JWT 环境变量注入）。
ADMIN_JWT=""

echo "=== 测试 WebDAV 存储配置 ==="
echo

# 管理员登录（示例：使用管理员用户名+密码获取 JWT）
echo "1. 管理员登录..."
LOGIN_RESULT=$(curl -s -X POST "$BASE_URL/admin/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"FileCodeBox2025"}')

JWT_TOKEN=$(echo $LOGIN_RESULT | grep -o '"token":"[^\"]*"' | cut -d'"' -f4)
if [ -z "$JWT_TOKEN" ]; then
  echo "❌ 无法获取 JWT，登录可能失败： $LOGIN_RESULT"
  exit 1
fi
echo "✅ 登录成功"
echo

# 配置 WebDAV 存储（使用测试配置）
echo "2. 配置 WebDAV 存储..."
WEBDAV_CONFIG=$(curl -s -X PUT "$BASE_URL/admin/storage/config" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "storage_type": "webdav",
    "config": {
      "hostname": "https://example-webdav.com",
      "username": "test_user",
      "password": "test_password",
      "root_path": "filecodebox_test"
    }
  }')

echo "WebDAV 配置结果: $WEBDAV_CONFIG"
echo

# 获取更新后的存储信息
echo "3. 获取更新后的存储信息..."
STORAGE_INFO=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/admin/storage")
echo "存储信息: $STORAGE_INFO"
echo

# 测试 WebDAV 连接（预期会失败，因为是假的服务器）
echo "4. 测试 WebDAV 连接..."
WEBDAV_TEST=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/admin/storage/test/webdav")
echo "WebDAV 连接测试: $WEBDAV_TEST"
echo

# 测试本地存储配置更新
echo "5. 更新本地存储配置..."
LOCAL_CONFIG=$(curl -s -X PUT "$BASE_URL/admin/storage/config" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "storage_type": "local",
    "config": {
      "storage_path": "./data/custom_storage"
    }
  }')

echo "本地存储配置结果: $LOCAL_CONFIG"
echo

echo "=== WebDAV 存储配置测试完成 ==="
