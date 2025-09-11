#!/bin/bash

# 多平台构建脚本
# 用于本地测试构建或手动发布

set -e

# 版本信息
VERSION=${1:-"dev-$(git rev-parse --short HEAD)"}
COMMIT=$(git rev-parse HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# 构建标志
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# 创建构建目录
BUILD_DIR="build"
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

echo "🚀 开始构建 FileCodeBox ${VERSION}"
echo "📅 构建时间: ${DATE}"
echo "📝 提交哈希: ${COMMIT}"
echo ""

# 定义平台列表
declare -a platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64" 
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

# 构建每个平台
for platform in "${platforms[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"
    
    # 确定输出文件名
    output_name="filecodebox-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "🔨 构建 ${GOOS}/${GOARCH} -> ${output_name}"
    
    # 构建二进制文件
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="${LDFLAGS}" \
        -o "${BUILD_DIR}/${output_name}" .
    
    # 设置执行权限 (非 Windows)
    if [ "$GOOS" != "windows" ]; then
        chmod +x "${BUILD_DIR}/${output_name}"
    fi
    
    # 创建发布包目录
    package_dir="${BUILD_DIR}/package-${GOOS}-${GOARCH}"
    mkdir -p "$package_dir"
    
    # 复制可执行文件
    cp "${BUILD_DIR}/${output_name}" "$package_dir/"
    
    # 创建 README
    cat > "$package_dir/README.txt" << EOF
FileCodeBox - 文件分享服务

平台: ${GOOS}/${GOARCH}
版本: ${VERSION}
构建时间: ${DATE}
Git 提交: ${COMMIT}

使用方法:
1. 运行可执行文件启动服务
2. 访问 http://localhost:12345
3. 管理员访问 http://localhost:12345/admin
4. 默认管理员密码: FileCodeBox2025

配置文件会在首次运行时自动创建。

更多信息: https://github.com/zy84338719/FileCodeBox
EOF
    
    # 创建启动脚本 (非 Windows)
    if [ "$GOOS" != "windows" ]; then
        cat > "$package_dir/start.sh" << 'EOF'
#!/bin/bash
echo "🚀 启动 FileCodeBox..."
echo "📱 用户界面: http://localhost:12345"
echo "⚙️ 管理界面: http://localhost:12345/admin"
echo "🔑 默认密码: FileCodeBox2025"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""
./$(basename "$package_dir" | sed 's/filecodebox-/filecodebox-/' | cut -d'-' -f1-3)
EOF
        chmod +x "$package_dir/start.sh"
    else
        # Windows 批处理文件
        cat > "$package_dir/start.bat" << 'EOF'
@echo off
echo 🚀 启动 FileCodeBox...
echo 📱 用户界面: http://localhost:12345
echo ⚙️ 管理界面: http://localhost:12345/admin
echo 🔑 默认密码: FileCodeBox2025
echo.
echo 按 Ctrl+C 停止服务
echo.
filecodebox-windows-amd64.exe
pause
EOF
    fi
    
    # 打包
    cd "${BUILD_DIR}"
    package_name="filecodebox-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        zip -r "${package_name}.zip" "package-${GOOS}-${GOARCH}/"
        echo "📦 已创建: ${package_name}.zip"
    else
        tar -czf "${package_name}.tar.gz" "package-${GOOS}-${GOARCH}/"
        echo "📦 已创建: ${package_name}.tar.gz"
    fi
    cd ..
    
    # 清理临时目录
    rm -rf "$package_dir"
    
    echo "✅ ${GOOS}/${GOARCH} 构建完成"
    echo ""
done

echo "🎉 所有平台构建完成！"
echo ""
echo "📁 构建文件位置: ${BUILD_DIR}/"
ls -lh ${BUILD_DIR}/

echo ""
echo "📋 构建摘要:"
echo "版本: ${VERSION}"
echo "平台数量: ${#platforms[@]}"
echo "构建时间: ${DATE}"

# 计算文件大小
total_size=$(du -sh ${BUILD_DIR} | cut -f1)
echo "总大小: ${total_size}"

echo ""
echo "🚀 测试本地构建:"
echo "  ./build/filecodebox-$(go env GOOS)-$(go env GOARCH)$(if [ '$(go env GOOS)' = 'windows' ]; then echo '.exe'; fi)"
