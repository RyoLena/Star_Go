// Package services internal/services/auth_service.go
package services

import (
	"errors"
	"star-go/internal/models"
	"star-go/internal/repository"
	"star-go/pkg/utils"

	"gorm.io/gorm"
)

// IAuthService 认证服务接口
type IAuthService interface {
	Register(username, password, email, nickname string) (*models.User, error)
	Login(username, password string) (string, string, *models.User, error)
	RefreshToken(refreshToken string) (string, error)
	VerifyToken(token string) (*models.User, error)
	ChangePassword(userID uint64, oldPassword, newPassword string) error
}

// AuthService 认证服务实现
type AuthService struct {
	userRepo repository.IUserRepository
}

// NewAuthService 创建认证服务实例
func NewAuthService() IAuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
	}
}

// Register 用户注册
func (s *AuthService) Register(username, password, email, nickname string) (*models.User, error) {
	// 检查用户名是否已存在
	existUser, _ := s.userRepo.FindByUsername(username)
	if existUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existUser, _ = s.userRepo.FindByEmail(email)
	if existUser != nil {
		return nil, errors.New("邮箱已存在")
	}

	// 创建用户
	user := &models.User{
		Username: username,
		Email:    email,
		Nickname: nickname,
		RoleID:   2, // 默认为普通用户角色
		Status:   models.StatusActive,
	}

	// 设置密码
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	// 保存用户
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (string, string, *models.User, error) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", nil, errors.New("用户不存在")
		}
		return "", "", nil, err
	}

	// 检查用户状态
	if !user.IsActive() {
		return "", "", nil, errors.New("用户已被禁用")
	}

	// 验证密码
	if !user.CheckPassword(password) {
		return "", "", nil, errors.New("密码错误")
	}

	// 更新最后登录时间
	user.UpdateLastLogin()
	if err := s.userRepo.Update(user); err != nil {
		return "", "", nil, err
	}

	// 生成访问令牌
	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	// 生成刷新令牌
	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(refreshToken string) (string, error) {
	// 验证刷新令牌
	claims, err := utils.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("无效的刷新令牌")
	}

	// 检查用户是否存在
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return "", errors.New("用户不存在")
	}

	// 检查用户状态
	if !user.IsActive() {
		return "", errors.New("用户已被禁用")
	}

	// 生成新的访问令牌
	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// 验证令牌
func (s *AuthService) VerifyToken(token string) (*models.User, error) {
	// 解析令牌
	claims, err := utils.ParseAccessToken(token)
	if err != nil {
		return nil, errors.New("无效的访问令牌")
	}

	// 查找用户
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("用户已被禁用")
	}

	return user, nil
}

// 修改密码
func (s *AuthService) ChangePassword(userID uint64, oldPassword, newPassword string) error {
	// 查找用户
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if !user.CheckPassword(oldPassword) {
		return errors.New("原密码错误")
	}

	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return err
	}

	// 更新用户
	return s.userRepo.Update(user)
}
