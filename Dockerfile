# 多阶段构建 FileCodeBox v2
# Stage 1: 构建前端
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

# 复制前端依赖文件
COPY frontend/package*.json ./

# 安装依赖
RUN npm ci

# 复制前端源代码
COPY frontend/ ./

# 构建前端
RUN npm run build

# Stage 2: 构建后端
FROM golang:1.25-alpine AS backend-builder

# 安装构建依赖
RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev \
    git \
    ca-certificates \
    tzdata

WORKDIR /app

# 复制后端依赖文件
COPY backend/go.mod backend/go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源代码
COPY backend/ ./

# 构建参数
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# 编译后端
RUN CGO_ENABLED=1 go build \
    -ldflags="-w -s \
    -X 'github.com/zy84338719/fileCodeBox/internal/models/service.Version=${VERSION}' \
    -X 'github.com/zy84338719/fileCodeBox/internal/models/service.GitCommit=${COMMIT}' \
    -X 'github.com/zy84338719/fileCodeBox/internal/models/service.BuildTime=${BUILD_TIME}'" \
    -o bin/server ./cmd/server

# Stage 3: 运行时镜像
FROM alpine:3.19

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    sqlite

# 创建非root用户
RUN addgroup -g 1000 app && \
    adduser -D -s /bin/sh -u 1000 -G app app

WORKDIR /app

# 从后端构建阶段复制二进制文件
COPY --from=backend-builder /app/bin/server ./

# 从前端构建阶段复制静态文件
COPY --from=frontend-builder /frontend/dist ./static/

# 创建数据目录并设置权限
RUN mkdir -p data configs && chown -R app:app /app

# 切换到非root用户
USER app

# 暴露端口
EXPOSE 12346

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:12346/health || exit 1

# 启动服务
CMD ["./server"]
