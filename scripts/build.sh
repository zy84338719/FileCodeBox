#!/bin/bash

# FileCodeBox 构建脚本
# 该脚本会自动获取版本信息并在编译时注入

set -e

# 定义颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_info "开始构建 FileCodeBox..."

# 获取版本信息 - 优先使用 Git tag
VERSION="dev"
if git describe --tags --exact-match HEAD >/dev/null 2>&1; then
    # 当前 commit 有确切的 tag
    VERSION=$(git describe --tags --exact-match HEAD)
    print_info "使用当前 commit 的 tag: $VERSION"
elif git describe --tags --abbrev=0 >/dev/null 2>&1; then
    # 使用最近的 tag 加上 commit 信息
    LATEST_TAG=$(git describe --tags --abbrev=0)
    COMMITS_SINCE_TAG=$(git rev-list --count ${LATEST_TAG}..HEAD)
    if [ "$COMMITS_SINCE_TAG" -gt 0 ]; then
        SHORT_COMMIT=$(git rev-parse --short HEAD)
        VERSION="${LATEST_TAG}-${COMMITS_SINCE_TAG}-g${SHORT_COMMIT}"
        print_info "使用最近的 tag 加提交数: $VERSION"
    else
        VERSION="$LATEST_TAG"
        print_info "使用最近的 tag: $VERSION"
    fi
else
    print_warning "未找到 Git tags，使用默认版本: $VERSION"
fi

# 获取 Git 提交哈希
GIT_COMMIT="unknown"
if git rev-parse --short HEAD >/dev/null 2>&1; then
    GIT_COMMIT=$(git rev-parse --short HEAD)
    # 检查是否有未提交的更改
    if ! git diff-index --quiet HEAD --; then
        GIT_COMMIT="${GIT_COMMIT}-dirty"
    fi
fi

# 获取构建时间
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

print_info "版本信息:"
print_info "  Version:    $VERSION"
print_info "  GitCommit:  $GIT_COMMIT"
print_info "  BuildTime:  $BUILD_TIME"

# 定义 ldflags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X 'github.com/zy84338719/filecodebox/internal/models/service.Version=$VERSION'"
LDFLAGS="$LDFLAGS -X 'github.com/zy84338719/filecodebox/internal/models/service.GitCommit=$GIT_COMMIT'"
LDFLAGS="$LDFLAGS -X 'github.com/zy84338719/filecodebox/internal/models/service.BuildTime=$BUILD_TIME'"

# 输出目录
OUTPUT_DIR="build"
mkdir -p "$OUTPUT_DIR"

# 构建二进制文件
BINARY_NAME="filecodebox"

# 检查是否指定了目标平台
if [ -n "$GOOS" ] && [ -n "$GOARCH" ]; then
    BINARY_NAME="${BINARY_NAME}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
    print_info "交叉编译目标: $GOOS/$GOARCH"
fi

OUTPUT_PATH="$OUTPUT_DIR/$BINARY_NAME"

print_info "开始编译..."
go build -ldflags "$LDFLAGS" -o "$OUTPUT_PATH" .

if [ $? -eq 0 ]; then
    print_success "构建完成: $OUTPUT_PATH"
    
    # 显示文件信息
    if command -v ls >/dev/null 2>&1; then
        ls -lh "$OUTPUT_PATH"
    fi
    
    # 如果是当前平台的构建，显示版本信息
    if [ -z "$GOOS" ] || [ "$GOOS" = "$(go env GOOS)" ]; then
        if [ -z "$GOARCH" ] || [ "$GOARCH" = "$(go env GOARCH)" ]; then
            print_info "验证版本信息:"
            "$OUTPUT_PATH" -version
        fi
    fi
else
    print_error "构建失败"
    exit 1
fi
