// 服务入口与路由注册
package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"dblive/internal/config"
	"dblive/internal/db"
	"dblive/internal/handler"
	"dblive/internal/logger"
	"dblive/internal/middleware"
	"dblive/internal/repo"
	"dblive/internal/service"
)

func main() {
	// 读取运行配置
	cfg := config.Load()

	appLogger, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("logger init failed: %v", err)
	}
	defer appLogger.Sync()

	// 打开数据库连接
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		appLogger.Fatal("db open failed", logger.Error(err))
	}
	defer database.Close()

	// 运行数据库迁移与初始化
	if err := db.Migrate(database, cfg); err != nil {
		appLogger.Fatal("db migrate failed", logger.Error(err))
	}

	userRepo := repo.NewUserRepo(database)
	tokenRepo := repo.NewTokenRepo(database)
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)
	biliCookieRepo := repo.NewBiliCookieRepo(database)
	settingsRepo := repo.NewSettingsRepo(database)
	recordingRepo := repo.NewRecordingRepo(database)
	biliService := service.NewBilibiliService(biliCookieRepo)
	settingsService := service.NewSettingsService(settingsRepo)
	recordingService := service.NewRecordingService(recordingRepo, biliService, settingsService)

	authHandler := handler.NewAuthHandler(authService)
	healthHandler := handler.NewHealthHandler()
	biliCookieHandler := handler.NewBilibiliCookieHandler(biliService)
	recordingHandler := handler.NewRecordingHandler(recordingService)
	settingsHandler := handler.NewSettingsHandler(settingsService)

	// 构建路由与中间件
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(appLogger))
	r.Use(middleware.CORS(cfg))
	r.Use(gin.Recovery())

	// API 版本路由分组
	api := r.Group("/api/v1")
	{
		api.GET("/health", healthHandler.Health)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}
		api.GET("/me", middleware.AuthRequired(cfg), authHandler.Me)

		// 登录后可访问的接口
		secured := api.Group("", middleware.AuthRequired(cfg))
		{
			secured.POST("/bilibili/cookie/qrcode", biliCookieHandler.GenerateQRCode)
			secured.POST("/bilibili/cookie/poll", biliCookieHandler.PollQRCode)
			secured.GET("/bilibili/cookie/status", biliCookieHandler.Status)

			secured.GET("/recordings", recordingHandler.List)
			secured.POST("/recordings", recordingHandler.Create)
			secured.PUT("/recordings/:id", recordingHandler.Update)
			secured.POST("/recordings/:id/start", recordingHandler.Start)
			secured.POST("/recordings/:id/stop", recordingHandler.Stop)
			secured.DELETE("/recordings/:id", recordingHandler.Delete)
			secured.GET("/settings", settingsHandler.Get)
			secured.PUT("/settings", settingsHandler.Update)
		}
	}

	// 启动后台巡检任务
	go recordingService.StartScheduler(context.Background())

	// 前端静态资源与 SPA 回退
	r.Static("/assets", "./public/assets")
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File("./public/index.html")
	})

	appLogger.Info("server starting", logger.String("addr", cfg.ServerAddr))
	if err := r.Run(cfg.ServerAddr); err != nil {
		appLogger.Fatal("server stopped", logger.Error(err))
	}
}
