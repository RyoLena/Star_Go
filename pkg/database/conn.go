// pkg/database/conn.go
package database

import (
	"fmt"
	"io"
	"log"
	"os"
	"star-go/pkg/config"
	"star-go/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB 全局数据库连接对象
var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	cfg := config.GetConfig().Database

	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// 配置GORM日志
	var logWriter io.Writer
	if cfg.DisableSqlLog {
		// 如果禁用SQL日志，将输出重定向到空设备
		logWriter = io.Discard
	} else {
		// 否则使用标准输出
		logWriter = os.Stdout
	}

	// 设置日志级别
	var logLevel gormlogger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	case "info":
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Info
	}

	// 创建GORM日志记录器
	gormLogger := gormlogger.New(
		log.New(logWriter, "\r\n", log.LstdFlags), // io writer
		gormlogger.Config{
			SlowThreshold:             cfg.SlowThreshold * time.Millisecond, // 慢查询阈值
			LogLevel:                  logLevel,                             // 日志级别
			IgnoreRecordNotFoundError: true,                                 // 忽略记录未找到错误
			Colorful:                  true,                                 // 彩色打印
		},
	)

	// 打开数据库连接
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层SQL连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接池失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Second)

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("测试数据库连接失败: %w", err)
	}

	logger.GetLogger().Info("数据库连接成功")

	// 根据配置决定是否自动迁移数据库
	if cfg.AutoMigrate {
		if err := RunMigrations(); err != nil {
			logger.GetLogger().Error("数据库迁移失败", zap.Error(err))
			return err
		}
	}

	// 根据配置决定是否初始化管理员账户
	if cfg.InitAdmin {
		if err := InitAdminUser(); err != nil {
			logger.GetLogger().Error("初始化管理员账户失败", zap.Error(err))
			return err
		}
	}
	return nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			logger.GetLogger().Error("获取数据库连接失败", zap.Error(err))
			return
		}
		if err := sqlDB.Close(); err != nil {
			logger.GetLogger().Error("关闭数据库连接失败", zap.Error(err))
			return
		}
		logger.GetLogger().Info("数据库连接已关闭")
	}
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}
