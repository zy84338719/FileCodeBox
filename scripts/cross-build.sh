#!/bin/bash

# FileCodeBox 交叉编译脚本
# 用于构建多平台二进制文件

set -e

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$PROJECT_ROOT"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 获取版本信息
VERSION="1.0.0"
if [ -f "VERSION" ]; then
    VERSION=$(cat VERSION | tr -d '\n')
elif git describe --tags --abbrev=0 >/dev/null 2>&1; then
    VERSION=$(git describe --tags --abbrev=0)
fi

GIT_COMMIT="unknown"
if git rev-parse --short HEAD >/dev/null 2>&1; then
    GIT_COMMIT=$(git rev-parse --short HEAD)
    if ! git diff-index --quiet HEAD --; then
        GIT_COMMIT="${GIT_COMMIT}-dirty"
    fi
fi

BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

print_info "开始交叉编译 FileCodeBox v$VERSION"
print_info "Git Commit: $GIT_COMMIT"
print_info "Build Time: $BUILD_TIME"

# 定义构建目标
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/386"
    "linux/arm"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/386"
    "freebsd/amd64"
    "freebsd/arm64"
)

# 创建输出目录
OUTPUT_DIR="dist"
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# 定义 ldflags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X 'github.com/zy84338719/filecodebox/internal/models/service.Version=$VERSION'"
LDFLAGS="$LDFLAGS -X 'github.com/zy84338719/filecodebox/internal/models/service.GitCommit=$GIT_COMMIT'"
LDFLAGS="$LDFLAGS -X 'github.com/zy84338719/filecodebox/internal/models/service.BuildTime=$BUILD_TIME'"

# 构建每个平台
FAILED_BUILDS=()

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    
    print_info "构建 $os/$arch..."
    
    binary_name="filecodebox-${VERSION}-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi
    
    output_path="$OUTPUT_DIR/$binary_name"
    
    # 设置环境变量并构建
    env GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o "$output_path" \
        . 2>/dev/null
    
    if [ $? -eq 0 ]; then
        # 获取文件大小
        if command -v stat >/dev/null 2>&1; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                size=$(stat -f%z "$output_path")
            else
                size=$(stat -c%s "$output_path")
            fi
            size_mb=$(echo "scale=2; $size / 1024 / 1024" | bc 2>/dev/null || echo "unknown")
            print_success "$os/$arch 构建完成 (${size_mb}MB)"
        else
            print_success "$os/$arch 构建完成"
        fi
    else
        print_error "$os/$arch 构建失败"
        FAILED_BUILDS+=("$platform")
        rm -f "$output_path"
    fi
done

# 创建校验和文件
if command -v sha256sum >/dev/null 2>&1; then
    print_info "生成 SHA256 校验和..."
    cd "$OUTPUT_DIR"
    sha256sum * > SHA256SUMS
    cd ..
elif command -v shasum >/dev/null 2>&1; then
    print_info "生成 SHA256 校验和..."
    cd "$OUTPUT_DIR"
    shasum -a 256 * > SHA256SUMS
    cd ..
fi

# 统计结果
total_platforms=${#PLATFORMS[@]}
failed_count=${#FAILED_BUILDS[@]}
success_count=$((total_platforms - failed_count))

echo
print_info "构建完成统计:"
print_info "  成功: $success_count/$total_platforms"

if [ ${#FAILED_BUILDS[@]} -gt 0 ]; then
    print_error "  失败的平台:"
    for platform in "${FAILED_BUILDS[@]}"; do
        print_error "    - $platform"
    done
fi

print_info "  输出目录: $OUTPUT_DIR"

# 显示输出文件
if [ "$(ls -1 "$OUTPUT_DIR" | wc -l)" -gt 0 ]; then
    echo
    print_info "构建的文件:"
    ls -lh "$OUTPUT_DIR"
fi

if [ ${#FAILED_BUILDS[@]} -eq 0 ]; then
    print_success "所有平台构建成功!"
    exit 0
else
    print_error "部分平台构建失败"
    exit 1
fi
