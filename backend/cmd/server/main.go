package main

import (
    "log"

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
    cfg := config.Load()

    appLogger, err := logger.New(cfg.LogLevel)
    if err != nil {
        log.Fatalf("logger init failed: %v", err)
    }
    defer appLogger.Sync()

    database, err := db.Open(cfg.DBPath)
    if err != nil {
        appLogger.Fatal("db open failed", logger.Error(err))
    }
    defer database.Close()

    if err := db.Migrate(database, cfg); err != nil {
        appLogger.Fatal("db migrate failed", logger.Error(err))
    }

    userRepo := repo.NewUserRepo(database)
    tokenRepo := repo.NewTokenRepo(database)
    authService := service.NewAuthService(userRepo, tokenRepo, cfg)

    authHandler := handler.NewAuthHandler(authService)
    healthHandler := handler.NewHealthHandler()

    r := gin.New()
    r.Use(middleware.RequestID())
    r.Use(middleware.Logger(appLogger))
    r.Use(middleware.CORS(cfg))
    r.Use(gin.Recovery())

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
    }

    appLogger.Info("server starting", logger.String("addr", cfg.ServerAddr))
    if err := r.Run(cfg.ServerAddr); err != nil {
        appLogger.Fatal("server stopped", logger.Error(err))
    }
}
