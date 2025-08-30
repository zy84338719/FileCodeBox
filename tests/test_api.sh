#!/bin/bash

# FileCodeBox Go版本测试脚本

BASE_URL="http://localhost:12345"

echo "=== FileCodeBox Go版本 API测试 ==="
echo

# 测试获取配置
echo "1. 测试获取配置..."
curl -s -X POST "${BASE_URL}/" | jq '.' || echo "配置获取失败"
echo

# 测试分享文本
echo "2. 测试分享文本..."
TEXT_RESULT=$(curl -s -X POST "${BASE_URL}/share/text/" \
  -d "text=这是一个测试文本内容" \
  -d "expire_value=1" \
  -d "expire_style=day")

echo "分享结果: $TEXT_RESULT"
TEXT_CODE=$(echo $TEXT_RESULT | jq -r '.detail.code' 2>/dev/null)

if [ "$TEXT_CODE" != "null" ] && [ "$TEXT_CODE" != "" ]; then
    echo "文本分享成功，提取码: $TEXT_CODE"
    
    # 测试获取文本
    echo "3. 测试获取文本..."
    curl -s -X POST "${BASE_URL}/share/select/" \
      -H "Content-Type: application/json" \
      -d "{\"code\": \"$TEXT_CODE\"}" | jq '.' || echo "获取文本失败"
else
    echo "文本分享失败"
fi
echo

# 测试文件上传
echo "4. 测试文件上传..."
echo "创建测试文件..."
echo "这是一个测试文件内容" > test_file.txt

FILE_RESULT=$(curl -s -X POST "${BASE_URL}/share/file/" \
  -F "file=@test_file.txt" \
  -F "expire_value=1" \
  -F "expire_style=day")

echo "文件上传结果: $FILE_RESULT"
FILE_CODE=$(echo $FILE_RESULT | jq -r '.detail.code' 2>/dev/null)

if [ "$FILE_CODE" != "null" ] && [ "$FILE_CODE" != "" ]; then
    echo "文件上传成功，提取码: $FILE_CODE"
    
    # 测试获取文件
    echo "5. 测试获取文件信息..."
    curl -s -X POST "${BASE_URL}/share/select/" \
      -H "Content-Type: application/json" \
      -d "{\"code\":\"$FILE_CODE\"}" | jq '.'
else
    echo "文件上传失败"
fi

# 清理测试文件
rm -f test_file.txt

echo
echo "=== 测试完成 ==="
