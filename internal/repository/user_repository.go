// Package repository internal/repository/user_repository.go
package repository

import (
	"star-go/internal/models"
	"star-go/pkg/database"

	"gorm.io/gorm"
)

// IUserRepository 用户仓库接口
type IUserRepository interface {
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id uint64) error
	FindByID(id uint64) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	List(page, size int, query string) ([]*models.User, int64, error)
}

// 用户仓库实现
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository() IUserRepository {
	return &UserRepository{
		db: database.GetDB(),
	}
}

// 创建用户
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// 更新用户
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// 删除用户
func (r *UserRepository) Delete(id uint64) error {
	return r.db.Delete(&models.User{}, id).Error
}

// 根据ID查找用户
func (r *UserRepository) FindByID(id uint64) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据用户名查找用户
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据邮箱查找用户
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 查询用户列表
func (r *UserRepository) List(page, size int, query string) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	db := r.db.Model(&models.User{}).Preload("Role")

	// 如果有查询条件，添加查询
	if query != "" {
		db = db.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	// 计算总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err = db.Offset(offset).Limit(size).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
