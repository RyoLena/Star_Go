// Package logger pkg/logger/logger.go
package logger

import (
	"errors"
	"os"
	"path/filepath"
	"star-go/pkg/config"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log 全局日志对象
var Log *zap.Logger

// InitLogger 初始化日志系统
func InitLogger() error {
	cfg := config.GetConfig().Log

	// 设置日志级别
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 设置控制台日志编码器
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}

	// 根据配置设置彩色输出
	if cfg.ColorOutput {
		consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		consoleEncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + t.Format("2006-01-02 15:04:05") + "]")
		}
	} else {
		consoleEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		consoleEncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + t.Format("2006-01-02 15:04:05") + "]")
		}
	}

	// 设置文件日志编码器
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心列表
	cores := []zapcore.Core{}

	// 添加控制台输出
	var consoleEncoder zapcore.Encoder
	if cfg.Format == "json" {
		consoleEncoder = zapcore.NewJSONEncoder(consoleEncoderConfig)
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
	}
	cores = append(cores, zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		level,
	))

	// 如果启用文件日志，添加文件输出
	if cfg.EnableFile {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// 配置日志输出
		hook := &lumberjack.Logger{
			Filename:   cfg.Filename,   // 日志文件路径
			MaxSize:    cfg.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
			MaxBackups: cfg.MaxBackups, // 日志文件最多保存多少个备份
			MaxAge:     cfg.MaxAge,     // 文件最多保存多少天
			Compress:   cfg.Compress,   // 是否压缩
		}

		// 创建文件日志编码器
		var fileEncoder zapcore.Encoder
		fileEncoder = zapcore.NewJSONEncoder(fileEncoderConfig)

		cores = append(cores, zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(hook),
			level,
		))
	}

	// 创建日志核心
	core := zapcore.NewTee(cores...)

	// 创建日志对象
	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// 替换全局logger
	zap.ReplaceGlobals(Log)

	return nil
}

// GetLogger 获取日志对象
func GetLogger() *zap.Logger {
	return Log
}

// CloseLogger 关闭日志
func CloseLogger() {
	if Log != nil {
		Log.Sync()
	}
}

// GinLogger 创建Gin日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取配置，检查是否启用请求日志
		cfg := config.GetConfig()

		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		// 只有在启用请求日志时才记录
		if cfg.Log.EnableConsole {
			cost := time.Since(start)
			// 转换为毫秒
			cost = cost / time.Millisecond

			// 简化日志输出，只记录关键信息
			Log.Info("请求",
				zap.Int("状态", c.Writer.Status()),
				zap.String("方法", c.Request.Method),
				zap.String("路径", path),
				zap.Duration("耗时", cost),
			)
		}
	}
}

// GinRecovery 创建Gin恢复中间件
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查是否已关闭连接
				var brokenPipe bool
				if ne, ok := err.(*os.PathError); ok {
					var se *os.SyscallError
					if errors.As(ne.Err, &se) {
						if se.Err != nil && se.Err.Error() == "broken pipe" {
							brokenPipe = true
						}
					}
				}

				// 记录堆栈信息
				httpRequest := c.Request.Method + " " + c.Request.URL.String()
				if brokenPipe {
					Log.Error("连接断开",
						zap.Any("错误", err),
						zap.String("请求", httpRequest),
					)
					// 如果连接已断开，直接返回
					c.Error(err.(error))
					c.Abort()
					return
				}

				if stack {
					Log.Error("恢复异常",
						zap.Any("错误", err),
						zap.String("请求", httpRequest),
						zap.Stack("堆栈"),
					)
				} else {
					Log.Error("恢复异常",
						zap.Any("错误", err),
						zap.String("请求", httpRequest),
					)
				}

				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
