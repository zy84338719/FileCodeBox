#!/bin/bash

# FileCodeBox 多架构Docker构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函数：打印彩色输出
print_status() {
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

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed!"
    exit 1
fi

# 默认参数
IMAGE_NAME="filecodebox"
IMAGE_TAG="latest"
PLATFORMS="linux/amd64,linux/arm64"
PUSH=false
BUILD_SINGLE=false
TARGET_PLATFORM=""

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        --tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        --platforms)
            PLATFORMS="$2"
            shift 2
            ;;
        --push)
            PUSH=true
            shift
            ;;
        --single)
            BUILD_SINGLE=true
            TARGET_PLATFORM="$2"
            shift 2
            ;;
        --help|-h)
            echo "FileCodeBox Docker 多架构构建脚本"
            echo ""
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --name NAME          Docker镜像名称 (默认: filecodebox)"
            echo "  --tag TAG            Docker镜像标签 (默认: latest)"
            echo "  --platforms PLATFORMS 目标平台，逗号分隔 (默认: linux/amd64,linux/arm64)"
            echo "  --push               构建后推送到registry"
            echo "  --single PLATFORM    构建单一平台并加载到本地"
            echo "  --help, -h           显示此帮助信息"
            echo ""
            echo "示例:"
            echo "  $0                                    # 本地多架构构建"
            echo "  $0 --push                            # 构建并推送"
            echo "  $0 --single linux/amd64              # 构建AMD64并加载到本地"
            echo "  $0 --single linux/arm64              # 构建ARM64并加载到本地"
            echo "  $0 --name myregistry/filecodebox --tag v1.0.0 --push"
            exit 0
            ;;
        *)
            print_error "未知参数: $1"
            echo "使用 --help 查看帮助信息"
            exit 1
            ;;
    esac
done

FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

print_status "FileCodeBox Docker 多架构构建"
print_status "镜像名称: ${FULL_IMAGE_NAME}"

# 检查buildx是否可用
if ! docker buildx version &> /dev/null; then
    print_error "Docker buildx is not available!"
    print_status "尝试安装buildx..."
    docker buildx install
fi

# 创建并使用buildx builder
BUILDER_NAME="filecodebox-builder"
if ! docker buildx ls | grep -q "$BUILDER_NAME"; then
    print_status "创建buildx builder: $BUILDER_NAME"
    docker buildx create --name "$BUILDER_NAME" --driver docker-container --bootstrap
fi

print_status "使用buildx builder: $BUILDER_NAME"
docker buildx use "$BUILDER_NAME"

# 构建参数
BUILD_ARGS=""
if [ "$PUSH" = true ]; then
    BUILD_ARGS="--push"
elif [ "$BUILD_SINGLE" = true ]; then
    BUILD_ARGS="--load"
    PLATFORMS="$TARGET_PLATFORM"
    print_status "构建单一平台: $TARGET_PLATFORM"
else
    BUILD_ARGS="--load"
    # 如果不推送，只构建当前平台
    CURRENT_ARCH=$(docker version --format '{{.Server.Arch}}')
    case $CURRENT_ARCH in
        amd64)
            PLATFORMS="linux/amd64"
            ;;
        arm64)
            PLATFORMS="linux/arm64"
            ;;
        *)
            print_warning "未识别的架构: $CURRENT_ARCH，使用默认平台"
            PLATFORMS="linux/amd64"
            ;;
    esac
fi

print_status "目标平台: $PLATFORMS"

# 执行构建
print_status "开始构建..."
docker buildx build \
    --platform "$PLATFORMS" \
    --tag "$FULL_IMAGE_NAME" \
    $BUILD_ARGS \
    .

if [ $? -eq 0 ]; then
    print_success "构建完成!"
    
    if [ "$BUILD_SINGLE" = true ] || ([ "$PUSH" = false ] && [[ ! "$PLATFORMS" =~ "," ]]); then
        print_status "验证镜像..."
        if docker images | grep -q "$IMAGE_NAME.*$IMAGE_TAG"; then
            print_success "镜像已加载到本地Docker"
            
            # 显示镜像信息
            print_status "镜像信息:"
            docker images | head -1
            docker images | grep "$IMAGE_NAME.*$IMAGE_TAG"
            
            # 测试运行
            print_status "测试运行容器..."
            CONTAINER_ID=$(docker run -d -p 12347:12345 "$FULL_IMAGE_NAME")
            sleep 3
            
            if docker ps | grep -q "$CONTAINER_ID"; then
                print_success "容器运行成功! 容器ID: $CONTAINER_ID"
                print_status "访问地址: http://localhost:12347"
                print_status "停止容器: docker stop $CONTAINER_ID"
            else
                print_error "容器启动失败"
                docker logs "$CONTAINER_ID"
                docker rm "$CONTAINER_ID"
            fi
        fi
    elif [ "$PUSH" = true ]; then
        print_success "镜像已推送到registry"
    fi
else
    print_error "构建失败!"
    exit 1
fi
