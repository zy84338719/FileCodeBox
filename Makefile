.PHONY: build run test clean docker-build docker-run dev release version deps tidy fmt vet check help

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
	-X 'github.com/zy84338719/filecodebox/internal/models/service.Version=$(VERSION)' \
	-X 'github.com/zy84338719/filecodebox/internal/models/service.GitCommit=$(COMMIT)' \
	-X 'github.com/zy84338719/filecodebox/internal/models/service.BuildTime=$(DATE)' \
	-w -s"

# 默认目标
all: build

# 编译项目（带版本信息）
build:
	@echo "Building FileCodeBox $(VERSION) ($(COMMIT)) at $(DATE)"
	go build $(LDFLAGS) -o filecodebox .

# 交叉编译（支持环境变量设置平台）
build-cross:
	@echo "Cross-compiling FileCodeBox $(VERSION) for $(GOOS)/$(GOARCH)"
	env GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build $(LDFLAGS) -o filecodebox .

# 发布构建（优化编译）
release:
	@echo "Building FileCodeBox release $(VERSION) ($(COMMIT)) at $(DATE)"
	CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o filecodebox .

# 运行项目
run: build
	./filecodebox

# 显示版本信息
version:
	@echo "FileCodeBox $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Built: $(DATE)"
	@echo "Go Version: $(GO_VERSION)"

# 运行测试
test:
	go test -race -coverprofile=coverage.out ./... -v

# 运行测试（简化版）
test-simple:
	go test ./... -v

# 运行开发模式（热重载需要安装air: go install github.com/cosmtrek/air@latest）
dev:
	air

# 清理编译文件
clean:
	rm -f filecodebox
	go clean

# 整理依赖
tidy:
	go mod tidy

# 检查代码格式
fmt:
	go fmt ./...

# 代码检查
vet:
	go vet ./...

# 构建Docker镜像
docker-build:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(DATE) -t filecodebox-go .

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
	go mod download

# 生成API文档
docs:
	@echo "API文档可通过 http://localhost:12345/swagger/index.html 访问"

# 查看帮助
help:
	@echo "可用的make命令："
	@echo "  build       - 编译项目（带版本信息）"
	@echo "  release     - 发布构建（优化编译）"
	@echo "  run         - 编译并运行项目"
	@echo "  version     - 显示版本信息"
	@echo "  test        - 运行测试"
	@echo "  dev         - 开发模式（需要安装air）"
	@echo "  clean       - 清理编译文件"
	@echo "  tidy        - 整理依赖"
	@echo "  fmt         - 格式化代码"
	@echo "  vet         - 代码检查"
	@echo "  check       - 完整检查（fmt + vet + test）"
	@echo "  docker-build - 构建Docker镜像"
	@echo "  docker-run  - 运行Docker容器"
	@echo "  docker-stop - 停止Docker容器"
	@echo "  docker-logs - 查看Docker日志"
	@echo "  docs        - 查看API文档说明"
	@echo "  help        - 显示此帮助信息"
