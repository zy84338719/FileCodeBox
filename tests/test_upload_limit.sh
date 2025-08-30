#!/bin/bash

# 测试上传大小限制功能

echo "=== 测试上传大小限制功能 ==="

# 1. 获取当前配置
echo "1. 获取当前配置:"
curl -s -H "Authorization: Bearer aaaaaa50124" http://localhost:12345/admin/config | jq '.detail.upload_size'

# 2. 设置上传限制为5MB进行测试
echo -e "\n2. 设置上传限制为5MB:"
curl -s -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer aaaaaa50124" \
  -d '{"upload_size": 5242880}' \
  http://localhost:12345/admin/config | jq '.message'

# 3. 确认配置已更新
echo -e "\n3. 确认配置已更新:"
curl -s -H "Authorization: Bearer aaaaaa50124" http://localhost:12345/admin/config | jq '.detail.upload_size'

# 4. 创建一个6MB的测试文件（超过限制）
echo -e "\n4. 创建6MB测试文件（超过5MB限制）:"
dd if=/dev/zero of=test_6mb.bin bs=1024 count=6144 2>/dev/null
echo "创建了6MB测试文件"

# 5. 测试上传6MB文件（应该失败）
echo -e "\n5. 测试上传6MB文件（应该失败）:"
curl -s -F "file=@test_6mb.bin" -F "expire_value=1" -F "expire_style=day" \
  http://localhost:12345/share/file/ | jq '.message'

# 6. 创建一个3MB的测试文件（在限制内）
echo -e "\n6. 创建3MB测试文件（在5MB限制内）:"
dd if=/dev/zero of=test_3mb.bin bs=1024 count=3072 2>/dev/null
echo "创建了3MB测试文件"

# 7. 测试上传3MB文件（应该成功）
echo -e "\n7. 测试上传3MB文件（应该成功）:"
curl -s -F "file=@test_3mb.bin" -F "expire_value=1" -F "expire_style=day" \
  http://localhost:12345/share/file/ | jq '.message'

# 8. 清理测试文件
echo -e "\n8. 清理测试文件:"
rm -f test_6mb.bin test_3mb.bin
echo "测试文件已清理"

echo -e "\n=== 测试完成 ==="
