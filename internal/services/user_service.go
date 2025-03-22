// Package services internal/services/user_service.go
package services

import (
	"errors"
	"star-go/internal/models"
	"star-go/internal/repository"
)

// IUserService 用户服务接口
type IUserService interface {
	GetUserByID(id uint64) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	ListUsers(page, pageSize int, search string) ([]*models.User, int64, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	DeleteUser(id uint64) error
	HasPermission(userID uint64, permission string) (bool, error)
}

// UserService 用户服务实现
type UserService struct {
	userRepo repository.IUserRepository
}

// NewUserService 创建用户服务实例
func NewUserService() IUserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

// 根据ID获取用户
func (s *UserService) GetUserByID(id uint64) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

// 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepo.FindByUsername(username)
}

// 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindByEmail(email)
}

// 获取用户列表
func (s *UserService) ListUsers(page, pageSize int, search string) ([]*models.User, int64, error) {
	return s.userRepo.List(page, pageSize, search)
}

// 创建用户
func (s *UserService) CreateUser(user *models.User) error {
	// 检查用户名是否已存在
	existingUser, _ := s.userRepo.FindByUsername(user.Username)
	if existingUser != nil {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, _ = s.userRepo.FindByEmail(user.Email)
	if existingUser != nil {
		return errors.New("邮箱已存在")
	}

	// 创建用户
	return s.userRepo.Create(user)
}

// 更新用户
func (s *UserService) UpdateUser(user *models.User) error {
	// 检查用户是否存在
	existingUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("用户不存在")
	}

	// 如果更新了用户名，检查是否与其他用户冲突
	if user.Username != existingUser.Username {
		conflictUser, _ := s.userRepo.FindByUsername(user.Username)
		if conflictUser != nil && conflictUser.ID != user.ID {
			return errors.New("用户名已存在")
		}
	}

	// 如果更新了邮箱，检查是否与其他用户冲突
	if user.Email != existingUser.Email {
		conflictUser, _ := s.userRepo.FindByEmail(user.Email)
		if conflictUser != nil && conflictUser.ID != user.ID {
			return errors.New("邮箱已存在")
		}
	}

	// 更新用户
	return s.userRepo.Update(user)
}

// 删除用户
func (s *UserService) DeleteUser(id uint64) error {
	// 检查用户是否存在
	existingUser, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("用户不存在")
	}

	// 删除用户
	return s.userRepo.Delete(id)
}

// 检查用户是否拥有指定权限
func (s *UserService) HasPermission(userID uint64, permission string) (bool, error) {
	// 获取用户
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.New("用户不存在")
	}

	// 检查用户是否拥有指定权限
	return user.HasPermission(permission), nil
}
