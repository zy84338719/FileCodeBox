#!/bin/bash

# FileCodeBox Go版本 JavaScript功能测试

BASE_URL="http://localhost:12345"

echo "=== FileCodeBox JavaScript功能测试 ==="
echo

# 1. 检查HTML页面中的JavaScript代码
echo "1. 检查HTML页面中的JavaScript代码..."
curl -s "$BASE_URL/" > /tmp/test_page.html

if grep -q "switchTab" /tmp/test_page.html; then
    echo "✅ 标签切换函数存在"
else
    echo "❌ 标签切换函数缺失"
fi

if grep -q "addEventListener" /tmp/test_page.html; then
    echo "✅ 事件监听器存在"
else
    echo "❌ 事件监听器缺失"
fi

if grep -q "fetch" /tmp/test_page.html; then
    echo "✅ Fetch API调用存在"
else
    echo "❌ Fetch API调用缺失"
fi

if grep -q "FormData" /tmp/test_page.html; then
    echo "✅ FormData处理存在"
else
    echo "❌ FormData处理缺失"
fi

if grep -q "dragover" /tmp/test_page.html; then
    echo "✅ 拖拽事件处理存在"
else
    echo "❌ 拖拽事件处理缺失"
fi

echo

# 2. 检查CSS样式
echo "2. 检查CSS样式..."
if grep -q "\.tab\.active" /tmp/test_page.html; then
    echo "✅ 活动标签样式存在"
else
    echo "❌ 活动标签样式缺失"
fi

if grep -q "\.tab-content\.active" /tmp/test_page.html; then
    echo "✅ 活动内容样式存在"
else
    echo "❌ 活动内容样式缺失"
fi

if grep -q "\.result\.show" /tmp/test_page.html; then
    echo "✅ 结果显示样式存在"
else
    echo "❌ 结果显示样式缺失"
fi

if grep -q "@media" /tmp/test_page.html; then
    echo "✅ 响应式样式存在"
else
    echo "❌ 响应式样式缺失"
fi

echo

# 3. 模拟前端请求测试
echo "3. 模拟前端AJAX请求测试..."

# 测试跨域请求
CORS_RESPONSE=$(curl -s -H "Origin: http://example.com" -H "Access-Control-Request-Method: POST" -H "Access-Control-Request-Headers: Content-Type" -X OPTIONS "$BASE_URL/share/text/")
echo "CORS预检请求响应码: $(curl -s -w "%{http_code}" -o /dev/null -H "Origin: http://example.com" -H "Access-Control-Request-Method: POST" -H "Access-Control-Request-Headers: Content-Type" -X OPTIONS "$BASE_URL/share/text/")"

# 测试JSON请求
JSON_RESPONSE=$(curl -s -X POST "$BASE_URL/share/select/" \
  -H "Content-Type: application/json" \
  -H "Origin: http://localhost:12345" \
  -d '{"code":"test123"}')

if echo "$JSON_RESPONSE" | grep -q "文件不存在"; then
    echo "✅ JSON请求处理正常"
else
    echo "❌ JSON请求处理异常"
fi

echo

# 4. 检查HTML模板变量替换
echo "4. 检查HTML模板变量替换..."
if grep -q "{{" /tmp/test_page.html; then
    echo "❌ 模板变量未完全替换:"
    grep "{{" /tmp/test_page.html | head -3
else
    echo "✅ 模板变量替换完整"
fi

echo

# 5. 检查表单字段
echo "5. 检查表单字段..."
FORM_FIELDS=("file-input" "text-form" "get-form" "expire_style" "expire_value")

for field in "${FORM_FIELDS[@]}"; do
    if grep -q "$field" /tmp/test_page.html; then
        echo "✅ 字段 $field 存在"
    else
        echo "❌ 字段 $field 缺失"
    fi
done

echo

# 6. 检查文件验证
echo "6. 测试文件大小验证..."
# 创建一个超大文件（模拟）
LARGE_FILE_RESPONSE=$(curl -s -X POST "$BASE_URL/share/file/" \
  -F "file=@/dev/null" \
  -F "expire_value=1" \
  -F "expire_style=hour" 2>/dev/null || echo "请求失败")

echo "大文件上传响应: $LARGE_FILE_RESPONSE"

echo

# 7. 检查页面元素结构
echo "7. 检查页面元素结构..."
REQUIRED_ELEMENTS=("container" "header" "logo" "tabs" "upload-area" "form-group" "btn")

for element in "${REQUIRED_ELEMENTS[@]}"; do
    if grep -q "class=\"[^\"]*$element" /tmp/test_page.html; then
        echo "✅ 元素 .$element 存在"
    else
        echo "❌ 元素 .$element 缺失"
    fi
done

echo

# 8. 检查错误处理
echo "8. 检查前端错误处理..."
if grep -q "alert" /tmp/test_page.html; then
    echo "✅ 错误提示处理存在"
else
    echo "❌ 错误提示处理缺失"
fi

if grep -q "catch" /tmp/test_page.html; then
    echo "✅ 异常捕获处理存在"
else
    echo "❌ 异常捕获处理缺失"
fi

# 清理
rm -f /tmp/test_page.html download_test.txt

echo
echo "=== JavaScript功能测试完成 ==="
echo
echo "建议进行的手动测试:"
echo "1. 在浏览器开发者工具中检查控制台错误"
echo "2. 测试文件拖拽到上传区域"
echo "3. 测试标签页点击切换"
echo "4. 测试表单提交和响应显示"
echo "5. 测试移动端响应式布局"
echo "6. 测试大文件上传的进度显示"
