.PHONY: build run test clean docker-build docker-run lint gen-http-new gen-http-update gen-http-update-all gen-http-init gen-rpc gen-rpc-all gen-thrift-new gen-thrift-update gen-thrift-update-all kitex-install hz-install thriftgo-install tools-install

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

# ===== Thrift HTTP 代码生成（生成到 gen/http/，与 proto 共享同一目录）=====
# 用法同 gen-http-*,只是 IDL 改成 *.thrift

# 初始化新项目（首次使用，thrift）
# 用法: make gen-thrift-new IDL=common.thrift
gen-thrift-new:
	@if [ -z "$(IDL)" ]; then \
		echo "用法: make gen-thrift-new IDL=<thrift文件名>"; \
		echo "示例: make gen-thrift-new IDL=common.thrift"; \
		exit 1; \
	fi
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) thrift-new $(IDL_PATH)/$(IDL)

# 更新 HTTP 代码（指定单个 thrift IDL）
# 用法: make gen-thrift-update IDL=common.thrift
gen-thrift-update:
	@if [ -z "$(IDL)" ]; then \
		echo "用法: make gen-thrift-update IDL=<thrift文件名>"; \
		echo "示例: make gen-thrift-update IDL=common.thrift"; \
		exit 1; \
	fi
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) thrift-update $(IDL_PATH)/$(IDL)

# 批量更新所有 thrift IDL
gen-thrift-update-all:
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) thrift-update-all

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

thriftgo-install:
	go install github.com/cloudwego/thriftgo/cmd/thriftgo@latest

kitex-install:
	go install github.com/cloudwego/kitex/tool/cmd/kitex@latest

tools-install: hz-install thriftgo-install kitex-install
	@echo "工具安装完成"

# ===== 兼容旧命令 =====

gen-new: gen-http-new
gen-update: gen-http-update
gen-update-all: gen-http-update-all

.PHONY: all
all: tidy build
