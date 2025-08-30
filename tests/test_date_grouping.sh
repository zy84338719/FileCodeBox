#!/bin/bash

# 测试日期分组存储功能

#!/bin/bash

# 测试日期分组存储功能

echo "=== 测试按日期分组存储文件 ==="
echo

# 检查服务器是否运行
echo "0. 检查服务器状态..."
for port in 12345 8080 3000; do
    if curl -s --connect-timeout 2 http://localhost:$port > /dev/null; then
        BASE_URL="http://localhost:$port"
        echo "✅ 服务器运行在端口 $port"
        break
    fi
done

if [ -z "$BASE_URL" ]; then
    echo "❌ 服务器未运行，尝试查看进程..."
    ps aux | grep filecodebox | grep -v grep || echo "未找到 filecodebox 进程"
    echo "请先启动服务器"
    exit 1
fi

# 创建测试文件
echo "1. 创建测试文件..."
echo "Hello World! This is a test file for date grouping - $(date)" > test_date_grouping.txt

# 上传文件 - 尝试不同的端点
echo "2. 上传文件..."
echo "尝试端点: ${BASE_URL}/share/file/"
UPLOAD_RESULT=$(curl -s -X POST "${BASE_URL}/share/file/" \
  -F "file=@test_date_grouping.txt" \
  -F "expire_value=1" \
  -F "expire_style=day")

echo "上传结果: $UPLOAD_RESULT"

# 检查结果并提取文件代码
if [[ $UPLOAD_RESULT == *"code"* ]]; then
    # 尝试多种方式提取代码
    FILE_CODE=$(echo $UPLOAD_RESULT | jq -r '.detail.code' 2>/dev/null)
    if [ "$FILE_CODE" == "null" ] || [ "$FILE_CODE" == "" ]; then
        FILE_CODE=$(echo $UPLOAD_RESULT | grep -o '"code":"[^"]*"' | cut -d'"' -f4)
    fi
    
    if [ "$FILE_CODE" != "" ]; then
        echo "✅ 文件上传成功!"
        echo "文件代码: $FILE_CODE"
        echo
        
        # 检查文件存储路径
        echo "3. 检查文件存储结构..."
        TODAY=$(date "+%Y/%m/%d")
        echo "预期存储路径: ./data/share/data/$TODAY/"
        
        if [ -d "./data/share/data/$TODAY" ]; then
            echo "✅ 日期目录创建成功!"
            echo "目录内容:"
            ls -la "./data/share/data/$TODAY/"
            echo
        else
            echo "❌ 日期目录未创建，检查实际结构..."
            echo "data目录结构:"
            find ./data -type d 2>/dev/null | head -20
            echo "查找所有文件:"
            find ./data -type f -name "*test*" 2>/dev/null | head -10
        fi
        
        # 测试下载
        echo "4. 测试文件下载..."
        DOWNLOAD_RESULT=$(curl -s "${BASE_URL}/s/$FILE_CODE" | head -30)
        if [[ $DOWNLOAD_RESULT == *"Hello World"* ]] || [[ $DOWNLOAD_RESULT == *"test file"* ]]; then
            echo "✅ 文件下载成功"
        else
            echo "❌ 文件下载失败"
            echo "下载响应前30行: $DOWNLOAD_RESULT"
        fi
    else
        echo "❌ 无法提取文件代码"
        echo "完整响应: $UPLOAD_RESULT"
    fi
else
    echo "❌ 文件上传失败"
    echo "详细错误: $UPLOAD_RESULT"
    
    # 尝试其他端点
    echo
    echo "尝试其他端点: ${BASE_URL}/share/file"
    UPLOAD_RESULT2=$(curl -s -X POST "${BASE_URL}/share/file" \
      -F "file=@test_date_grouping.txt" \
      -F "expire_value=1" \
      -F "expire_style=day")
    echo "结果: $UPLOAD_RESULT2"
fi

# 清理测试文件
rm -f test_date_grouping.txt

echo
echo "=== 测试完成 ==="
