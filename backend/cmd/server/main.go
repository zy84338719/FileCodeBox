package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/zy84338719/fileCodeBox/backend/cmd/server/bootstrap"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数。
	// --config 指定配置文件路径；亦可通过 CONFIG_PATH 环境变量设置（flag 优先）。
	configPath := flag.String("config", "", "配置文件路径（默认 configs/config.yaml，可用 CONFIG_PATH 环境变量覆盖）")
	flag.Parse()

	h, err := bootstrap.Bootstrap(*configPath)
	if err != nil {
		logger.Fatal("bootstrap failed", zap.Error(err))
	}
	defer bootstrap.Cleanup()

	go func() {
		logger.Info("Server starting...")
		h.Spin()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	h.Shutdown(context.Background())
	logger.Info("Server stopped")
}
