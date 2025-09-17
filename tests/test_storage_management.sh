#!/bin/bash

# 测试存储管理功能

BASE_URL="http://localhost:12345"
# 注意：静态管理员令牌已弃用。测试请先通过管理员用户名/密码登录获取 JWT（可通过 ADMIN_JWT 环境变量注入）
# 示例：ADMIN_JWT=$(curl -s -X POST "$BASE_URL/admin/login" -d '{"username":"admin","password":"yourpass"}' | jq -r '.data.token')
ADMIN_JWT=""

echo "=== 测试存储管理功能 ==="
echo

# 检查服务器是否运行
echo "0. 检查服务器状态..."
if ! curl -s --connect-timeout 2 $BASE_URL > /dev/null; then
    echo "❌ 服务器未运行，请先启动服务器"
    exit 1
fi
echo "✅ 服务器运行正常"
echo

# 获取管理员 Token（优先使用 ADMIN_JWT；若未设置则使用 ADMIN_USERNAME/ADMIN_PASSWORD 登录获取）
echo "1. 管理员登录..."
if [[ -n "$ADMIN_JWT" ]]; then
    JWT_TOKEN="$ADMIN_JWT"
    echo "✅ 已使用环境提供的 ADMIN_JWT"
else
    if [[ -z "$ADMIN_USERNAME" || -z "$ADMIN_PASSWORD" ]]; then
        echo "❌ 未提供 ADMIN_JWT，也未设置 ADMIN_USERNAME/ADMIN_PASSWORD 环境变量。无法登录。"
        exit 1
    fi

    LOGIN_RESULT=$(curl -s -X POST "$BASE_URL/admin/login" \
      -H "Content-Type: application/json" \
      -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}")

    if [[ $LOGIN_RESULT == *"token"* ]]; then
        JWT_TOKEN=$(echo $LOGIN_RESULT | grep -o '"token":"[^\"]*"' | cut -d'"' -f4)
        echo "✅ 管理员登录成功"
    else
        echo "❌ 管理员登录失败"
        echo "详细信息: $LOGIN_RESULT"
        exit 1
    fi
fi
echo

# 获取存储信息
echo "2. 获取存储信息..."
STORAGE_INFO=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/admin/storage")
echo "存储信息: $STORAGE_INFO"
echo

# 测试本地存储连接
echo "3. 测试本地存储连接..."
LOCAL_TEST=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/admin/storage/test/local")
echo "本地存储测试: $LOCAL_TEST"
echo

# 如果有 WebDAV 配置，测试 WebDAV 连接
echo "4. 测试 WebDAV 存储连接..."
WEBDAV_TEST=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/admin/storage/test/webdav")
echo "WebDAV 存储测试: $WEBDAV_TEST"
echo

# 测试存储切换 (切换到本地存储)
echo "5. 测试存储切换..."
SWITCH_RESULT=$(curl -s -X POST "$BASE_URL/admin/storage/switch" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"storage_type":"local"}')
echo "存储切换结果: $SWITCH_RESULT"
echo

# 测试文件上传（验证存储系统工作）
echo "6. 测试文件上传..."
echo "测试存储系统 - $(date)" > test_storage_file.txt

UPLOAD_RESULT=$(curl -s -X POST "$BASE_URL/share/file/" \
  -F "file=@test_storage_file.txt" \
  -F "expire_value=1" \
  -F "expire_style=day")

echo "文件上传结果: $UPLOAD_RESULT"

# 检查文件是否按日期存储
if [[ $UPLOAD_RESULT == *"code"* ]]; then
    FILE_CODE=$(echo $UPLOAD_RESULT | grep -o '"code":"[^"]*"' | cut -d'"' -f4)
    echo "✅ 文件上传成功，文件代码: $FILE_CODE"
    
    # 检查日期目录
    TODAY=$(date "+%Y/%m/%d")
    EXPECTED_PATH="./data/share/data/$TODAY"
    
    if [ -d "$EXPECTED_PATH" ]; then
        echo "✅ 按日期分组存储成功: $EXPECTED_PATH"
        echo "目录内容:"
        ls -la "$EXPECTED_PATH/" | head -5
    else
        echo "❌ 日期分组存储失败"
        echo "查找文件位置:"
        find ./data -name "*test_storage*" 2>/dev/null | head -5
    fi
else
    echo "❌ 文件上传失败"
fi

# 清理测试文件
rm -f test_storage_file.txt

echo
echo "=== 存储管理功能测试完成 ==="
