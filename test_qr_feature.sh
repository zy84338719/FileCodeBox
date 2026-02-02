#!/bin/bash

# 二维码功能测试脚本

echo "=== 二维码功能测试 ==="
echo

# 启动服务
echo "1. 启动 FileCodeBox 服务..."
cd /Users/zhangyi/FileCodeBox
go run main.go &

# 等待服务启动
sleep 3

# 测试文本分享功能
echo "2. 测试文本分享并检查二维码数据..."
TEXT_RESPONSE=$(curl -s -X POST "http://localhost:12345/share/text/" \
  -d "text=测试二维码功能" \
  -d "expire_value=1" \
  -d "expire_style=day")

echo "文本分享响应:"
echo "$TEXT_RESPONSE" | jq '.' 2>/dev/null || echo "$TEXT_RESPONSE"
echo

# 检查响应中是否包含二维码相关字段
if echo "$TEXT_RESPONSE" | grep -q "full_share_url" && echo "$TEXT_RESPONSE" | grep -q "qr_code_data"; then
    echo "✅ 文本分享响应包含二维码数据字段"
else
    echo "❌ 文本分享响应缺少二维码数据字段"
fi

echo

# 测试文件上传功能
echo "3. 测试文件上传并检查二维码数据..."
echo "创建测试文件..."
echo "这是一个测试文件内容" > test_qr_file.txt

FILE_RESPONSE=$(curl -s -X POST "http://localhost:12345/share/file/" \
  -F "file=@test_qr_file.txt" \
  -F "expire_value=1" \
  -F "expire_style=day")

echo "文件上传响应:"
echo "$FILE_RESPONSE" | jq '.' 2>/dev/null || echo "$FILE_RESPONSE"
echo

# 检查响应中是否包含二维码相关字段
if echo "$FILE_RESPONSE" | grep -q "full_share_url" && echo "$FILE_RESPONSE" | grep -q "qr_code_data"; then
    echo "✅ 文件上传响应包含二维码数据字段"
else
    echo "❌ 文件上传响应缺少二维码数据字段"
fi

echo

# 清理测试文件
rm -f test_qr_file.txt

# 关闭服务
echo "4. 关闭测试服务..."
pkill -f "go run main.go"

echo "测试完成！"