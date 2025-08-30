#!/bin/bash

# FileCodeBox Go版本网页功能测试脚本

BASE_URL="http://localhost:12345"

echo "=== FileCodeBox Go版本 网页功能测试 ==="
echo

# 1. 测试首页加载
echo "1. 测试首页加载..."
HOME_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/homepage.html "$BASE_URL")
if [ "$HOME_RESPONSE" = "200" ]; then
    echo "✅ 首页加载成功 (HTTP 200)"
    # 检查关键元素
    if grep -q "FileCodeBox" /tmp/homepage.html; then
        echo "✅ 页面标题正确"
    else
        echo "❌ 页面标题缺失"
    fi
    
    if grep -q "文件分享" /tmp/homepage.html; then
        echo "✅ 文件分享标签存在"
    else
        echo "❌ 文件分享标签缺失"
    fi
    
    if grep -q "文本分享" /tmp/homepage.html; then
        echo "✅ 文本分享标签存在"
    else
        echo "❌ 文本分享标签缺失"
    fi
    
    if grep -q "获取分享" /tmp/homepage.html; then
        echo "✅ 获取分享标签存在"
    else
        echo "❌ 获取分享标签缺失"
    fi
else
    echo "❌ 首页加载失败 (HTTP $HOME_RESPONSE)"
fi
echo

# 2. 测试静态资源
echo "2. 测试静态资源..."
ASSETS_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/assets/")
if [ "$ASSETS_RESPONSE" = "200" ] || [ "$ASSETS_RESPONSE" = "404" ]; then
    echo "✅ 静态资源路径可访问"
else
    echo "❌ 静态资源路径异常 (HTTP $ASSETS_RESPONSE)"
fi
echo

# 3. 测试配置API（网页会调用）
echo "3. 测试配置API..."
CONFIG_RESPONSE=$(curl -s -X POST "$BASE_URL/" -w "%{http_code}")
if echo "$CONFIG_RESPONSE" | grep -q "200"; then
    echo "✅ 配置API响应正常"
    
    # 检查配置字段
    if echo "$CONFIG_RESPONSE" | grep -q "uploadSize"; then
        echo "✅ 上传大小配置存在"
    else
        echo "❌ 上传大小配置缺失"
    fi
    
    if echo "$CONFIG_RESPONSE" | grep -q "expireStyle"; then
        echo "✅ 过期样式配置存在"
    else
        echo "❌ 过期样式配置缺失"
    fi
    
    if echo "$CONFIG_RESPONSE" | grep -q "enableChunk"; then
        echo "✅ 分片上传配置存在"
    else
        echo "❌ 分片上传配置缺失"
    fi
else
    echo "❌ 配置API响应异常"
fi
echo

# 4. 测试文本分享功能
echo "4. 测试文本分享功能..."
TEXT_RESPONSE=$(curl -s -X POST "$BASE_URL/share/text/" \
  -F "text=网页测试文本内容" \
  -F "expire_value=1" \
  -F "expire_style=hour")

if echo "$TEXT_RESPONSE" | grep -q '"code":200'; then
    echo "✅ 文本分享功能正常"
    TEXT_CODE=$(echo "$TEXT_RESPONSE" | sed -n 's/.*"code":"\([^"]*\)".*/\1/p')
    echo "   提取码: $TEXT_CODE"
    
    # 测试获取文本
    if [ ! -z "$TEXT_CODE" ]; then
        GET_RESPONSE=$(curl -s -X POST "$BASE_URL/share/select/" \
          -H "Content-Type: application/json" \
          -d "{\"code\":\"$TEXT_CODE\"}")
        
        if echo "$GET_RESPONSE" | grep -q "网页测试文本内容"; then
            echo "✅ 文本获取功能正常"
        else
            echo "❌ 文本获取功能异常"
            echo "   响应: $GET_RESPONSE"
        fi
    fi
else
    echo "❌ 文本分享功能异常"
    echo "   响应: $TEXT_RESPONSE"
fi
echo

# 5. 测试文件上传功能
echo "5. 测试文件上传功能..."
echo "这是网页测试文件内容" > web_test_file.txt

FILE_RESPONSE=$(curl -s -X POST "$BASE_URL/share/file/" \
  -F "file=@web_test_file.txt" \
  -F "expire_value=1" \
  -F "expire_style=hour")

if echo "$FILE_RESPONSE" | grep -q '"code":200'; then
    echo "✅ 文件上传功能正常"
    FILE_CODE=$(echo "$FILE_RESPONSE" | sed -n 's/.*"code":"\([^"]*\)".*/\1/p')
    echo "   提取码: $FILE_CODE"
    
    # 测试获取文件信息
    if [ ! -z "$FILE_CODE" ]; then
        FILE_GET_RESPONSE=$(curl -s -X POST "$BASE_URL/share/select/" \
          -H "Content-Type: application/json" \
          -d "{\"code\":\"$FILE_CODE\"}")
        
        if echo "$FILE_GET_RESPONSE" | grep -q "web_test_file.txt"; then
            echo "✅ 文件信息获取正常"
        else
            echo "❌ 文件信息获取异常"
            echo "   响应: $FILE_GET_RESPONSE"
        fi
        
        # 测试文件下载
        DOWNLOAD_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/downloaded_file.txt "$BASE_URL/share/download?code=$FILE_CODE")
        if [ "$DOWNLOAD_RESPONSE" = "200" ]; then
            echo "✅ 文件下载功能正常"
            if grep -q "这是网页测试文件内容" /tmp/downloaded_file.txt; then
                echo "✅ 下载文件内容正确"
            else
                echo "❌ 下载文件内容不正确"
            fi
        else
            echo "❌ 文件下载功能异常 (HTTP $DOWNLOAD_RESPONSE)"
        fi
    fi
else
    echo "❌ 文件上传功能异常"
    echo "   响应: $FILE_RESPONSE"
fi
echo

# 6. 测试robots.txt
echo "6. 测试robots.txt..."
ROBOTS_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/robots.txt "$BASE_URL/robots.txt")
if [ "$ROBOTS_RESPONSE" = "200" ]; then
    echo "✅ robots.txt 可访问"
    if grep -q "User-agent" /tmp/robots.txt; then
        echo "✅ robots.txt 内容正确"
    else
        echo "❌ robots.txt 内容异常"
    fi
else
    echo "❌ robots.txt 访问失败 (HTTP $ROBOTS_RESPONSE)"
fi
echo

# 7. 测试错误处理
echo "7. 测试错误处理..."
ERROR_RESPONSE=$(curl -s -X POST "$BASE_URL/share/select/" \
  -H "Content-Type: application/json" \
  -d "{\"code\":\"nonexistent123\"}")

if echo "$ERROR_RESPONSE" | grep -q '"code":404'; then
    echo "✅ 错误处理正常（不存在的代码返回404）"
else
    echo "❌ 错误处理异常"
    echo "   响应: $ERROR_RESPONSE"
fi
echo

# 8. 测试参数验证
echo "8. 测试参数验证..."
INVALID_RESPONSE=$(curl -s -X POST "$BASE_URL/share/text/" \
  -F "expire_value=-1" \
  -F "expire_style=invalid")

if echo "$INVALID_RESPONSE" | grep -q '"code":400'; then
    echo "✅ 参数验证正常（无效参数返回400）"
else
    echo "❌ 参数验证异常"
    echo "   响应: $INVALID_RESPONSE"
fi
echo

# 清理测试文件
rm -f web_test_file.txt /tmp/homepage.html /tmp/robots.txt /tmp/downloaded_file.txt

echo "=== 网页功能测试完成 ==="
echo
echo "建议在浏览器中手动测试以下功能:"
echo "1. 文件拖拽上传"
echo "2. 标签页切换"
echo "3. 表单验证"
echo "4. 响应式布局"
echo "5. JavaScript交互"
