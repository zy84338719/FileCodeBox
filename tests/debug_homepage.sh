#!/bin/bash

echo "=== FileCodeBox 主页调试报告 ==="
echo "时间: $(date)"
echo

echo "1. 服务器状态检查:"
if curl -s --connect-timeout 5 "http://0.0.0.0:12345/" >/dev/null; then
    echo "✅ 服务器正在运行"
else
    echo "❌ 服务器连接失败"
    exit 1
fi

echo
echo "2. 主页响应检查:"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://0.0.0.0:12345/")
echo "HTTP状态码: $HTTP_CODE"

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ HTTP响应正常"
else
    echo "❌ HTTP响应异常"
fi

echo
echo "3. 内容长度检查:"
CONTENT_LENGTH=$(curl -s "http://0.0.0.0:12345/" | wc -c)
echo "内容长度: $CONTENT_LENGTH 字节"

if [ "$CONTENT_LENGTH" -gt 1000 ]; then
    echo "✅ 内容长度正常"
else
    echo "❌ 内容长度异常"
fi

echo
echo "4. HTML结构检查:"
if curl -s "http://0.0.0.0:12345/" | grep -q "<html"; then
    echo "✅ 包含HTML标签"
else
    echo "❌ 缺少HTML标签"
fi

if curl -s "http://0.0.0.0:12345/" | grep -q "<body"; then
    echo "✅ 包含body标签"
else
    echo "❌ 缺少body标签"
fi

if curl -s "http://0.0.0.0:12345/" | grep -q "container"; then
    echo "✅ 包含container元素"
else
    echo "❌ 缺少container元素"
fi

echo
echo "5. 静态资源检查:"
for resource in css/base.css js/main.js assets/images/logo.svg; do
    if curl -s --head "http://0.0.0.0:12345/$resource" | grep -q "200 OK"; then
        echo "✅ $resource 可访问"
    else
        echo "❌ $resource 不可访问"
    fi
done

echo
echo "6. 内容样本 (前10行):"
curl -s "http://0.0.0.0:12345/" | head -10

echo
echo "7. API配置检查:"
if curl -s "http://0.0.0.0:12345/api/config" | grep -q '"code":200'; then
    echo "✅ API配置正常"
else
    echo "❌ API配置异常"
fi

echo
echo "=== 调试完成 ==="