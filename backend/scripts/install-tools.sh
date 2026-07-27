#!/usr/bin/env bash
#
# install-tools.sh — 一键安装 hz + thriftgo 工具
#
# 用法:
#   ./scripts/install-tools.sh          # 安装全部
#   ./scripts/install-tools.sh hz       # 只装 hz
#   ./scripts/install-tools.sh thriftgo # 只装 thriftgo
#
# 需要 GOPATH/bin 在 PATH 里（hz/thriftgo 默认装到 $GOPATH/bin）

set -euo pipefail

# 颜色
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC}  $*"; }

install_hz() {
  log "安装 hz (CloudWeGo Hertz 代码生成器)..."
  go install github.com/cloudwego/hertz/cmd/hz@latest
  log "hz 安装完成: $(which hz 2>/dev/null || echo "$GOPATH/bin/hz")"
}

install_thriftgo() {
  log "安装 thriftgo (thrift IDL 解析器)..."
  go install github.com/cloudwego/thriftgo/cmd/thriftgo@latest
  log "thriftgo 安装完成: $(which thriftgo 2>/dev/null || echo "$GOPATH/bin/thriftgo")"
}

case "${1:-all}" in
  hz)
    install_hz
    ;;
  thriftgo)
    install_thriftgo
    ;;
  all|"")
    install_hz
    install_thriftgo
    log "全部工具安装完成 ✅"
    ;;
  *)
    echo "未知命令: $1"
    echo "用法: $0 [hz|thriftgo|all]"
    exit 1
    ;;
esac

warn "如果 which hz 找不到，请把 \$GOPATH/bin 加到 PATH:"
warn "  export PATH=\$PATH:\$(go env GOPATH)/bin"
