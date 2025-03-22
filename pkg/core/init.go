// Package core pkg/core/init.go
package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"star-go/pkg/cache"
	"star-go/pkg/config"
	"star-go/pkg/database"
	"star-go/pkg/logger"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Application 应用程序结构
type Application struct {
	server *http.Server
	engine *gin.Engine
}

// InitConfig 初始化配置
func InitConfig(configPath string) error {
	return config.InitConfig(configPath)
}

// InitLogger 初始化日志系统
func InitLogger() error {
	return logger.InitLogger()
}

// InitDatabase 初始化数据库
func InitDatabase() error {
	return database.InitDatabase()
}

// InitCache 初始化缓存
func InitCache() error {
	return cache.InitCache()
}

// InitGin 初始化Gin引擎
func InitGin() *gin.Engine {
	// 设置Gin模式
	gin.SetMode(config.GetConfig().Server.Mode)

	// 如果配置了禁用调试输出，则禁用Gin的调试输出
	if config.GetConfig().Server.DisableDebug {
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = io.Discard
	}

	// 创建Gin引擎
	router := gin.New()

	// 使用自定义日志和恢复中间件
	router.Use(logger.GinLogger())
	router.Use(logger.GinRecovery(true))

	return router
}

// NewApplication 创建应用程序
func NewApplication(router *gin.Engine) *Application {
	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.GetConfig().Server.Host, config.GetConfig().Server.Port),
		Handler:      router,
		ReadTimeout:  config.GetConfig().Server.ReadTimeout * time.Second,
		WriteTimeout: config.GetConfig().Server.WriteTimeout * time.Second,
	}

	return &Application{
		server: srv,
		engine: router,
	}
}

// Engine 获取Gin引擎
func (app *Application) Engine() *gin.Engine {
	return app.engine
}

// Run 启动应用程序
func (app *Application) Run() {
	// 启动HTTP服务器
	go func() {
		host := config.GetConfig().Server.Host
		port := config.GetConfig().Server.Port

		// 根据监听地址确定显示的URL
		var serverURL string
		if host == "0.0.0.0" {
			serverURL = fmt.Sprintf("http://localhost:%d", port)
		} else {
			serverURL = fmt.Sprintf("http://%s:%d", host, port)
		}

		logger.GetLogger().Info(fmt.Sprintf("服务器启动在 %s", serverURL))
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.GetLogger().Fatal("启动服务器失败",
				zap.Error(err))
		}
	}()

	// 等待中断信号优雅关闭服务器
	app.gracefulShutdown()
}

// 优雅关闭
func (app *Application) gracefulShutdown() {
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.GetLogger().Info("正在关闭服务器...")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := app.server.Shutdown(ctx); err != nil {
		logger.GetLogger().Fatal("服务器关闭失败",
			zap.Error(err))
	}

	// 关闭数据库连接
	database.CloseDatabase()

	// 关闭日志系统
	logger.CloseLogger()

	logger.GetLogger().Info("服务器已关闭")
}
