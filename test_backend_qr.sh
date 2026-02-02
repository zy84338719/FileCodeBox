#!/bin/bash

# 后端二维码功能测试脚本

echo "=== 后端二维码功能测试 ==="
echo

# 启动服务
echo "1. 启动 FileCodeBox 服务..."
cd /Users/zhangyi/FileCodeBox
go run main.go &

# 等待服务启动
sleep 5

# 测试二维码API
echo "2. 测试二维码生成API..."
TEST_DATA="https://example.com/test-share"
TEST_SIZE="200"

echo "测试PNG格式二维码生成..."
PNG_RESPONSE=$(curl -s -w "%{http_code}" "http://localhost:12345/api/qrcode/generate?data=${TEST_DATA}&size=${TEST_SIZE}" -o /tmp/test_qr.png)

HTTP_CODE="${PNG_RESPONSE: -3}"
CONTENT_LENGTH=$(wc -c < /tmp/test_qr.png)

echo "HTTP状态码: $HTTP_CODE"
echo "图片大小: ${CONTENT_LENGTH} bytes"

if [ "$HTTP_CODE" = "200" ] && [ "$CONTENT_LENGTH" -gt 100 ]; then
    echo "✅ PNG二维码生成成功"
    file /tmp/test_qr.png
else
    echo "❌ PNG二维码生成失败"
fi

echo

echo "测试Base64格式二维码生成..."
BASE64_RESPONSE=$(curl -s "http://localhost:12345/api/qrcode/base64?data=${TEST_DATA}&size=${TEST_SIZE}")

echo "Base64响应:"
echo "$BASE64_RESPONSE" | jq '.' 2>/dev/null || echo "$BASE64_RESPONSE"

if echo "$BASE64_RESPONSE" | grep -q "qr_code" && echo "$BASE64_RESPONSE" | grep -q "data:image/png;base64"; then
    echo "✅ Base64二维码生成成功"
else
    echo "❌ Base64二维码生成失败"
fi

echo

# 测试分享功能的二维码集成
echo "3. 测试分享功能集成..."
echo "创建测试文件..."
echo "二维码集成测试文件内容" > test_integration.txt

SHARE_RESPONSE=$(curl -s -X POST "http://localhost:12345/share/file/" \
  -F "file=@test_integration.txt" \
  -F "expire_value=1" \
  -F "expire_style=day")

echo "分享响应:"
echo "$SHARE_RESPONSE" | jq '.' 2>/dev/null || echo "$SHARE_RESPONSE"

if echo "$SHARE_RESPONSE" | grep -q "full_share_url" && echo "$SHARE_RESPONSE" | grep -q "qr_code_data"; then
    echo "✅ 分享功能包含二维码数据字段"
    
    # 提取二维码数据
    QR_DATA=$(echo "$SHARE_RESPONSE" | jq -r '.data.full_share_url' 2>/dev/null)
    if [ "$QR_DATA" != "null" ] && [ -n "$QR_DATA" ]; then
        echo "二维码数据: $QR_DATA"
        
        # 测试使用该数据生成二维码
        INTEGRATION_QR=$(curl -s -w "%{http_code}" "http://localhost:12345/api/qrcode/generate?data=${QR_DATA}&size=200" -o /tmp/integration_qr.png)
        INTEGRATION_CODE="${INTEGRATION_QR: -3}"
        INTEGRATION_SIZE=$(wc -c < /tmp/integration_qr.png)
        
        if [ "$INTEGRATION_CODE" = "200" ] && [ "$INTEGRATION_SIZE" -gt 100 ]; then
            echo "✅ 分享用二维码生成成功 (${INTEGRATION_SIZE} bytes)"
        else
            echo "❌ 分享用二维码生成失败"
        fi
    fi
else
    echo "❌ 分享功能缺少二维码数据字段"
fi

echo

# 清理测试文件
rm -f test_integration.txt /tmp/test_qr.png /tmp/integration_qr.png

# 关闭服务
echo "4. 关闭测试服务..."
pkill -f "go run main.go"

echo "后端二维码功能测试完成！"