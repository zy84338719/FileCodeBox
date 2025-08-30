#!/bin/bash

# 快速发布脚本 - 简化版本
# 用于快速提交代码和推送到GitHub

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}FileCodeBox 快速发布工具${NC}"
echo "=================================="

# 检查Git状态
echo -e "${BLUE}📋 检查Git状态...${NC}"
git status

# 获取提交信息
if [[ -n "$1" ]]; then
    COMMIT_MSG="$1"
else
    echo ""
    echo -e "${YELLOW}💬 请输入提交信息:${NC}"
    read -r COMMIT_MSG
fi

if [[ -z "$COMMIT_MSG" ]]; then
    COMMIT_MSG="更新代码 $(date '+%Y-%m-%d %H:%M:%S')"
fi

echo ""
echo -e "${BLUE}📝 提交信息: ${NC}$COMMIT_MSG"

# 确认操作
echo ""
echo -e "${YELLOW}🔍 即将执行的操作:${NC}"
echo "1. git add ."
echo "2. git commit -m \"$COMMIT_MSG\""
echo "3. git push origin main"

echo ""
read -p "是否继续? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    echo "操作已取消"
    exit 0
fi

# 执行Git操作
echo ""
echo -e "${BLUE}📦 添加文件到暂存区...${NC}"
git add .

echo -e "${BLUE}💾 提交更改...${NC}"
git commit -m "$COMMIT_MSG"

echo -e "${BLUE}🚀 推送到GitHub...${NC}"
git push origin main

echo ""
echo -e "${GREEN}✅ 发布完成!${NC}"
echo -e "${GREEN}🎉 代码已成功推送到GitHub${NC}"

# 显示最新的提交信息
echo ""
echo -e "${BLUE}📋 最新提交信息:${NC}"
git log --oneline -1
