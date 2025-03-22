// Package middleware pkg/middleware/auth.go
package middleware

import (
	"fmt"
	"net/http"
	"star-go/internal/services"
	"star-go/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证格式错误，应为 'Bearer {token}'",
			})
			c.Abort()
			return
		}

		// 解析令牌
		token := parts[1]
		claims, err := utils.ParseAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证令牌: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// RoleAuth 角色授权中间件
func RoleAuth(roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证用户",
			})
			c.Abort()
			return
		}

		// 创建用户服务
		userService := services.NewUserService()

		// 获取用户信息
		user, err := userService.GetUserByID(userID.(uint64))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取用户信息失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户角色
		fmt.Println("User Role:", user.Role.Code, "Required Role:", roleCode)
		if user.Role == nil || user.Role.Code != roleCode {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足，需要 " + roleCode + " 角色",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionAuth 权限授权中间件
func PermissionAuth(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证用户",
			})
			c.Abort()
			return
		}

		// 创建用户服务
		userService := services.NewUserService()

		// 检查用户是否有指定权限
		hasPermission, err := userService.HasPermission(userID.(uint64), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "检查权限失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足，需要 " + permission + " 权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleAndPermissionAuth 同时需要角色和权限的中间件
func RoleAndPermissionAuth(roleCode string, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证用户",
			})
			c.Abort()
			return
		}

		// 创建用户服务
		userService := services.NewUserService()

		// 获取用户信息
		user, err := userService.GetUserByID(userID.(uint64))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取用户信息失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户角色
		if user.Role == nil || user.Role.Code != roleCode {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足，需要 " + roleCode + " 角色",
			})
			c.Abort()
			return
		}

		// 检查用户是否有指定权限
		hasPermission, err := userService.HasPermission(userID.(uint64), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "检查权限失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足，需要 " + permission + " 权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleOrPermissionAuth 需要角色或权限的中间件（满足其一即可）
func RoleOrPermissionAuth(roleCode string, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证用户",
			})
			c.Abort()
			return
		}

		// 创建用户服务
		userService := services.NewUserService()

		// 获取用户信息
		user, err := userService.GetUserByID(userID.(uint64))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取用户信息失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户角色
		hasRole := user.Role != nil && user.Role.Code == roleCode

		// 检查用户是否有指定权限
		hasPermission, err := userService.HasPermission(userID.(uint64), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "检查权限失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 如果既没有所需角色也没有所需权限，则拒绝访问
		if !hasRole && !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足，需要 " + roleCode + " 角色或 " + permission + " 权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AnyPermissionAuth 需要多个权限中的任意一个的中间件
func AnyPermissionAuth(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证用户",
			})
			c.Abort()
			return
		}

		// 创建用户服务
		userService := services.NewUserService()

		// 检查用户是否有指定权限中的任意一个
		for _, permission := range permissions {
			hasPermission, err := userService.HasPermission(userID.(uint64), permission)
			if err != nil {
				continue // 忽略错误，继续检查其他权限
			}
			if hasPermission {
				// 有任意一个权限即可通过
				c.Next()
				return
			}
		}

		// 构建权限列表字符串
		permissionList := strings.Join(permissions, ", ")

		// 所有权限都没有，拒绝访问
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，需要以下权限之一: " + permissionList,
		})
		c.Abort()
	}
}

// AllPermissionsAuth 需要所有指定权限的中间件
func AllPermissionsAuth(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证用户",
			})
			c.Abort()
			return
		}

		// 创建用户服务
		userService := services.NewUserService()

		// 检查用户是否拥有所有指定权限
		for _, permission := range permissions {
			hasPermission, err := userService.HasPermission(userID.(uint64), permission)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "检查权限失败: " + err.Error(),
				})
				c.Abort()
				return
			}
			if !hasPermission {
				// 缺少任意一个权限都拒绝访问
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "权限不足，需要权限: " + permission,
				})
				c.Abort()
				return
			}
		}

		// 拥有所有权限，允许访问
		c.Next()
	}
}
