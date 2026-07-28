# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

# Copy frontend package files
COPY frontend/package.json frontend/package-lock.json* ./

# Install frontend dependencies
RUN npm ci

# Copy frontend source code
COPY frontend/ ./

# Build frontend
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go module files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./

# Build backend binary
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X 'github.com/zy84338719/fileCodeBox/backend/internal/models/service.Version=${VERSION}' \
              -X 'github.com/zy84338719/fileCodeBox/backend/internal/models/service.GitCommit=${COMMIT}' \
              -X 'github.com/zy84338719/fileCodeBox/backend/internal/models/service.BuildTime=${BUILD_TIME}' \
              -w -s" \
    -o bin/server ./cmd/server

# Stage 3: Runtime
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies（wget 用于 HEALTHCHECK）
RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -g 1000 app && \
    adduser -D -s /bin/sh -u 1000 -G app app && \
    mkdir -p /app/data /app/static /app/config && \
    chown -R app:app /app

# Copy backend binary from builder
COPY --from=backend-builder /app/bin/server ./server

# Copy frontend dist to static directory
COPY --from=frontend-builder /frontend/dist ./static/

# Copy config template
COPY backend/configs ./config/

# 非 root 运行
USER app

# Expose port（与默认配置 12345 一致）
EXPOSE 12345

# Set timezone
ENV TZ=Asia/Shanghai

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:12345/live || exit 1

# Run the server（--config 指定配置文件，亦可被 CONFIG_PATH env 覆盖）
CMD ["./server", "--config", "./config/config.yaml"]
