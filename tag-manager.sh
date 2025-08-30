#!/bin/bash

# 标签管理脚本
# 用于管理Git标签的创建、删除和推送

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

show_help() {
    cat << EOF
Git标签管理工具

用法: $0 <命令> [选项]

命令:
    create <version>    创建新标签
    delete <version>    删除标签
    list               列出所有标签
    push <version>     推送标签到远程
    pull               拉取远程标签
    show <version>     显示标签详情

选项:
    -f, --force        强制执行操作
    -m, --message      标签描述信息

示例:
    $0 create v1.0.0                    # 创建v1.0.0标签
    $0 create v1.0.1 -m "修复重要bug"    # 创建带描述的标签
    $0 delete v1.0.0                    # 删除v1.0.0标签
    $0 push v1.0.0                      # 推送v1.0.0标签
    $0 list                             # 列出所有标签

EOF
}

# 创建标签
create_tag() {
    local version=$1
    local message=$2
    local force=$3
    
    if [[ -z "$version" ]]; then
        echo -e "${RED}错误: 请提供版本号${NC}"
        exit 1
    fi
    
    # 验证版本号格式
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
        echo -e "${RED}错误: 版本号格式不正确${NC}"
        echo "正确格式: v1.0.0, v1.2.3-beta, v2.0.0-rc1"
        exit 1
    fi
    
    # 检查标签是否已存在
    if git tag -l | grep -q "^$version$"; then
        if [[ "$force" == "true" ]]; then
            echo -e "${YELLOW}警告: 标签 $version 已存在，将强制覆盖${NC}"
            git tag -d "$version"
        else
            echo -e "${RED}错误: 标签 $version 已存在${NC}"
            echo "使用 -f 选项强制覆盖"
            exit 1
        fi
    fi
    
    # 默认标签消息
    if [[ -z "$message" ]]; then
        message="Release $version"
    fi
    
    echo -e "${BLUE}创建标签: $version${NC}"
    echo -e "${BLUE}描述信息: $message${NC}"
    
    # 创建标签
    git tag -a "$version" -m "$message"
    echo -e "${GREEN}✅ 标签 $version 创建成功${NC}"
    
    # 询问是否推送
    read -p "是否推送到远程仓库? (Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        git push origin "$version"
        echo -e "${GREEN}✅ 标签已推送到远程仓库${NC}"
    fi
}

# 删除标签
delete_tag() {
    local version=$1
    local force=$2
    
    if [[ -z "$version" ]]; then
        echo -e "${RED}错误: 请提供版本号${NC}"
        exit 1
    fi
    
    # 检查标签是否存在
    if ! git tag -l | grep -q "^$version$"; then
        echo -e "${RED}错误: 标签 $version 不存在${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}警告: 即将删除标签 $version${NC}"
    
    if [[ "$force" != "true" ]]; then
        read -p "确认删除? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "操作已取消"
            exit 0
        fi
    fi
    
    # 删除本地标签
    git tag -d "$version"
    echo -e "${GREEN}✅ 本地标签 $version 已删除${NC}"
    
    # 删除远程标签
    read -p "是否同时删除远程标签? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git push origin ":refs/tags/$version"
        echo -e "${GREEN}✅ 远程标签 $version 已删除${NC}"
    fi
}

# 列出标签
list_tags() {
    echo -e "${BLUE}📋 所有标签:${NC}"
    if git tag -l | wc -l | grep -q "^0$"; then
        echo "暂无标签"
    else
        git tag -l --sort=-version:refname | head -20
        
        local total=$(git tag -l | wc -l)
        if [[ $total -gt 20 ]]; then
            echo "..."
            echo -e "${YELLOW}(显示最新20个标签，总共 $total 个)${NC}"
        fi
    fi
}

# 推送标签
push_tag() {
    local version=$1
    
    if [[ -z "$version" ]]; then
        echo -e "${RED}错误: 请提供版本号${NC}"
        exit 1
    fi
    
    if [[ "$version" == "all" ]]; then
        echo -e "${BLUE}推送所有标签到远程仓库...${NC}"
        git push origin --tags
        echo -e "${GREEN}✅ 所有标签已推送${NC}"
    else
        # 检查标签是否存在
        if ! git tag -l | grep -q "^$version$"; then
            echo -e "${RED}错误: 标签 $version 不存在${NC}"
            exit 1
        fi
        
        echo -e "${BLUE}推送标签 $version 到远程仓库...${NC}"
        git push origin "$version"
        echo -e "${GREEN}✅ 标签 $version 已推送${NC}"
    fi
}

# 拉取远程标签
pull_tags() {
    echo -e "${BLUE}拉取远程标签...${NC}"
    git fetch origin --tags
    echo -e "${GREEN}✅ 远程标签已同步${NC}"
}

# 显示标签详情
show_tag() {
    local version=$1
    
    if [[ -z "$version" ]]; then
        echo -e "${RED}错误: 请提供版本号${NC}"
        exit 1
    fi
    
    if ! git tag -l | grep -q "^$version$"; then
        echo -e "${RED}错误: 标签 $version 不存在${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}标签详情: $version${NC}"
    echo "=================================="
    git show "$version"
}

# 主函数
main() {
    local command=$1
    local force=false
    local message=""
    
    # 解析选项
    shift
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--force)
                force=true
                shift
                ;;
            -m|--message)
                message="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                version="$1"
                shift
                ;;
        esac
    done
    
    # 执行命令
    case $command in
        create)
            create_tag "$version" "$message" "$force"
            ;;
        delete)
            delete_tag "$version" "$force"
            ;;
        list)
            list_tags
            ;;
        push)
            push_tag "$version"
            ;;
        pull)
            pull_tags
            ;;
        show)
            show_tag "$version"
            ;;
        *)
            echo -e "${RED}错误: 未知命令 '$command'${NC}"
            show_help
            exit 1
            ;;
    esac
}

# 检查参数
if [[ $# -eq 0 ]]; then
    show_help
    exit 1
fi

# 检查是否在Git仓库中
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}错误: 当前目录不是Git仓库${NC}"
    exit 1
fi

# 执行主函数
main "$@"
