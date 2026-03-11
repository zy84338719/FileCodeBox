package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/zy84338719/fileCodeBox/cmd/server/bootstrap"
	conf "github.com/zy84338719/fileCodeBox/internal/conf"
	"github.com/zy84338719/fileCodeBox/internal/repo/db"
)

// CORS 跨域中间件
func CORS() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		if origin == "" {
			origin = "*"
		}

		// 设置 CORS 头
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24小时

		// 处理预检请求
		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next(ctx)
	}
}

func main() {
	// 1. 初始化配置
	log.Println("Loading configuration...")
	cfg, err := bootstrap.InitConfig("configs/config.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load config: %v, using defaults", err)
		cfg = &bootstrap.Config{
			Server: conf.ServerConfig{
				Host: "0.0.0.0",
				Port: 12345,
				Mode: "debug",
			},
			Database: conf.DatabaseConfig{
				Driver: "sqlite",
				DBName: "./data/filecodebox.db",
			},
		}
	}

	// 2. 初始化数据库
	log.Println("Initializing database...")
	database, err := bootstrap.InitDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置全局数据库实例
	db.SetDatabaseInstance(database)

	// 3. 创建默认管理员
	if err := bootstrap.CreateDefaultAdmin(database); err != nil {
		log.Printf("Warning: Failed to create admin user: %v", err)
	}

	// 4. 启动 HTTP 服务器
	log.Printf("Starting HTTP server on %s:%d", cfg.Server.Host, cfg.Server.Port)

	opts := []config.Option{
		server.WithHostPorts(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
		server.WithDisablePrintRoute(cfg.Server.Mode != "debug"),
	}
	h := server.Default(opts...)

	// 添加 CORS 中间件
	h.Use(CORS())

	// 注册路由
	register(h)

	log.Println("Server started successfully")
	log.Println("API Endpoints:")
	log.Println("  - POST http://localhost:12345/user/register")
	log.Println("  - POST http://localhost:12345/user/login")
	log.Println("  - POST http://localhost:12345/share/text/")
	log.Println("  - POST http://localhost:12345/share/file/")
	log.Println("  - GET  http://localhost:12345/share/select/")

	h.Spin()
}
