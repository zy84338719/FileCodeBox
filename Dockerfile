# 多阶段构建：第一阶段用于编译
FROM golang:1.22-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev \
    git \
    ca-certificates

WORKDIR /app

# 复制依赖文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译应用程序
RUN CGO_ENABLED=1 go build -ldflags="-w -s" -o filecodebox .

# 第二阶段：运行时镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    sqlite

# 创建非root用户
RUN addgroup -g 1000 app && \
    adduser -D -s /bin/sh -u 1000 -G app app

WORKDIR /app

# 从构建阶段复制文件
COPY --from=builder /app/filecodebox .
COPY --from=builder /app/themes ./themes

# 创建数据目录并设置权限
RUN mkdir -p data && chown -R app:app /app

# 切换到非root用户
USER app

EXPOSE 12345

CMD ["./filecodebox"]
