.PHONY: build run test clean docker-build docker-run lint gen-http-new gen-http-update gen-http-update-all gen-http-init gen-rpc gen-rpc-all kitex-install hz-install

APP_NAME=FileCodeBox
MODULE_NAME=github.com/zy84338719/fileCodeBox/backend
MAIN_PATH=./cmd/server
IDL_PATH=idl
GEN_SCRIPT=./scripts/gen.sh

# ===== 构建与运行 =====

build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)/main.go

dev:
	go run $(MAIN_PATH)/main.go

test:
	go test ./... -v

clean:
	rm -rf bin/
	rm -rf logs/

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

# ===== Docker =====

docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker run -p 8888:8888 $(APP_NAME):latest

# ===== Hz HTTP 代码生成 (生成到 gen/http/) =====

# 初始化新项目（首次使用）
# 用法: make gen-http-new IDL=common.proto
gen-http-new:
	@if [ -z "$(IDL)" ]; then \
		echo "用法: make gen-http-new IDL=<proto文件名>"; \
		echo "示例: make gen-http-new IDL=common.proto"; \
		exit 1; \
	fi
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) hz-new $(IDL_PATH)/$(IDL)

# 更新 HTTP 代码（指定单个 IDL）
# 用法: make gen-http-update IDL=common.proto
gen-http-update:
	@if [ -z "$(IDL)" ]; then \
		echo "用法: make gen-http-update IDL=<proto文件名>"; \
		echo "示例: make gen-http-update IDL=common.proto"; \
		echo "      make gen-http-update IDL=http/health.proto"; \
		exit 1; \
	fi
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) hz-update $(IDL_PATH)/$(IDL)

# 批量更新所有 HTTP IDL（扫描 idl/*.proto 和 idl/http/*.proto）
gen-http-update-all:
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) hz-update-all

# 强制重新初始化 .hz 配置（备份 handler 后执行 hz new --force）
# 用法: make gen-http-init IDL=common.proto
gen-http-init:
	@if [ -z "$(IDL)" ]; then \
		echo "用法: make gen-http-init IDL=<proto文件名>"; \
		echo "示例: make gen-http-init IDL=common.proto"; \
		exit 1; \
	fi
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) hz-init $(IDL_PATH)/$(IDL)

# ===== Kitex RPC 代码生成 (生成到 gen/rpc/) =====

# 生成 RPC 代码
# 用法: make gen-rpc IDL=rpc/user.proto
gen-rpc:
	@if [ -z "$(IDL)" ]; then \
		echo "用法: make gen-rpc IDL=<proto文件路径>"; \
		echo "示例: make gen-rpc IDL=rpc/user.proto"; \
		exit 1; \
	fi
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) kitex $(IDL_PATH)/$(IDL)

# 生成所有 RPC 代码
gen-rpc-all:
	@chmod +x $(GEN_SCRIPT)
	@for f in $$(find $(IDL_PATH)/rpc -name "*.proto" ! -name "base.proto" 2>/dev/null); do \
		if [ -f "$$f" ]; then \
			echo "生成 RPC: $$f ..."; \
			$(GEN_SCRIPT) kitex "$$f"; \
		fi \
	done

# ===== 工具安装 =====

hz-install:
	go install github.com/cloudwego/hertz/cmd/hz@latest

kitex-install:
	go install github.com/cloudwego/kitex/tool/cmd/kitex@latest

tools-install: hz-install kitex-install
	@echo "工具安装完成"

# ===== 兼容旧命令 =====

gen-new: gen-http-new
gen-update: gen-http-update
gen-update-all: gen-http-update-all

.PHONY: all
all: tidy build
