.PHONY: all build build-frontend build-backend copy-frontend run test clean docker-build docker-run docker-stop docker-logs dev release version deps tidy fmt vet check help

# 版本信息 - 优先使用 Git tag
VERSION ?= $(shell \
	if git describe --tags --exact-match HEAD >/dev/null 2>&1; then \
		git describe --tags --exact-match HEAD; \
	elif git describe --tags --abbrev=0 >/dev/null 2>&1; then \
		LATEST_TAG=$$(git describe --tags --abbrev=0); \
		COMMITS_SINCE_TAG=$$(git rev-list --count $${LATEST_TAG}..HEAD); \
		if [ "$${COMMITS_SINCE_TAG}" -gt 0 ]; then \
			SHORT_COMMIT=$$(git rev-parse --short HEAD); \
			echo "$${LATEST_TAG}-$${COMMITS_SINCE_TAG}-g$${SHORT_COMMIT}"; \
		else \
			echo "$${LATEST_TAG}"; \
		fi; \
	else \
		echo "dev"; \
	fi)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%d %H:%M:%S UTC")
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS := -ldflags "\
	-X 'github.com/zy84338719/fileCodeBox/backend/internal/models/service.Version=$(VERSION)' \
	-X 'github.com/zy84338719/fileCodeBox/backend/internal/models/service.GitCommit=$(COMMIT)' \
	-X 'github.com/zy84338719/fileCodeBox/backend/internal/models/service.BuildTime=$(DATE)' \
	-w -s"

# 默认目标
all: build

# 构建前端
build-frontend:
	@echo "Building frontend..."
	cd frontend && npm run build

# 构建后端
build-backend:
	@echo "Building backend $(VERSION) ($(COMMIT)) at $(DATE)"
	cd backend && go build $(LDFLAGS) -o bin/server ./cmd/server

# 复制前端构建产物到后端 static 目录
copy-frontend:
	@echo "Copying frontend dist to backend/static..."
	@mkdir -p backend/static
	@cp -r frontend/dist/* backend/static/

# 完整构建（前端 + 后端 + 复制）
build: build-frontend build-backend copy-frontend
	@echo "Build complete!"

# 交叉编译（支持环境变量设置平台）
build-cross:
	@echo "Cross-compiling FileCodeBox $(VERSION) for $(GOOS)/$(GOARCH)"
	cd backend && env GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build $(LDFLAGS) -o bin/server ./cmd/server

# 发布构建（优化编译）
release:
	@echo "Building FileCodeBox release $(VERSION) ($(COMMIT)) at $(DATE)"
	cd backend && CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o bin/server ./cmd/server

# 运行项目
run: build
	cd backend && ./bin/server

# 显示版本信息
version:
	@echo "FileCodeBox $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Built: $(DATE)"
	@echo "Go Version: $(GO_VERSION)"

# 运行测试
test:
	cd backend && go test -race -coverprofile=coverage.out ./... -v

# 运行测试（简化版）
test-simple:
	cd backend && go test ./... -v

# 运行开发模式（热重载需要安装air: go install github.com/cosmtrek/air@latest）
dev:
	cd backend && air

# 清理编译文件
clean:
	rm -rf backend/bin
	rm -rf backend/static
	rm -rf frontend/dist
	cd backend && go clean

# 整理依赖
tidy:
	cd backend && go mod tidy

# 检查代码格式
fmt:
	cd backend && go fmt ./...

# 代码检查
vet:
	cd backend && go vet ./...

# 构建Docker镜像
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(DATE) \
		-t filecodebox:$(VERSION) \
		-t filecodebox:latest \
		.

# 运行Docker容器
docker-run: docker-build
	docker-compose up -d

# 停止Docker容器
docker-stop:
	docker-compose down

# 查看Docker日志
docker-logs:
	docker-compose logs -f

# 完整的检查流程
check: fmt vet test

# 安装依赖
deps:
	cd backend && go mod download
	cd frontend && npm install

# 生成API文档
docs:
	@echo "API文档可通过 http://localhost:8888/swagger/index.html 访问"

# 查看帮助
help:
	@echo "可用的make命令："
	@echo "  build          - 完整构建（前端 + 后端 + 复制）"
	@echo "  build-frontend - 仅构建前端"
	@echo "  build-backend  - 仅构建后端"
	@echo "  copy-frontend  - 复制前端产物到后端 static 目录"
	@echo "  release        - 发布构建（优化编译）"
	@echo "  run            - 编译并运行项目"
	@echo "  version        - 显示版本信息"
	@echo "  test           - 运行测试"
	@echo "  dev            - 开发模式（需要安装air）"
	@echo "  clean          - 清理编译文件"
	@echo "  tidy           - 整理依赖"
	@echo "  fmt            - 格式化代码"
	@echo "  vet            - 代码检查"
	@echo "  check          - 完整检查（fmt + vet + test）"
	@echo "  docker-build   - 构建Docker镜像"
	@echo "  docker-run     - 运行Docker容器"
	@echo "  docker-stop    - 停止Docker容器"
	@echo "  docker-logs    - 查看Docker日志"
	@echo "  docs           - 查看API文档说明"
	@echo "  deps           - 安装所有依赖（后端 + 前端）"
	@echo "  help           - 显示此帮助信息"
