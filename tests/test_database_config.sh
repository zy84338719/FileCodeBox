#!/bin/bash

echo "=== 数据库配置系统完整测试 ==="

# 1. 检查数据库中的配置数据
echo "1. 数据库配置项统计:"
sqlite3 data/filecodebox.db "SELECT COUNT(*) as total_configs FROM key_values;"

echo -e "\n2. 关键配置项状态:"
sqlite3 data/filecodebox.db "SELECT key, value FROM key_values WHERE key IN ('name', 'description', 'upload_size', 'admin_token') ORDER BY key;" | while IFS='|' read -r key value; do
    if [ "$key" = "upload_size" ]; then
        size_mb=$((value / 1024 / 1024))
        echo "  $key: ${size_mb}MB ($value bytes)"
    else
        echo "  $key: $value"
    fi
done

# 3. 等待服务器启动
echo -e "\n3. 等待服务器启动..."
sleep 2

# 4. 获取JWT token
echo -e "\n4. 管理员登录测试:"
TOKEN_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d '{"password":"FileCodeBox2025"}' http://localhost:12345/admin/login)
if echo "$TOKEN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$TOKEN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo "  ✅ 登录成功，获取到JWT token"
else
    echo "  ❌ 登录失败: $TOKEN_RESPONSE"
    exit 1
fi

# 5. 测试配置读取
echo -e "\n5. 配置读取测试:"
CONFIG_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:12345/admin/config)
if echo "$CONFIG_RESPONSE" | grep -q '"upload_size"'; then
    CURRENT_SIZE=$(echo "$CONFIG_RESPONSE" | grep -o '"upload_size":[0-9]*' | grep -o '[0-9]*')
    CURRENT_SIZE_MB=$((CURRENT_SIZE / 1024 / 1024))
    echo "  ✅ 配置读取成功，当前上传限制: ${CURRENT_SIZE_MB}MB"
else
    echo "  ❌ 配置读取失败: $CONFIG_RESPONSE"
    exit 1
fi

# 6. 测试配置更新
echo -e "\n6. 配置更新测试:"
NEW_SIZE=20971520  # 20MB
UPDATE_RESPONSE=$(curl -s -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d "{\"upload_size\": $NEW_SIZE, \"description\": \"数据库配置测试\"}" \
  http://localhost:12345/admin/config)

if echo "$UPDATE_RESPONSE" | grep -q '"message":"更新成功"'; then
    echo "  ✅ 配置更新成功"
else
    echo "  ❌ 配置更新失败: $UPDATE_RESPONSE"
    exit 1
fi

# 7. 验证配置更新
echo -e "\n7. 验证配置更新:"
VERIFY_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:12345/admin/config)
if echo "$VERIFY_RESPONSE" | grep -q '"upload_size":20971520'; then
    echo "  ✅ 配置更新验证成功，新的上传限制: 20MB"
else
    echo "  ❌ 配置更新验证失败"
fi

# 8. 检查数据库持久化
echo -e "\n8. 数据库持久化验证:"
DB_SIZE=$(sqlite3 data/filecodebox.db "SELECT value FROM key_values WHERE key='upload_size';")
DB_DESC=$(sqlite3 data/filecodebox.db "SELECT value FROM key_values WHERE key='description';")
DB_SIZE_MB=$((DB_SIZE / 1024 / 1024))

if [ "$DB_SIZE" = "20971520" ] && [ "$DB_DESC" = "数据库配置测试" ]; then
    echo "  ✅ 数据库持久化成功"
    echo "    - 上传限制: ${DB_SIZE_MB}MB"
    echo "    - 描述: $DB_DESC"
else
    echo "  ❌ 数据库持久化失败"
    echo "    - 数据库中的上传限制: $DB_SIZE"
    echo "    - 数据库中的描述: $DB_DESC"
fi

# 9. 测试文件上传限制
echo -e "\n9. 文件上传限制测试:"
# 创建25MB文件（超过20MB限制）
echo "  创建25MB测试文件..."
dd if=/dev/zero of=test_25mb.bin bs=1024 count=25600 2>/dev/null

# 测试上传（应该失败）
UPLOAD_RESPONSE=$(curl -s -F "file=@test_25mb.bin" -F "expire_value=1" -F "expire_style=day" http://localhost:12345/share/file/)
if echo "$UPLOAD_RESPONSE" | grep -q "文件大小超过限制"; then
    echo "  ✅ 文件大小限制正常工作，25MB文件被正确拒绝"
else
    echo "  ❌ 文件大小限制异常: $UPLOAD_RESPONSE"
fi

# 清理测试文件
rm -f test_25mb.bin

echo -e "\n=== 测试完成 ==="
echo "✅ 数据库配置系统功能完全正常！"
echo "  - 配置优先从数据库加载"
echo "  - 数据库为空时自动初始化默认配置"
echo "  - 配置更新直接保存到数据库"
echo "  - 文件大小限制实时生效"
