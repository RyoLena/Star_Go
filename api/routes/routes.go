// Package routes api/routes/routes.go
package routes

import (
	"star-go/internal/controllers"
	"star-go/pkg/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine) {
	// 添加全局中间件
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.Cors())

	// 添加全局限流中间件 - 每分钟180个请求
	router.Use(middleware.RateLimit(180, time.Minute))

	// API版本分组
	apiGroup := router.Group("/api")
	{
		// 设置认证相关路由
		setupAuthRoutes(apiGroup)

		// 设置用户相关路由
		setupUserRoutes(apiGroup)
	}
}

// 设置认证相关路由
func setupAuthRoutes(apiGroup *gin.RouterGroup) {
	// 创建认证控制器实例
	authController := controllers.NewAuthController()
	// 公开路由组
	publicGroup := apiGroup.Group("/auth")
	{
		// 用户注册
		publicGroup.POST("/register", authController.Register)
		// 用户登录
		publicGroup.POST("/login", authController.Login)
		// 刷新令牌
		publicGroup.POST("/refresh", authController.RefreshToken)
	}

	// 需要认证的路由组
	authGroup := apiGroup.Group("/auth")
	authGroup.Use(middleware.JWTAuth())
	{
		// 获取当前用户信息
		authGroup.GET("/user", authController.GetUserInfo)
		// 修改密码
		authGroup.POST("/change-password", authController.ChangePassword)
	}
}

// 设置用户相关路由
func setupUserRoutes(apiGroup *gin.RouterGroup) {
	// 创建用户控制器实例
	userController := controllers.NewUserController()

	// 基础用户路由组 - 仅需要认证
	baseUserGroup := apiGroup.Group("/users")
	baseUserGroup.Use(middleware.JWTAuth())

	// 使用权限鉴权 - 查看用户列表需要 user:list 权限
	baseUserGroup.GET("", middleware.PermissionAuth("user:list"), userController.GetUsers)

	// 使用权限鉴权 - 查看单个用户需要多个权限中的任意一个
	baseUserGroup.GET("/:id", middleware.AnyPermissionAuth("user:list", "user:read"), userController.GetUser)

	// 使用角色或权限鉴权 - 创建用户需要 admin 角色或 user:create 权限
	baseUserGroup.POST("", middleware.RoleOrPermissionAuth("admin", "user:create"), userController.CreateUser)

	// 使用角色和权限鉴权 - 更新用户需要同时具备 admin 角色和 user:update 权限
	baseUserGroup.PUT("/:id", middleware.RoleAndPermissionAuth("admin", "user:update"), userController.UpdateUser)

	// 使用多权限鉴权 - 删除用户需要同时具备 user:delete 和 user:manage 权限
	baseUserGroup.DELETE("/:id", middleware.AllPermissionsAuth("user:delete", "user:manage"), userController.DeleteUser)

	// 管理员路由组 - 需要角色
	adminGroup := apiGroup.Group("/admin/users")
	adminGroup.Use(middleware.JWTAuth())
	adminGroup.Use(middleware.RoleAuth("admin")) // 使用admin角色
	{
		// 管理员特殊操作 - 这里可以添加其他管理员特有的操作
		// 例如，复用现有的删除用户操作
		adminGroup.DELETE("/:id", userController.DeleteUser)
	}
}

func setupSMSRoutes(apiGroup *gin.RouterGroup) {
	
}
