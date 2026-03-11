#!/bin/bash

# FileCodeBox Go版本发布脚本
# 用于自动化版本管理、代码提交和标签发布

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
FileCodeBox Go版本发布脚本

用法: $0 [选项] <版本号>

选项:
    -h, --help          显示此帮助信息
    -d, --dry-run       模拟运行，不实际执行操作
    -f, --force         强制执行，跳过确认
    -m, --message       自定义提交信息
    -p, --pre-release   标记为预发布版本
    -b, --build         发布前构建Docker镜像
    
版本号格式: v1.0.0, v1.2.3-beta, v2.0.0-rc1

示例:
    $0 v1.0.0                    # 发布v1.0.0版本
    $0 v1.1.0-beta --pre-release # 发布v1.1.0-beta预发布版本
    $0 v1.0.1 -m "修复重要bug"    # 自定义提交信息
    $0 v1.0.0 --dry-run         # 模拟运行
    $0 v1.0.0 --build           # 发布前构建Docker镜像

EOF
}

# 检查Git仓库状态
check_git_status() {
    log_info "检查Git仓库状态..."
    
    # 检查是否在Git仓库中
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "当前目录不是Git仓库"
        exit 1
    fi
    
    # 检查是否有未提交的更改
    if ! git diff-index --quiet HEAD --; then
        log_warning "发现未提交的更改:"
        git status --porcelain
        if [[ "$FORCE" != "true" ]]; then
            read -p "是否继续? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_info "操作已取消"
                exit 0
            fi
        fi
    fi
    
    # 检查当前分支
    current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" ]]; then
        log_warning "当前分支: $current_branch (建议在main分支发布)"
        if [[ "$FORCE" != "true" ]]; then
            read -p "是否继续? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_info "操作已取消"
                exit 0
            fi
        fi
    fi
}

# 验证版本号格式
validate_version() {
    local version=$1
    
    # 版本号格式验证 (支持 v1.0.0, v1.0.0-beta, v1.0.0-rc1 等)
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
        log_error "版本号格式错误: $version"
        log_info "正确格式: v1.0.0, v1.2.3-beta, v2.0.0-rc1"
        exit 1
    fi
    
    # 检查标签是否已存在
    if git tag -l | grep -q "^$version$"; then
        log_error "标签 $version 已存在"
        if [[ "$FORCE" != "true" ]]; then
            log_info "使用 --force 强制覆盖现有标签"
            exit 1
        else
            log_warning "将覆盖现有标签: $version"
        fi
    fi
}

# 更新版本信息
update_version_info() {
    local version=$1

    log_info "版本信息由 Git tag 提供，跳过文件内版本更新 ($version)"
}

# 运行测试
run_tests() {
    log_info "运行测试..."
    
    if [[ -f "go.mod" ]]; then
        # Go项目测试
        if ! go test ./...; then
            log_error "测试失败"
            exit 1
        fi
        log_success "Go测试通过"
    fi
    
    # 运行自定义测试脚本
    if [[ -f "tests/run_all_tests.sh" ]]; then
        log_info "运行自定义测试脚本..."
        if ! bash tests/run_all_tests.sh; then
            log_error "自定义测试失败"
            exit 1
        fi
        log_success "自定义测试通过"
    fi
}

# 构建项目
build_project() {
    log_info "构建项目..."
    
    # Go项目构建
    if [[ -f "go.mod" ]]; then
        local version_value="${VERSION#v}"
        local git_commit="unknown"
        if git rev-parse --short HEAD >/dev/null 2>&1; then
            git_commit=$(git rev-parse --short HEAD)
            if ! git diff-index --quiet HEAD --; then
                git_commit="${git_commit}-dirty"
            fi
        fi
        local build_time=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

        local ldflags="-w -s"
        ldflags="$ldflags -X 'github.com/zy84338719/fileCodeBox/backend/internal/models/service.Version=$version_value'"
        ldflags="$ldflags -X 'github.com/zy84338719/fileCodeBox/backend/internal/models/service.GitCommit=$git_commit'"
        ldflags="$ldflags -X 'github.com/zy84338719/fileCodeBox/backend/internal/models/service.BuildTime=$build_time'"

        if ! go build -ldflags="$ldflags" -o filecodebox .; then
            log_error "构建失败"
            exit 1
        fi
        log_success "Go项目构建成功"
    fi
    
    # Docker镜像构建
    if [[ "$BUILD_DOCKER" == "true" ]]; then
        log_info "构建Docker镜像..."
        if [[ -f "Dockerfile" ]]; then
            if ! docker build -t "filecodebox:$VERSION" .; then
                log_error "Docker镜像构建失败"
                exit 1
            fi
            log_success "Docker镜像构建成功"
            
            # 多架构构建
            if [[ -f "build-docker.sh" ]]; then
                log_info "构建多架构Docker镜像..."
                if ! ./build-docker.sh; then
                    log_warning "多架构Docker镜像构建失败"
                else
                    log_success "多架构Docker镜像构建成功"
                fi
            fi
        fi
    fi
}

# 生成变更日志
generate_changelog() {
    local version=$1
    local last_tag
    
    log_info "生成变更日志..."
    
    # 获取上一个标签
    last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    
    if [[ -n "$last_tag" ]]; then
        log_info "从 $last_tag 到 $version 的变更:"
        git log --oneline --pretty=format:"- %s" "$last_tag"..HEAD
    else
        log_info "首次发布，显示最近10个提交:"
        git log --oneline --pretty=format:"- %s" -10
    fi
    
    echo ""
}

# 提交和推送代码
commit_and_push() {
    local version=$1
    local commit_message="$COMMIT_MESSAGE"
    
    if [[ -z "$commit_message" ]]; then
        commit_message="Release $version

🚀 版本发布: $version

自动生成的发布提交"
    fi
    
    log_info "提交更改..."
    
    # 添加所有更改
    git add .
    
    # 检查是否有更改需要提交
    if git diff --staged --quiet; then
        log_info "没有更改需要提交"
    else
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "[模拟] 将提交: $commit_message"
        else
            git commit -m "$commit_message"
            log_success "代码已提交"
        fi
    fi
    
    # 推送到远程仓库
    log_info "推送到远程仓库..."
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[模拟] 将推送到 origin main"
    else
        if ! git push origin main; then
            log_error "推送失败"
            exit 1
        fi
        log_success "代码已推送到远程仓库"
    fi
}

# 创建和推送标签
create_and_push_tag() {
    local version=$1
    local tag_message="Release $version"
    
    if [[ "$PRE_RELEASE" == "true" ]]; then
        tag_message="$tag_message (Pre-release)"
    fi
    
    log_info "创建标签 $version..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[模拟] 将创建标签: $version"
        log_info "[模拟] 标签信息: $tag_message"
    else
        # 创建带注释的标签
        if [[ "$FORCE" == "true" ]] && git tag -l | grep -q "^$version$"; then
            git tag -d "$version"  # 删除本地标签
            git push origin ":refs/tags/$version" 2>/dev/null || true  # 删除远程标签
        fi
        
        git tag -a "$version" -m "$tag_message"
        log_success "标签 $version 已创建"
        
        # 推送标签
        log_info "推送标签到远程仓库..."
        if ! git push origin "$version"; then
            log_error "标签推送失败"
            exit 1
        fi
        log_success "标签已推送到远程仓库"
    fi
}

# 显示发布信息
show_release_info() {
    local version=$1
    
    echo ""
    echo "======================================"
    log_success "发布完成! 🎉"
    echo "======================================"
    echo "版本: $version"
    echo "分支: $(git branch --show-current)"
    echo "提交: $(git rev-parse --short HEAD)"
    echo "远程仓库: $(git remote get-url origin)"
    echo ""
    
    if [[ "$DRY_RUN" != "true" ]]; then
        echo "标签链接: $(git remote get-url origin)/releases/tag/$version"
        echo ""
        log_info "下一步操作建议:"
        echo "1. 在GitHub上编辑Release说明"
        if [[ "$BUILD_DOCKER" == "true" ]]; then
            echo "2. 推送Docker镜像到仓库"
            echo "   docker push your-registry/filecodebox:$version"
        fi
        echo "3. 通知团队新版本发布"
    fi
}

# 主函数
main() {
    # 默认值
    DRY_RUN=false
    FORCE=false
    PRE_RELEASE=false
    BUILD_DOCKER=false
    COMMIT_MESSAGE=""
    VERSION=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -p|--pre-release)
                PRE_RELEASE=true
                shift
                ;;
            -b|--build)
                BUILD_DOCKER=true
                shift
                ;;
            -m|--message)
                COMMIT_MESSAGE="$2"
                shift 2
                ;;
            v*)
                VERSION="$1"
                shift
                ;;
            *)
                log_error "未知选项: $1"
                echo "使用 $0 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    # 检查版本号参数
    if [[ -z "$VERSION" ]]; then
        log_error "请提供版本号"
        echo "使用 $0 --help 查看帮助"
        exit 1
    fi
    
    # 显示运行模式
    if [[ "$DRY_RUN" == "true" ]]; then
        log_warning "模拟运行模式 - 不会执行实际操作"
    fi
    
    log_info "开始发布流程: $VERSION"
    
    # 执行发布流程
    check_git_status
    validate_version "$VERSION"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        update_version_info "$VERSION"
        run_tests
        build_project
    fi
    
    generate_changelog "$VERSION"
    commit_and_push "$VERSION"
    create_and_push_tag "$VERSION"
    show_release_info "$VERSION"
}

# 脚本入口
main "$@"
