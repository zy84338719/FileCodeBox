#!/bin/bash

# FileCodeBox Go版本性能基准测试

BASE_URL="http://localhost:12345"

echo "=== FileCodeBox Go版本 性能基准测试 ==="
echo

# 测试并发文本分享
echo "1. 并发文本分享测试 (10个并发请求)..."
start_time=$(date +%s.%N)

for i in {1..10}; do
    (
        curl -s -X POST "${BASE_URL}/share/text/" \
          -F "text=并发测试文本内容 $i" \
          -F "expire_value=1" \
          -F "expire_style=hour" > /tmp/concurrent_$i.log
    ) &
done

wait

end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
echo "并发文本分享完成，耗时: ${duration}s"

# 统计成功数量
success_count=0
for i in {1..10}; do
    if grep -q '"code":200' /tmp/concurrent_$i.log 2>/dev/null; then
        success_count=$((success_count + 1))
    fi
done
echo "成功: $success_count/10"
echo

# 测试并发文件上传
echo "2. 并发文件上传测试 (5个并发请求)..."
# 创建测试文件
for i in {1..5}; do
    echo "测试文件内容 $i $(date)" > test_file_$i.txt
done

start_time=$(date +%s.%N)

for i in {1..5}; do
    (
        curl -s -X POST "${BASE_URL}/share/file/" \
          -F "file=@test_file_$i.txt" \
          -F "expire_value=1" \
          -F "expire_style=hour" > /tmp/upload_$i.log
    ) &
done

wait

end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
echo "并发文件上传完成，耗时: ${duration}s"

# 统计成功数量
success_count=0
for i in {1..5}; do
    if grep -q '"code":200' /tmp/upload_$i.log 2>/dev/null; then
        success_count=$((success_count + 1))
    fi
done
echo "成功: $success_count/5"
echo

# 测试内存和CPU使用情况
echo "3. 系统资源使用情况..."
if command -v ps &> /dev/null; then
    PID=$(pgrep filecodebox)
    if [ ! -z "$PID" ]; then
        echo "FileCodeBox进程信息:"
        ps -p $PID -o pid,pcpu,pmem,rss,vsz,comm
    fi
fi
echo

# 测试响应时间
echo "4. 响应时间测试..."
echo "首页响应时间:"
curl -o /dev/null -s -w "连接时间: %{time_connect}s, 总时间: %{time_total}s\n" $BASE_URL

echo "配置API响应时间:"
curl -o /dev/null -s -w "连接时间: %{time_connect}s, 总时间: %{time_total}s\n" -X POST $BASE_URL/

echo "文本分享响应时间:"
curl -o /dev/null -s -w "连接时间: %{time_connect}s, 总时间: %{time_total}s\n" \
  -X POST "${BASE_URL}/share/text/" \
  -F "text=响应时间测试" \
  -F "expire_value=1" \
  -F "expire_style=minute"

# 清理测试文件
rm -f test_file_*.txt /tmp/concurrent_*.log /tmp/upload_*.log

echo
echo "=== 性能基准测试完成 ==="
