// Package database pkg/database/migration.go
package database

import (
	"star-go/internal/models"
	"star-go/pkg/logger"

	"go.uber.org/zap"
)

// RunMigrations 执行数据库迁移
func RunMigrations() error {
	logger.GetLogger().Info("开始执行数据库迁移...")

	// 自动迁移数据表结构
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Role{},
	); err != nil {
		logger.GetLogger().Error("数据库迁移失败", zap.Error(err))
		return err
	}

	logger.GetLogger().Info("数据库迁移完成")
	return nil
}

// InitAdminUser 初始化管理员账户
func InitAdminUser() error {
	// 初始化角色
	if err := initRoles(); err != nil {
		return err
	}

	// 初始化管理员账户
	if err := initAdmin(); err != nil {
		return err
	}

	return nil
}

// 初始化角色
func initRoles() error {
	logger.GetLogger().Info("检查并初始化角色...")

	// 检查角色表是否已有数据
	var count int64
	if err := DB.Model(&models.Role{}).Count(&count).Error; err != nil {
		logger.GetLogger().Error("查询角色表失败", zap.Error(err))
		return err
	}

	// 如果没有角色数据，则创建默认角色
	if count == 0 {
		logger.GetLogger().Info("创建默认角色...")

		// 创建超级管理员角色
		superuserRole := &models.Role{
			Name:        "超级管理员",
			Code:        models.RoleSuperuser,
			Description: "系统最高管理员，拥有所有权限",
			Permissions: models.Permissions{models.PermAll},
		}

		// 创建管理员角色
		adminRole := &models.Role{
			Name:        "管理员",
			Code:        models.RoleAdmin,
			Description: "系统管理员，拥有所有权限",
			Permissions: models.Permissions{models.PermAll},
		}

		// 创建普通用户角色
		userRole := &models.Role{
			Name:        "普通用户",
			Code:        models.RoleUser,
			Description: "普通用户，拥有基本权限",
			Permissions: models.Permissions{
				models.PermUserView,
				models.PermUserEdit,
				models.PermContentView,
				models.PermContentCreate,
				models.PermContentEdit,
			},
		}

		// 创建访客角色
		guestRole := &models.Role{
			Name:        "访客",
			Code:        models.RoleGuest,
			Description: "访客，仅拥有查看权限",
			Permissions: models.Permissions{
				models.PermUserView,
				models.PermContentView,
			},
		}

		// 批量创建角色
		roles := []*models.Role{superuserRole, adminRole, userRole, guestRole}
		if err := DB.Create(&roles).Error; err != nil {
			logger.GetLogger().Error("创建默认角色失败", zap.Error(err))
			return err
		}

		logger.GetLogger().Info("默认角色创建成功")
	}

	return nil
}

// 初始化管理员账户
func initAdmin() error {
	logger.GetLogger().Info("检查并初始化管理员账户...")

	// 检查是否已存在管理员账户
	var count int64
	if err := DB.Model(&models.User{}).Where("role_id = ?", 1).Count(&count).Error; err != nil {
		logger.GetLogger().Error("查询管理员账户失败", zap.Error(err))
		return err
	}

	// 如果没有管理员账户，则创建默认管理员
	if count == 0 {
		logger.GetLogger().Info("创建默认管理员账户...")

		// 创建管理员用户
		admin := &models.User{
			Username: "admin",
			Email:    "admin@example.com",
			Nickname: "系统管理员",
			RoleID:   1, // 管理员角色ID为1
			Status:   models.StatusActive,
		}

		// 设置默认密码
		if err := admin.SetPassword("123456"); err != nil {
			logger.GetLogger().Error("设置管理员密码失败", zap.Error(err))
			return err
		}

		// 创建管理员账户
		if err := DB.Create(admin).Error; err != nil {
			logger.GetLogger().Error("创建管理员账户失败", zap.Error(err))
			return err
		}

		logger.GetLogger().Info("默认管理员账户创建成功")
	}

	return nil
}
