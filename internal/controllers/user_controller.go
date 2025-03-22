// internal/controllers/user_controller.go
package controllers

import (
	"strconv"
	"star-go/internal/models"
	"star-go/internal/services"
	"star-go/pkg/utils"

	"github.com/gin-gonic/gin"
)

// 用户控制器
type UserController struct {
	userService services.IUserService
}

// 创建用户控制器实例
func NewUserController() *UserController {
	return &UserController{
		userService: services.NewUserService(),
	}
}

// 用户创建请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"required,min=2,max=50"`
	RoleID   uint   `json:"role_id" binding:"required"`
}

// 用户更新请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	RoleID   uint   `json:"role_id" binding:"required"`
}

// 获取用户列表
func (c *UserController) GetUsers(ctx *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	search := ctx.DefaultQuery("search", "")

	// 获取用户列表
	users, total, err := c.userService.ListUsers(page, pageSize, search)
	if err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}

	// 构造响应数据
	var userList []gin.H
	for _, user := range users {
		userList = append(userList, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"nickname": user.Nickname,
			"role": gin.H{
				"id":   user.Role.ID,
				"name": user.Role.Name,
				"code": user.Role.Code,
			},
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	utils.Success(ctx, gin.H{
		"list":  userList,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// 根据ID获取用户
func (c *UserController) GetUser(ctx *gin.Context) {
	// 获取用户ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, "无效的用户ID", nil)
		return
	}

	// 获取用户
	user, err := c.userService.GetUserByID(id)
	if err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}
	if user == nil {
		utils.FailWithMessage(ctx, utils.NOT_FOUND, "用户不存在", nil)
		return
	}

	// 返回用户信息
	utils.Success(ctx, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"nickname": user.Nickname,
		"role": gin.H{
			"id":   user.Role.ID,
			"name": user.Role.Name,
			"code": user.Role.Code,
		},
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

// 创建用户
func (c *UserController) CreateUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, err.Error(), nil)
		return
	}

	// 创建用户对象
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Nickname: req.Nickname,
		RoleID:   req.RoleID,
	}

	// 设置密码
	if err := user.SetPassword(req.Password); err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, "密码设置失败: "+err.Error(), nil)
		return
	}

	// 创建用户
	if err := c.userService.CreateUser(user); err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}

	// 返回用户信息
	utils.Success(ctx, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

// 更新用户
func (c *UserController) UpdateUser(ctx *gin.Context) {
	// 获取用户ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, "无效的用户ID", nil)
		return
	}

	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, err.Error(), nil)
		return
	}

	// 获取用户
	user, err := c.userService.GetUserByID(id)
	if err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}
	if user == nil {
		utils.FailWithMessage(ctx, utils.NOT_FOUND, "用户不存在", nil)
		return
	}

	// 更新用户信息
	user.Nickname = req.Nickname
	user.Email = req.Email
	user.RoleID = req.RoleID

	// 保存更新
	if err := c.userService.UpdateUser(user); err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}

	// 返回成功信息
	utils.SuccessWithMessage(ctx, "用户更新成功", nil)
}

// 删除用户
func (c *UserController) DeleteUser(ctx *gin.Context) {
	// 获取用户ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, "无效的用户ID", nil)
		return
	}

	// 删除用户
	if err := c.userService.DeleteUser(id); err != nil {
		utils.FailWithMessage(ctx, utils.ERROR, err.Error(), nil)
		return
	}

	// 返回成功信息
	utils.SuccessWithMessage(ctx, "用户删除成功", nil)
}
