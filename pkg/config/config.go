// Package config pkg/config/config.go
package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

var globalConfig AppConfig

// AppConfig 应用配置
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	Cache    CacheConfig    `mapstructure:"cache"` // 缓存配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host             string        `mapstructure:"host"` // 服务器监听地址
	Port             int           `mapstructure:"port"`
	Mode             string        `mapstructure:"mode"`
	ReadTimeout      time.Duration `mapstructure:"readTimeout"`
	WriteTimeout     time.Duration `mapstructure:"writeTimeout"`
	DisableDebug     bool          `mapstructure:"disableDebug"`     // 是否禁用Gin调试输出
	EnableRequestLog bool          `mapstructure:"enableRequestLog"` // 是否启用请求日志
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            string        `mapstructure:"type"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	LogLevel        string        `mapstructure:"logLevel"`
	SlowThreshold   time.Duration `mapstructure:"slowThreshold"`
	DisableSqlLog   bool          `mapstructure:"disableSqlLog"` // 是否禁用SQL日志
	AutoMigrate     bool          `mapstructure:"autoMigrate"`   // 是否自动迁移数据库
	InitAdmin       bool          `mapstructure:"initAdmin"`     // 是否初始化管理员账户
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret           string        `mapstructure:"secret"`
	AccessTokenExp   time.Duration `mapstructure:"accessTokenExp"`
	RefreshTokenExp  time.Duration `mapstructure:"refreshTokenExp"`
	TokenIssuer      string        `mapstructure:"tokenIssuer"`
	RefreshTokenSize int           `mapstructure:"refreshTokenSize"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Type         string        `mapstructure:"type"`         // 缓存类型 (redis, memory)
	Host         string        `mapstructure:"host"`         // Redis主机
	Port         int           `mapstructure:"port"`         // Redis端口
	Password     string        `mapstructure:"password"`     // Redis密码
	DB           int           `mapstructure:"db"`           // Redis数据库索引
	PoolSize     int           `mapstructure:"poolSize"`     // 连接池大小
	MinIdleConns int           `mapstructure:"minIdleConns"` // 最小空闲连接数
	MaxRetries   int           `mapstructure:"maxRetries"`   // 最大重试次数
	DialTimeout  time.Duration `mapstructure:"dialTimeout"`  // 连接超时时间（秒）
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`  // 读取超时时间（秒）
	WriteTimeout time.Duration `mapstructure:"writeTimeout"` // 写入超时时间（秒）
	DefaultTTL   time.Duration `mapstructure:"defaultTTL"`   // 默认过期时间（秒）
	Prefix       string        `mapstructure:"prefix"`       // 键前缀
	EnableLog    bool          `mapstructure:"enableLog"`    // 是否启用日志
}

// LogConfig 日志配置
type LogConfig struct {
	Level         string `mapstructure:"level"`
	Filename      string `mapstructure:"filename"`
	MaxSize       int    `mapstructure:"maxSize"`
	MaxBackups    int    `mapstructure:"maxBackups"`
	MaxAge        int    `mapstructure:"maxAge"`
	Compress      bool   `mapstructure:"compress"`
	EnableFile    bool   `mapstructure:"enableFile"`    // 是否启用文件日志
	Format        string `mapstructure:"format"`        // 日志格式 json/console
	ColorOutput   bool   `mapstructure:"colorOutput"`   // 控制台日志是否彩色输出
	EnableConsole bool   `mapstructure:"enableConsole"` // 是否启用控制台日志
}

// 全局配置变量

// InitConfig 初始化配置
func InitConfig(configPath string) error {
	// 设置配置文件路径
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(&globalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	log.Println("配置文件加载成功")
	return nil
}

// GetConfig 获取配置
func GetConfig() *AppConfig {
	return &globalConfig
}
