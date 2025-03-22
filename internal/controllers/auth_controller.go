// Package controllers internal/controllers/auth_controller.go
package controllers

import (
	"star-go/internal/services"
	"star-go/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct {
	authService services.IAuthService
}

// NewAuthController 创建认证控制器实例
func NewAuthController() *AuthController {
	return &AuthController{
		authService: services.NewAuthService(),
	}
}

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"required,min=2,max=50"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest 刷新令牌请求参数
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest 修改密码请求参数
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20"`
}

// Register 用户注册
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, err.Error(), nil)
		return
	}

	user, err := c.authService.Register(req.Username, req.Password, req.Email, req.Nickname)
	if err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}

	utils.Success(ctx, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
	})
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, err.Error(), nil)
		return
	}

	accessToken, refreshToken, user, err := c.authService.Login(req.Username, req.Password)
	if err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}

	utils.Success(ctx, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"email":    user.Email,
			"role": gin.H{
				"id":   user.Role.ID,
				"name": user.Role.Name,
				"code": user.Role.Code,
			},
		},
	})
}

// RefreshToken 刷新令牌
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, err.Error(), nil)
		return
	}

	accessToken, err := c.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}

	utils.Success(ctx, gin.H{
		"access_token": accessToken,
	})
}

// GetUserInfo 获取当前用户信息
func (c *AuthController) GetUserInfo(ctx *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.FailWithMessage(ctx, utils.UNAUTHORIZED, "未找到用户信息", nil)
		return
	}

	// 验证令牌并获取用户信息
	token := ctx.GetHeader("Authorization")
	if token == "" {
		utils.FailWithMessage(ctx, utils.UNAUTHORIZED, "未提供认证令牌", nil)
		return
	}

	// 去掉Bearer前缀
	token = token[7:]
	user, err := c.authService.VerifyToken(token)
	if err != nil {
		utils.FailWithMessage(ctx, utils.UNAUTHORIZED, err.Error(), nil)
		return
	}

	// 检查令牌中的用户ID是否与上下文中的一致
	if user.ID != userID.(uint64) {
		utils.FailWithMessage(ctx, utils.UNAUTHORIZED, "用户身份验证失败", nil)
		return
	}

	utils.Success(ctx, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"email":    user.Email,
			"role": gin.H{
				"id":   user.Role.ID,
				"name": user.Role.Name,
				"code": user.Role.Code,
			},
		},
	})
}

// ChangePassword 修改密码
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	var req ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, err.Error(), nil)
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.FailWithMessage(ctx, utils.UNAUTHORIZED, "未找到用户信息", nil)
		return
	}

	err := c.authService.ChangePassword(userID.(uint64), req.OldPassword, req.NewPassword)
	if err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}

	utils.SuccessWithMessage(ctx, "密码修改成功", nil)
}
