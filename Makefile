.PHONY: build run test clean docker-build docker-run dev release

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS := -ldflags "\
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE) \
	-w -s"

# 默认目标
all: build

# 编译项目（带版本信息）
build:
	@echo "Building FileCodeBox $(VERSION) ($(COMMIT)) at $(DATE)"
	go build $(LDFLAGS) -o filecodebox .

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
