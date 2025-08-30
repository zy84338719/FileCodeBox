#!/bin/bash

# 简单的配置管理测试
echo "=== FileCodeBox 基础配置测试 ==="

# 首先获取管理员token
echo "1. 管理员登录..."
LOGIN_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
  -d '{"password": "FileCodeBox2025"}' \
  http://localhost:12345/admin/login)

if echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "✅ 登录成功"
else
    echo "❌ 登录失败: $LOGIN_RESPONSE"
    exit 1
fi

echo -e "\n2. 测试获取配置..."
RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:12345/admin/config)

if echo "$RESPONSE" | grep -q '"upload_size"'; then
    UPLOAD_SIZE=$(echo "$RESPONSE" | grep -o '"upload_size":[0-9]*' | grep -o '[0-9]*')
    UPLOAD_SIZE_MB=$((UPLOAD_SIZE / 1024 / 1024))
    echo "✅ 当前上传大小限制: ${UPLOAD_SIZE_MB}MB (${UPLOAD_SIZE} bytes)"
else
    echo "❌ 无法获取配置信息: $RESPONSE"
    exit 1
fi

echo -e "\n3. 测试设置上传限制为5MB..."
UPDATE_RESPONSE=$(curl -s -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"upload_size": 5242880}' \
  http://localhost:12345/admin/config)

if echo "$UPDATE_RESPONSE" | grep -q '"code":200'; then
    echo "✅ 配置更新成功"
else
    echo "❌ 配置更新失败: $UPDATE_RESPONSE"
fi

echo -e "\n4. 确认配置已更新..."
RESPONSE2=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:12345/admin/config)
if echo "$RESPONSE2" | grep -q '"upload_size"'; then
    UPLOAD_SIZE2=$(echo "$RESPONSE2" | grep -o '"upload_size":[0-9]*' | grep -o '[0-9]*')
    UPLOAD_SIZE_MB2=$((UPLOAD_SIZE2 / 1024 / 1024))
    echo "✅ 新的上传大小限制: ${UPLOAD_SIZE_MB2}MB (${UPLOAD_SIZE2} bytes)"
    
    if [ "$UPLOAD_SIZE2" = "5242880" ]; then
        echo "✅ 配置更新验证成功"
    else
        echo "⚠️  配置未正确更新"
    fi
else
    echo "❌ 无法获取更新后的配置信息"
fi

echo -e "\n=== 测试完成 ==="
