#!/bin/bash

# 测试断点续传是否遵循日期存储模式

echo "🔄 测试断点续传的日期存储..."

# 启动应用（后台运行）
cd /Users/zhangyi/FileCodeBox/go
./filecodebox &
APP_PID=$!
sleep 3

echo "📅 测试断点续传日期存储功能..."

# 创建测试文件
TEST_FILE="test_date_storage.bin"
dd if=/dev/urandom of="$TEST_FILE" bs=1M count=2 2>/dev/null

# 计算文件哈希
FILE_HASH=$(sha256sum "$TEST_FILE" | awk '{print $1}')
FILE_SIZE=$(stat -f%z "$TEST_FILE")

echo "📂 文件大小: $FILE_SIZE bytes"
echo "🔑 文件哈希: $FILE_HASH"

# 初始化断点续传
echo "🚀 初始化断点续传..."
INIT_RESPONSE=$(curl -s -X POST "http://localhost:12345/chunk/upload/init/" \
  -H "Content-Type: application/json" \
  -d "{
    \"file_name\": \"$TEST_FILE\",
    \"file_size\": $FILE_SIZE,
    \"file_hash\": \"$FILE_HASH\",
    \"chunk_size\": 524288
  }")

echo "📋 初始化响应: $INIT_RESPONSE"

# 提取uploadID
UPLOAD_ID=$(echo "$INIT_RESPONSE" | grep -o '"upload_id":"[^"]*"' | cut -d'"' -f4)
echo "🆔 上传ID: $UPLOAD_ID"

if [ -z "$UPLOAD_ID" ]; then
    echo "❌ 获取uploadID失败"
    kill $APP_PID
    rm -f "$TEST_FILE"
    exit 1
fi

# 分割文件为块
split -b 524288 "$TEST_FILE" chunk_

# 上传所有块
echo "📤 上传所有块..."
CHUNK_FILES=(chunk_aa chunk_ab chunk_ac chunk_ad)
for i in {0..3}; do
    CHUNK_FILE="${CHUNK_FILES[$i]}"
    if [ -f "$CHUNK_FILE" ]; then
        echo "  📤 上传块 $i: $CHUNK_FILE"
        UPLOAD_RESPONSE=$(curl -s -X POST "http://localhost:12345/chunk/upload/chunk/$UPLOAD_ID/$i" \
          -F "chunk=@$CHUNK_FILE")
        echo "  📤 响应: $UPLOAD_RESPONSE"
    else
        echo "  ❌ 块文件不存在: $CHUNK_FILE"
    fi
done

# 检查存储位置
echo "📁 检查块存储位置..."
CURRENT_DATE=$(date +%Y/%m/%d)
CHUNK_DIR="./data/share/data/chunks/$UPLOAD_ID"

if [ -d "$CHUNK_DIR" ]; then
    echo "✅ 块文件存储在: $CHUNK_DIR"
    ls -la "$CHUNK_DIR"
else
    echo "❌ 块文件目录不存在: $CHUNK_DIR"
fi

# 完成上传（模拟只有一个块的情况）
echo "✅ 完成上传..."
COMPLETE_RESPONSE=$(curl -s -X POST "http://localhost:12345/chunk/upload/complete/$UPLOAD_ID" \
  -H "Content-Type: application/json" \
  -d "{
    \"expire_value\": 7,
    \"expire_style\": \"day\"
  }")

echo "🏁 完成响应: $COMPLETE_RESPONSE"

# 检查最终文件是否存储在日期目录中
echo "📅 检查最终文件日期存储..."
FINAL_FILE_PATTERN="./data/share/data/$CURRENT_DATE/*$TEST_FILE*"
FINAL_FILES=$(find ./data/share/data/$CURRENT_DATE -name "*$TEST_FILE*" 2>/dev/null)

if [ -n "$FINAL_FILES" ]; then
    echo "✅ 文件已正确存储在日期目录中:"
    echo "$FINAL_FILES"
else
    echo "❌ 文件未在日期目录中找到"
    echo "🔍 检查其他位置..."
    find ./data/share/data -name "*$TEST_FILE*" -type f 2>/dev/null
fi

# 清理
echo "🧹 清理测试文件..."
kill $APP_PID
rm -f "$TEST_FILE" chunk_*

echo "✅ 断点续传日期存储测试完成!"
