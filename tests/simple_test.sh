#!/bin/bash

# 等待服务器启动
sleep 2

echo "测试管理员配置API..."

# 测试获取配置
echo "获取当前配置:"
RESPONSE=$(curl -s -H "Authorization: Bearer aaaaaa50124" http://localhost:12345/admin/config)
echo "响应: $RESPONSE"

if echo "$RESPONSE" | grep -q '"upload_size"'; then
    UPLOAD_SIZE=$(echo "$RESPONSE" | grep -o '"upload_size":[0-9]*' | grep -o '[0-9]*')
    UPLOAD_SIZE_MB=$((UPLOAD_SIZE / 1024 / 1024))
    echo "当前上传大小限制: ${UPLOAD_SIZE_MB}MB (${UPLOAD_SIZE} bytes)"
else
    echo "无法获取配置信息"
fi

# 测试设置较小的上传限制
echo -e "\n设置上传限制为5MB:"
UPDATE_RESPONSE=$(curl -s -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer aaaaaa50124" \
  -d '{"upload_size": 5242880}' \
  http://localhost:12345/admin/config)
echo "更新响应: $UPDATE_RESPONSE"

# 再次获取配置确认
echo -e "\n确认配置已更新:"
RESPONSE2=$(curl -s -H "Authorization: Bearer aaaaaa50124" http://localhost:12345/admin/config)
if echo "$RESPONSE2" | grep -q '"upload_size"'; then
    UPLOAD_SIZE2=$(echo "$RESPONSE2" | grep -o '"upload_size":[0-9]*' | grep -o '[0-9]*')
    UPLOAD_SIZE_MB2=$((UPLOAD_SIZE2 / 1024 / 1024))
    echo "新的上传大小限制: ${UPLOAD_SIZE_MB2}MB (${UPLOAD_SIZE2} bytes)"
else
    echo "无法获取更新后的配置信息"
fi
