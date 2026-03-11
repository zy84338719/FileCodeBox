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
    -ldflags="-X 'github.com/zy84338719/fileCodeBox/internal/models/service.Version=${VERSION}' \
              -X 'github.com/zy84338719/fileCodeBox/internal/models/service.GitCommit=${COMMIT}' \
              -X 'github.com/zy84338719/fileCodeBox/internal/models/service.BuildTime=${BUILD_TIME}' \
              -w -s" \
    -o bin/server ./cmd/server

# Stage 3: Runtime
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    mkdir -p /app/data /app/static /app/config

# Copy backend binary from builder
COPY --from=backend-builder /app/bin/server ./server

# Copy frontend dist to static directory
COPY --from=frontend-builder /frontend/dist ./static/

# Copy config template
COPY backend/configs ./config/

# Expose port
EXPOSE 8888

# Set timezone
ENV TZ=Asia/Shanghai

# Run the server
CMD ["./server", "--config", "./config/config.yaml"]
