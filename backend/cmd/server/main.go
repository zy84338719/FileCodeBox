package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zy84338719/fileCodeBox/cmd/server/bootstrap"
	"github.com/zy84338719/fileCodeBox/internal/pkg/logger"
)

func main() {
	h, err := bootstrap.Bootstrap()
	if err != nil {
		fmt.Printf("Bootstrap failed: %v\n", err)
		os.Exit(1)
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
