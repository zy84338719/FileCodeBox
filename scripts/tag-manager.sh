#!/bin/bash

# 标签管理脚本 v2.0
# 用于管理Git标签的创建、删除和推送
# 支持语义化版本控制和自动化发布

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m'

# 版本信息
SCRIPT_VERSION="2.0.0"

# 全局变量
TAG_PREFIX="${TAG_PREFIX:-v}"
AUTO_PUSH="${AUTO_PUSH:-false}"
DEFAULT_BUMP="${DEFAULT_BUMP:-patch}"
VERBOSE=false
QUIET=false
DRY_RUN=false
AUTO_GENERATE_NOTES=false
PRE_RELEASE=false

# 工具函数
log_info() {
    [[ "$QUIET" == "true" ]] && return
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    [[ "$QUIET" == "true" ]] && return
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}" >&2
}

log_error() {
    echo -e "${RED}❌ $1${NC}" >&2
}

log_verbose() {
    [[ "$VERBOSE" == "true" ]] && echo -e "${CYAN}🔍 $1${NC}"
}

# 执行命令（支持干运行模式）
execute_cmd() {
    local cmd="$1"
    local description="$2"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "${PURPLE}[DRY RUN]${NC} $description"
        echo -e "${CYAN}Command:${NC} $cmd"
        return 0
    fi
    
    log_verbose "执行: $cmd"
    if eval "$cmd"; then
        return 0
    else
        local exit_code=$?
        log_error "命令执行失败: $cmd"
        return $exit_code
    fi
}

# 获取最新标签
get_latest_tag() {
    git tag -l --sort=-version:refname | grep "^${TAG_PREFIX}" | head -1 || echo ""
}

# 解析版本号
parse_version() {
    local version="$1"
    # 移除前缀
    version=${version#$TAG_PREFIX}
    # 移除预发布标识
    version=${version%%-*}
    
    if [[ $version =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        echo "${BASH_REMATCH[1]} ${BASH_REMATCH[2]} ${BASH_REMATCH[3]}"
    else
        echo ""
    fi
}

# 验证版本号格式
validate_version() {
    local version="$1"
    
    if [[ -z "$version" ]]; then
        log_error "版本号不能为空"
        return 1
    fi
    
    # 检查是否以指定前缀开头
    if [[ ! $version =~ ^${TAG_PREFIX}[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
        log_error "版本号格式不正确"
        echo "正确格式: ${TAG_PREFIX}1.0.0, ${TAG_PREFIX}1.2.3-beta, ${TAG_PREFIX}2.0.0-rc.1"
        return 1
    fi
    
    return 0
}

# 递增版本号
bump_version() {
    local current="$1"
    local bump_type="$2"
    local pre_release="$3"
    
    if [[ -z "$current" ]]; then
        echo "${TAG_PREFIX}1.0.0"
        return
    fi
    
    local parsed
    parsed=$(parse_version "$current")
    if [[ -z "$parsed" ]]; then
        log_error "无法解析当前版本号: $current"
        exit 1
    fi
    
    read -r major minor patch <<< "$parsed"
    
    case "$bump_type" in
        major)
            ((major++))
            minor=0
            patch=0
            ;;
        minor)
            ((minor++))
            patch=0
            ;;
        patch)
            ((patch++))
            ;;
        *)
            log_error "未知的递增类型: $bump_type"
            exit 1
            ;;
    esac
    
    local new_version="${TAG_PREFIX}${major}.${minor}.${patch}"
    
    if [[ "$pre_release" == "true" ]]; then
        new_version="${new_version}-rc.1"
    fi
    
    echo "$new_version"
}

# 生成更新日志
generate_changelog() {
    local version="$1"
    local previous_version="$2"
    
    if [[ -z "$previous_version" ]]; then
        previous_version=$(get_latest_tag)
        if [[ -z "$previous_version" ]]; then
            log_warning "没有找到之前的版本，生成完整日志"
            git log --pretty=format:"- %s (%h)" --reverse
            return
        fi
    fi
    
    log_info "生成从 $previous_version 到 $version 的更新日志"
    
    echo "## $version"
    echo ""
    echo "### 🚀 新特性"
    git log "$previous_version..HEAD" --pretty=format:"- %s (%h)" --grep="feat\|新增\|添加\|新功能" || true
    echo ""
    echo "### 🐛 Bug修复"
    git log "$previous_version..HEAD" --pretty=format:"- %s (%h)" --grep="fix\|修复\|bugfix" || true
    echo ""
    echo "### 📝 文档更新"
    git log "$previous_version..HEAD" --pretty=format:"- %s (%h)" --grep="docs\|文档" || true
    echo ""
    echo "### 🔧 其他改进"
    git log "$previous_version..HEAD" --pretty=format:"- %s (%h)" --invert-grep --grep="feat\|fix\|docs\|新增\|修复\|文档" || true
    echo ""
}

show_help() {
    cat << EOF
${WHITE}Git标签管理工具 v${SCRIPT_VERSION}${NC}

${BLUE}用法:${NC} $0 <命令> [选项]

${BLUE}命令:${NC}
    ${GREEN}create${NC} <version>       创建新标签
    ${GREEN}delete${NC} <version>       删除标签
    ${GREEN}list${NC}                  列出所有标签
    ${GREEN}push${NC} <version>        推送标签到远程
    ${GREEN}pull${NC}                  拉取远程标签
    ${GREEN}show${NC} <version>        显示标签详情
    ${GREEN}latest${NC}                显示最新标签
    ${GREEN}bump${NC} <type>           自动递增版本号 (major|minor|patch)
    ${GREEN}changelog${NC} [version]   生成更新日志
    ${GREEN}compare${NC} <v1> [v2]     比较两个版本的差异
    ${GREEN}validate${NC} <version>    验证版本号格式
    ${GREEN}auto${NC}                  自动创建下一个版本

${BLUE}选项:${NC}
    ${CYAN}-f, --force${NC}           强制执行操作
    ${CYAN}-m, --message${NC}         标签描述信息
    ${CYAN}-p, --push${NC}            创建后自动推送
    ${CYAN}-d, --dry-run${NC}         显示将要执行的操作但不实际执行
    ${CYAN}-q, --quiet${NC}           静默模式
    ${CYAN}-v, --verbose${NC}         详细输出
    ${CYAN}--pre-release${NC}         标记为预发布版本
    ${CYAN}--release-notes${NC}       自动生成发布说明

${BLUE}示例:${NC}
    $0 create v1.0.0                       # 创建v1.0.0标签
    $0 create v1.0.1 -m "修复重要bug" -p    # 创建并推送标签
    $0 bump patch -m "修复bug"              # 自动递增补丁版本
    $0 bump minor --pre-release            # 创建预发布版本
    $0 delete v1.0.0                       # 删除v1.0.0标签
    $0 compare v1.0.0 v1.1.0               # 比较两个版本
    $0 changelog v1.1.0                    # 生成v1.1.0的更新日志
    $0 auto -m "自动发布"                   # 自动创建下一版本

${BLUE}环境变量:${NC}
    ${CYAN}TAG_PREFIX${NC}            标签前缀 (默认: v)
    ${CYAN}AUTO_PUSH${NC}             自动推送 (true/false)
    ${CYAN}DEFAULT_BUMP${NC}          默认递增类型 (major/minor/patch)

EOF
}

# 创建标签
create_tag() {
    local version="$1"
    local message="$2"
    local force="$3"
    local auto_push="$4"
    
    if [[ -z "$version" ]]; then
        log_error "请提供版本号"
        exit 1
    fi
    
    # 验证版本号格式
    if ! validate_version "$version"; then
        exit 1
    fi
    
    # 检查标签是否已存在
    if git tag -l | grep -q "^$version$"; then
        if [[ "$force" == "true" ]]; then
            log_warning "标签 $version 已存在，将强制覆盖"
            execute_cmd "git tag -d '$version'" "删除现有本地标签"
            # 尝试删除远程标签
            if git ls-remote --tags origin | grep -q "refs/tags/$version"; then
                execute_cmd "git push origin ':refs/tags/$version'" "删除远程标签"
            fi
        else
            log_error "标签 $version 已存在"
            echo "使用 -f 选项强制覆盖，或使用 'bump' 命令自动生成新版本"
            exit 1
        fi
    fi
    
    # 默认标签消息
    if [[ -z "$message" ]]; then
        if [[ "$AUTO_GENERATE_NOTES" == "true" ]]; then
            local latest_tag
            latest_tag=$(get_latest_tag)
            message="Release $version"$'\n\n'"$(generate_changelog "$version" "$latest_tag")"
        else
            message="Release $version"
        fi
    fi
    
    log_info "创建标签: $version"
    log_verbose "描述信息: $message"
    
    # 创建标签
    execute_cmd "git tag -a '$version' -m '$message'" "创建标签"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "标签 $version 创建成功"
    fi
    
    # 自动推送或询问推送
    if [[ "$auto_push" == "true" || "$AUTO_PUSH" == "true" ]]; then
        push_tag "$version"
    elif [[ "$DRY_RUN" != "true" ]]; then
        read -p "是否推送到远程仓库? (Y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            push_tag "$version"
        fi
    fi
}

# 自动创建版本
auto_create() {
    local bump_type="${DEFAULT_BUMP}"
    local message="$1"
    local force="$2"
    local auto_push="$3"
    
    local latest_tag
    latest_tag=$(get_latest_tag)
    
    if [[ -z "$latest_tag" ]]; then
        log_info "没有找到现有标签，创建初始版本"
        latest_tag="${TAG_PREFIX}0.0.0"
    fi
    
    log_info "当前最新版本: $latest_tag"
    
    local new_version
    new_version=$(bump_version "$latest_tag" "$bump_type" "$PRE_RELEASE")
    
    log_info "自动生成新版本: $new_version"
    
    create_tag "$new_version" "$message" "$force" "$auto_push"
}

# 递增版本
bump_tag() {
    local bump_type="$1"
    local message="$2"
    local force="$3"
    local auto_push="$4"
    
    if [[ ! "$bump_type" =~ ^(major|minor|patch)$ ]]; then
        log_error "递增类型必须是: major, minor, 或 patch"
        exit 1
    fi
    
    local latest_tag
    latest_tag=$(get_latest_tag)
    
    if [[ -z "$latest_tag" ]]; then
        log_info "没有找到现有标签，创建初始版本"
        latest_tag="${TAG_PREFIX}0.0.0"
    fi
    
    log_info "当前最新版本: $latest_tag"
    
    local new_version
    new_version=$(bump_version "$latest_tag" "$bump_type" "$PRE_RELEASE")
    
    log_info "递增 $bump_type 版本: $latest_tag → $new_version"
    
    if [[ -z "$message" ]]; then
        case "$bump_type" in
            major) message="Major release $new_version - 重大更新" ;;
            minor) message="Minor release $new_version - 新功能" ;;
            patch) message="Patch release $new_version - Bug修复" ;;
        esac
    fi
    
    create_tag "$new_version" "$message" "$force" "$auto_push"
}

# 显示最新标签
show_latest() {
    local latest_tag
    latest_tag=$(get_latest_tag)
    
    if [[ -z "$latest_tag" ]]; then
        log_warning "没有找到任何标签"
        return 1
    fi
    
    echo -e "${WHITE}📋 最新标签:${NC} ${GREEN}$latest_tag${NC}"
    
    # 显示标签详情
    if [[ "$VERBOSE" == "true" ]]; then
        echo ""
        show_tag "$latest_tag"
    fi
}

# 比较版本
compare_versions() {
    local version1="$1"
    local version2="$2"
    
    if [[ -z "$version1" ]]; then
        log_error "请提供第一个版本号"
        exit 1
    fi
    
    if [[ -z "$version2" ]]; then
        version2="HEAD"
        log_info "比较 $version1 与当前HEAD"
    else
        log_info "比较 $version1 与 $version2"
    fi
    
    echo -e "${WHITE}📊 版本比较: $version1 ↔ $version2${NC}"
    echo "=================================="
    
    # 提交数量统计
    local commit_count
    if [[ "$version2" == "HEAD" ]]; then
        commit_count=$(git rev-list --count "$version1..HEAD" 2>/dev/null || echo "0")
    else
        commit_count=$(git rev-list --count "$version1..$version2" 2>/dev/null || echo "0")
    fi
    
    echo -e "${CYAN}提交数量:${NC} $commit_count"
    
    # 文件变更统计
    echo -e "${CYAN}文件变更:${NC}"
    if [[ "$version2" == "HEAD" ]]; then
        git diff --stat "$version1..HEAD" 2>/dev/null || echo "无变更"
    else
        git diff --stat "$version1..$version2" 2>/dev/null || echo "无变更"
    fi
    
    echo ""
    echo -e "${CYAN}详细提交记录:${NC}"
    if [[ "$version2" == "HEAD" ]]; then
        git log --oneline "$version1..HEAD" 2>/dev/null || echo "无新提交"
    else
        git log --oneline "$version1..$version2" 2>/dev/null || echo "无新提交"
    fi
}

# 验证版本号命令
validate_version_cmd() {
    local version="$1"
    
    if validate_version "$version"; then
        log_success "版本号格式正确: $version"
        
        local parsed
        parsed=$(parse_version "$version")
        if [[ -n "$parsed" ]]; then
            read -r major minor patch <<< "$parsed"
            echo -e "${CYAN}解析结果:${NC}"
            echo -e "  主版本: $major"
            echo -e "  次版本: $minor"
            echo -e "  补丁版本: $patch"
        fi
    else
        exit 1
    fi
}

# 删除标签
delete_tag() {
    local version="$1"
    local force="$2"
    
    if [[ -z "$version" ]]; then
        log_error "请提供版本号"
        exit 1
    fi
    
    # 检查标签是否存在
    if ! git tag -l | grep -q "^$version$"; then
        log_error "标签 $version 不存在"
        exit 1
    fi
    
    log_warning "即将删除标签 $version"
    
    if [[ "$force" != "true" && "$DRY_RUN" != "true" ]]; then
        read -p "确认删除? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "操作已取消"
            exit 0
        fi
    fi
    
    # 删除本地标签
    execute_cmd "git tag -d '$version'" "删除本地标签"
    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "本地标签 $version 已删除"
    fi
    
    # 检查并删除远程标签
    if git ls-remote --tags origin | grep -q "refs/tags/$version"; then
        if [[ "$DRY_RUN" != "true" ]]; then
            read -p "是否同时删除远程标签? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                execute_cmd "git push origin ':refs/tags/$version'" "删除远程标签"
                log_success "远程标签 $version 已删除"
            fi
        else
            execute_cmd "git push origin ':refs/tags/$version'" "删除远程标签"
        fi
    fi
}

# 列出标签
list_tags() {
    local show_all="$1"
    
    log_info "📋 标签列表:"
    
    if git tag -l | wc -l | grep -q "^0$"; then
        echo "暂无标签"
        return 0
    fi
    
    if [[ "$show_all" == "true" || "$VERBOSE" == "true" ]]; then
        # 显示所有标签，带详细信息
        echo -e "${WHITE}标签${NC} ${CYAN}创建时间${NC} ${YELLOW}提交${NC} ${GREEN}描述${NC}"
        echo "================================================================"
        git tag -l --sort=-version:refname | while read -r tag; do
            local date
            local commit
            local subject
            date=$(git log -1 --format=%ai "$tag" 2>/dev/null | cut -d' ' -f1)
            commit=$(git rev-parse --short "$tag" 2>/dev/null)
            subject=$(git tag -l --format='%(contents:subject)' "$tag" 2>/dev/null)
            printf "%-15s %-12s %-8s %s\n" "$tag" "$date" "$commit" "$subject"
        done
    else
        # 简化显示
        git tag -l --sort=-version:refname | head -20
        
        local total
        total=$(git tag -l | wc -l | tr -d ' ')
        if [[ $total -gt 20 ]]; then
            echo "..."
            echo -e "${YELLOW}(显示最新20个标签，总共 $total 个，使用 -v 查看详细信息)${NC}"
        fi
    fi
}

# 推送标签
push_tag() {
    local version="$1"
    
    if [[ -z "$version" ]]; then
        log_error "请提供版本号"
        exit 1
    fi
    
    if [[ "$version" == "all" ]]; then
        log_info "推送所有标签到远程仓库..."
        execute_cmd "git push origin --tags" "推送所有标签"
        if [[ "$DRY_RUN" != "true" ]]; then
            log_success "所有标签已推送"
        fi
    else
        # 检查标签是否存在
        if ! git tag -l | grep -q "^$version$"; then
            log_error "标签 $version 不存在"
            exit 1
        fi
        
        log_info "推送标签 $version 到远程仓库..."
        execute_cmd "git push origin '$version'" "推送标签"
        if [[ "$DRY_RUN" != "true" ]]; then
            log_success "标签 $version 已推送"
        fi
    fi
}

# 拉取远程标签
pull_tags() {
    log_info "拉取远程标签..."
    execute_cmd "git fetch origin --tags" "拉取远程标签"
    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "远程标签已同步"
        
        # 显示新拉取的标签
        if [[ "$VERBOSE" == "true" ]]; then
            echo ""
            list_tags
        fi
    fi
}

# 显示标签详情
show_tag() {
    local version="$1"
    
    if [[ -z "$version" ]]; then
        log_error "请提供版本号"
        exit 1
    fi
    
    if ! git tag -l | grep -q "^$version$"; then
        log_error "标签 $version 不存在"
        exit 1
    fi
    
    echo -e "${WHITE}📋 标签详情: $version${NC}"
    echo "=================================="
    
    # 显示标签信息
    git show "$version" --no-patch --format=fuller
    
    echo ""
    echo -e "${CYAN}📈 统计信息:${NC}"
    
    # 获取上一个标签
    local prev_tag
    prev_tag=$(git tag -l --sort=-version:refname | grep -A1 "^$version$" | tail -1)
    
    if [[ -n "$prev_tag" && "$prev_tag" != "$version" ]]; then
        local commit_count
        commit_count=$(git rev-list --count "$prev_tag..$version" 2>/dev/null || echo "0")
        echo -e "  自 $prev_tag 以来的提交数: $commit_count"
        
        echo ""
        echo -e "${CYAN}📝 主要变更:${NC}"
        git log --oneline "$prev_tag..$version" 2>/dev/null | head -10
    fi
}

# 主函数
main() {
    local command=""
    local force=false
    local auto_push=false
    local message=""
    local version=""
    local version2=""
    local show_all=false
    
    # 首先检查全局选项
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            --version)
                echo "Tag Manager v${SCRIPT_VERSION}"
                exit 0
                ;;
            -*)
                # 保存选项，稍后处理
                break
                ;;
            *)
                # 第一个非选项参数是命令
                if [[ -z "$command" ]]; then
                    command="$1"
                    shift
                    break
                fi
                ;;
        esac
        shift
    done
    
    # 解析剩余选项和参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--force)
                force=true
                shift
                ;;
            -p|--push)
                auto_push=true
                shift
                ;;
            -m|--message)
                message="$2"
                shift 2
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -q|--quiet)
                QUIET=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -a|--all)
                show_all=true
                shift
                ;;
            --pre-release)
                PRE_RELEASE=true
                shift
                ;;
            --release-notes)
                AUTO_GENERATE_NOTES=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            --version)
                echo "Tag Manager v${SCRIPT_VERSION}"
                exit 0
                ;;
            *)
                if [[ -z "$version" ]]; then
                    version="$1"
                else
                    # 用于compare命令的第二个版本
                    version2="$1"
                fi
                shift
                ;;
        esac
    done
    
    # 如果没有提供命令，显示帮助
    if [[ -z "$command" ]]; then
        show_help
        exit 1
    fi
    
    # 执行命令
    case $command in
        create)
            create_tag "$version" "$message" "$force" "$auto_push"
            ;;
        delete)
            delete_tag "$version" "$force"
            ;;
        list)
            list_tags "$show_all"
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
        latest)
            show_latest
            ;;
        bump)
            if [[ -z "$version" ]]; then
                log_error "请指定递增类型: major, minor, 或 patch"
                exit 1
            fi
            bump_tag "$version" "$message" "$force" "$auto_push"
            ;;
        auto)
            auto_create "$message" "$force" "$auto_push"
            ;;
        changelog)
            if [[ -z "$version" ]]; then
                version=$(get_latest_tag)
                if [[ -z "$version" ]]; then
                    log_error "没有找到标签，请指定版本号"
                    exit 1
                fi
            fi
            generate_changelog "$version"
            ;;
        compare)
            compare_versions "$version" "$version2"
            ;;
        validate)
            validate_version_cmd "$version"
            ;;
        *)
            log_error "未知命令: '$command'"
            echo ""
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
    log_error "当前目录不是Git仓库"
    exit 1
fi

# 预解析关键选项
for arg in "$@"; do
    case "$arg" in
        --dry-run|-d)
            DRY_RUN=true
            ;;
        --quiet|-q)
            QUIET=true
            ;;
        --verbose|-v)
            VERBOSE=true
            ;;
    esac
done

# 保存第一个参数用于工作区状态检查
FIRST_COMMAND="$1"

# 检查工作区状态
if [[ -n "$(git status --porcelain)" ]] && [[ "$FIRST_COMMAND" == "create" || "$FIRST_COMMAND" == "bump" || "$FIRST_COMMAND" == "auto" ]]; then
    log_warning "工作区有未提交的更改"
    if [[ "$DRY_RUN" != "true" ]]; then
        read -p "是否继续? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "操作已取消"
            exit 0
        fi
    else
        log_info "干运行模式：忽略工作区状态检查"
    fi
fi

# 执行主函数
main "$@"
