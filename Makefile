.PHONY: build run test clean docker-build docker-run dev

# 默认目标
all: build

# 编译项目
build:
	go build -o filecodebox .

# 运行项目
run: build
	./filecodebox

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
	docker build -t filecodebox-go .

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
	@echo "API文档可通过 http://localhost:12345/api/doc 访问"

# 查看帮助
help:
	@echo "可用的make命令："
	@echo "  build       - 编译项目"
	@echo "  run         - 编译并运行项目"
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
