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

# tools-install 安装全部工具（hz + thriftgo + kitex）
tools-install: hz-install thriftgo-install kitex-install
	@echo "工具安装完成"

# install-tools 是 tools-install 的友好别名（CI 使用）
install-tools: tools-install
	@echo "alias: install-tools -> tools-install"

# ===== 顶层 IDL 生成（CI 一键入口） =====

# 新版 thrift IDL（无 legacy proto-gen 产物）— 安全生成
THRIFT_NEW_IDLS := idl/notify.thrift idl/presign.thrift idl/ratelimit.thrift idl/share_anonymous.thrift

# thrift-gen 批量生成全部 thrift IDL（推荐 CI 使用）
# 流程：
#   1. 仅生成"新" thrift IDL（无 legacy .pb.go 残留的服务）
#   2. legacy IDL（admin/user/share/...）保留 proto-gen 代码直到 task4 完成
#   3. 跳过 go mod tidy / go fmt（保持幂等，避免无关格式变更）
#
# 如需强制全量重新生成（含 legacy 服务），用 thrift-gen-all（注意：会破坏 build 直到 task4 完成）
thrift-gen:
	@echo "==> 批量生成 thrift IDL（新服务）..."
	@chmod +x $(GEN_SCRIPT)
	@for f in $(THRIFT_NEW_IDLS); do \
		echo "-------------------------------------------"; \
		echo "处理 thrift IDL: $$f"; \
		SKIP_TIDY=1 SKIP_FMT=1 $(GEN_SCRIPT) thrift-update $$f || { echo "❌ 生成失败: $$f"; exit 1; }; \
	done
	@echo "✅ thrift-gen 完成（仅新服务）"

# thrift-gen-all 强制全量重新生成（task4 完成后用）
thrift-gen-all:
	@echo "==> 清理 proto 残留 (.pb.go)..."
	@find $(MODEL_DIR) -name "*.pb.go" -type f -delete 2>/dev/null || true
	@echo "==> 批量生成全部 thrift IDL..."
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) thrift-update-all

# hz-gen 批量生成全部 proto IDL（保留兼容）
hz-gen:
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) hz-update-all

# rpc-gen 批量生成全部 kitex RPC IDL
rpc-gen:
	@chmod +x $(GEN_SCRIPT)
	$(GEN_SCRIPT) kitex $$(find $(IDL_PATH)/rpc -name "*.proto" 2>/dev/null | head -1)

# generate 顶层入口：thrift 优先
generate: thrift-gen
	@echo "✅ IDL 代码生成完成"

# ===== 兼容旧命令 =====

gen-new: gen-http-new
gen-update: gen-http-update
gen-update-all: gen-http-update-all

.PHONY: all build run test clean docker-build docker-run lint gen-http-new gen-http-update gen-http-update-all gen-http-init gen-rpc gen-rpc-all gen-thrift-new gen-thrift-update gen-thrift-update-all kitex-install hz-install thriftgo-install tools-install install-tools thrift-gen thrift-gen-all hz-gen rpc-gen generate
all: tidy build
