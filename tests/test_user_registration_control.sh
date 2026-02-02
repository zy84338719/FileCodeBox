#!/bin/bash

# 测试用户注册禁用功能
# 验证 issue #20 的修复

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_TOKEN=""

echo "========================================="
echo "测试用户注册控制功能"
echo "========================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试结果统计
PASSED=0
FAILED=0

# 测试函数
test_case() {
    local test_name="$1"
    local expected="$2"
    local actual="$3"
    
    if [ "$expected" = "$actual" ]; then
        echo -e "${GREEN}✓ PASSED${NC}: $test_name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC}: $test_name"
        echo "  Expected: $expected"
        echo "  Actual: $actual"
        FAILED=$((FAILED + 1))
    fi
}

# 1. 获取系统信息，检查注册状态
echo ""
echo "1. 检查系统信息 API"
SYSTEM_INFO=$(curl -s "$BASE_URL/user/system-info")
echo "系统信息: $SYSTEM_INFO"

# 解析 allow_user_registration 字段
ALLOW_REG=$(echo "$SYSTEM_INFO" | grep -o '"allow_user_registration":[0-9]' | cut -d':' -f2)
echo "当前注册状态: $ALLOW_REG (0=禁止, 1=允许)"

# 2. 测试当注册被禁用时
echo ""
echo "2. 测试注册被禁用的情况"
if [ "$ALLOW_REG" = "0" ]; then
    echo "注册当前已禁用，测试访问注册 API..."
    
    # 尝试访问注册页面
    echo "  a. 访问注册页面 /user/register"
    REG_PAGE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/user/register")
    test_case "注册页面应返回 404" "404" "$REG_PAGE_STATUS"
    
    # 尝试调用注册 API
    echo "  b. 调用注册 API POST /user/register"
    REG_API_RESPONSE=$(curl -s -X POST "$BASE_URL/user/register" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser",
            "email": "test@example.com",
            "password": "test123456"
        }')
    
    # 检查响应状态码
    REG_API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/user/register" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser",
            "email": "test@example.com",
            "password": "test123456"
        }')
    
    echo "  注册 API 响应: $REG_API_RESPONSE"
    test_case "注册 API 应返回 403 或 404" "40[34]" "$(echo $REG_API_STATUS | grep -o '40[34]' || echo 'wrong')"
    
    # 检查响应消息
    if echo "$REG_API_RESPONSE" | grep -q "不允许\|禁用\|forbidden"; then
        echo -e "${GREEN}✓ PASSED${NC}: 注册 API 返回了正确的错误消息"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC}: 注册 API 未返回预期的错误消息"
        FAILED=$((FAILED + 1))
    fi
    
else
    echo "注册当前已启用，测试访问注册功能..."
    
    # 访问注册页面应该成功
    echo "  a. 访问注册页面 /user/register"
    REG_PAGE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/user/register")
    test_case "注册页面应返回 200" "200" "$REG_PAGE_STATUS"
    
    # 注册 API 应该可用
    echo "  b. 注册 API 应该可用（不测试实际注册，只检查端点存在）"
    echo -e "${YELLOW}⚠ INFO${NC}: 注册功能已启用，跳过详细测试以避免创建测试数据"
fi

# 3. 检查前端注册按钮显示逻辑
echo ""
echo "3. 检查前端首页注册按钮"
INDEX_PAGE=$(curl -s "$BASE_URL/")

if [ "$ALLOW_REG" = "0" ]; then
    echo "  注册禁用时，前端应通过 JS 动态隐藏注册按钮"
    echo -e "${YELLOW}⚠ INFO${NC}: 前端逻辑通过 auth.js 动态控制，需要浏览器环境测试"
else
    echo "  注册启用时，前端应显示注册按钮"
    echo -e "${YELLOW}⚠ INFO${NC}: 前端逻辑通过 auth.js 动态控制，需要浏览器环境测试"
fi

# 4. 测试登录功能不受影响
echo ""
echo "4. 验证登录功能正常（不受注册禁用影响）"
LOGIN_PAGE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/user/login")
test_case "登录页面应该可访问" "200" "$LOGIN_PAGE_STATUS"

# 输出测试总结
echo ""
echo "========================================="
echo "测试总结"
echo "========================================="
echo -e "${GREEN}通过: $PASSED${NC}"
echo -e "${RED}失败: $FAILED${NC}"
echo "总计: $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}所有测试通过！${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}有测试失败，请检查！${NC}"
    exit 1
fi
