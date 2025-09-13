package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/database"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "fix-stats" {
		fmt.Println("Usage: go run fix_stats.go fix-stats")
		os.Exit(1)
	}

	// 初始化配置
	manager := config.InitManager()

	// 初始化数据库
	db, err := database.InitWithManager(manager)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化Repository管理器
	repositoryManager := repository.NewRepositoryManager(db)

	// 初始化用户服务
	userService := services.NewUserService(repositoryManager, manager)

	// 修复所有用户的统计数据
	fmt.Println("开始修复用户统计数据...")
	err = userService.RecalculateAllUsersStats()
	if err != nil {
		log.Fatalf("Failed to recalculate user stats: %v", err)
	}

	fmt.Println("用户统计数据修复完成！")
}
