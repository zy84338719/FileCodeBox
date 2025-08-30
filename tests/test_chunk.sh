#!/bin/bash

# FileCodeBox Go版本分片上传测试脚本

BASE_URL="http://localhost:12345"

echo "=== FileCodeBox Go版本 分片上传测试 ==="
echo

# 创建一个大一点的测试文件
echo "1. 创建测试文件..."
dd if=/dev/zero of=large_test_file.bin bs=1024 count=100 2>/dev/null
echo "创建了100KB的测试文件"

# 计算文件哈希
FILE_HASH=$(sha256sum large_test_file.bin | cut -d' ' -f1)
FILE_SIZE=$(stat -f%z large_test_file.bin 2>/dev/null || stat -c%s large_test_file.bin)
CHUNK_SIZE=32768  # 32KB分片

echo "文件大小: $FILE_SIZE 字节"
echo "文件哈希: $FILE_HASH"
echo "分片大小: $CHUNK_SIZE 字节"
echo

# 初始化分片上传
echo "2. 初始化分片上传..."
INIT_RESULT=$(curl -s -X POST "${BASE_URL}/chunk/upload/init/" \
  -H "Content-Type: application/json" \
  -d "{
    \"file_name\": \"large_test_file.bin\",
    \"file_size\": $FILE_SIZE,
    \"chunk_size\": $CHUNK_SIZE,
    \"file_hash\": \"$FILE_HASH\"
  }")

echo "初始化结果: $INIT_RESULT"
UPLOAD_ID=$(echo $INIT_RESULT | jq -r '.detail.upload_id' 2>/dev/null)
TOTAL_CHUNKS=$(echo $INIT_RESULT | jq -r '.detail.total_chunks' 2>/dev/null)

if [ "$UPLOAD_ID" != "null" ] && [ "$UPLOAD_ID" != "" ]; then
    echo "分片上传初始化成功，上传ID: $UPLOAD_ID"
    echo "总分片数: $TOTAL_CHUNKS"
    echo
    
    # 分割文件并上传分片
    echo "3. 上传分片..."
    split -b $CHUNK_SIZE large_test_file.bin chunk_
    
    CHUNK_INDEX=0
    for chunk_file in chunk_*; do
        echo "上传分片 $CHUNK_INDEX..."
        CHUNK_RESULT=$(curl -s -X POST "${BASE_URL}/chunk/upload/chunk/$UPLOAD_ID/$CHUNK_INDEX" \
          -F "chunk=@$chunk_file")
        
        echo "分片 $CHUNK_INDEX 结果: $CHUNK_RESULT"
        CHUNK_INDEX=$((CHUNK_INDEX + 1))
    done
    
    echo
    echo "4. 完成上传..."
    COMPLETE_RESULT=$(curl -s -X POST "${BASE_URL}/chunk/upload/complete/$UPLOAD_ID" \
      -H "Content-Type: application/json" \
      -d "{
        \"expire_value\": 1,
        \"expire_style\": \"day\"
      }")
    
    echo "完成上传结果: $COMPLETE_RESULT"
    
    # 清理临时文件
    rm -f chunk_*
else
    echo "分片上传初始化失败"
fi

# 清理测试文件
rm -f large_test_file.bin

echo
echo "=== 分片上传测试完成 ==="
