package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-devops/internal/api"
	"go-devops/internal/config"
	"go-devops/internal/database"
	"go-devops/internal/logger"
	"go-devops/internal/middleware"
	"go-devops/internal/scheduler"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	cfg := config.Load()

	// 初始化日志系统
	logger.Init()
	logger.Info("应用程序启动")

	// 初始化数据库
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("数据库初始化失败:", err)
	}
	logger.Info("数据库初始化成功")

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建路由 (使用New()避免重复中间件)
	r := gin.New()

	// 中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// 静态文件服务
	r.Static("/static", "./web/dist")
	r.StaticFile("/", "./web/dist/index.html")

	// API路由
	apiRouter := r.Group("/api/v1")
	api.SetupRoutes(apiRouter, db)

	// 启动定时任务调度器
	taskScheduler := scheduler.NewScheduler(db)
	taskScheduler.Start()

	// 优雅关闭处理
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		logger.Info("接收到关闭信号，正在关闭服务...")
		taskScheduler.Stop()
		os.Exit(0)
	}()

	// 启动服务器
	logger.Infof("服务器启动在端口 %s", cfg.Port)
	logger.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
